// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package elasticsearch5

import (
	"net/url"
	"sync"
	"time"

	"github.com/golang/glog"
	es5Common "k8s.io/heapster/common/elasticsearch5"
	event_core "k8s.io/heapster/events/core"
	"k8s.io/heapster/metrics/core"
	kube_api "k8s.io/kubernetes/pkg/api"
)

const (
	typeName = "events"
)

// SaveDataFunc is a pluggable function to enforce limits on the object
type SaveDataFunc func(date time.Time, sinkData []interface{}) error

type elasticSearchSink struct {
	esSvc     es5Common.ElasticSearchService
	saveData  SaveDataFunc
	flushData func() error
	sync.RWMutex
}

type EsSinkPoint struct {
	Count                    interface{}
	Metadata                 interface{}
	InvolvedObject           interface{}
	Source                   interface{}
	FirstOccurrenceTimestamp time.Time
	LastOccurrenceTimestamp  time.Time
	Message                  string
	Reason                   string
	Type                     string
	EventTags                map[string]string
}

func eventToPoint(event *kube_api.Event, clusterName string) (*EsSinkPoint, error) {
	point := EsSinkPoint{
		FirstOccurrenceTimestamp: event.FirstTimestamp.Time.UTC(),
		LastOccurrenceTimestamp:  event.LastTimestamp.Time.UTC(),
		Message:                  event.Message,
		Reason:                   event.Reason,
		Type:                     event.Type,
		Count:                    event.Count,
		Metadata:                 event.ObjectMeta,
		InvolvedObject:           event.InvolvedObject,
		Source:                   event.Source,
		EventTags: map[string]string{
			"eventID":      string(event.UID),
			"cluster_name": clusterName,
		},
	}
	if event.InvolvedObject.Kind == "Pod" {
		point.EventTags[core.LabelPodId.Key] = string(event.InvolvedObject.UID)
		point.EventTags[core.LabelPodName.Key] = event.InvolvedObject.Name
	}
	point.EventTags[core.LabelHostname.Key] = event.Source.Host
	return &point, nil
}

func (sink *elasticSearchSink) ExportEvents(eventBatch *event_core.EventBatch) {
	sink.Lock()
	defer sink.Unlock()
	for _, event := range eventBatch.Events {
		point, err := eventToPoint(event, sink.esSvc.ClusterName)
		if err != nil {
			glog.Warningf("Failed to convert event to point: %v", err)
		}
		err = sink.saveData(point.LastOccurrenceTimestamp, []interface{}{*point})
		if err != nil {
			glog.Warningf("Failed to export data to ElasticSearch5 sink: %v", err)
		}
	}
	err := sink.flushData()
	if err != nil {
		glog.Warningf("Failed to flushing data to ElasticSearch5 sink: %v", err)
	}
}

func (sink *elasticSearchSink) Name() string {
	return "ElasticSearch5 Sink"
}

func (sink *elasticSearchSink) Stop() {
	// nothing needs to be done.
}

func NewElasticSearchSink(uri *url.URL) (event_core.EventSink, error) {
	var esSink elasticSearchSink
	esSvc, err := es5Common.CreateElasticSearchService(uri)
	if err != nil {
		glog.Warning("Failed to config ElasticSearch5")
		return nil, err
	}

	esSink.esSvc = *esSvc
	esSink.saveData = func(date time.Time, sinkData []interface{}) error {
		return esSvc.SaveData(date, typeName, sinkData)
	}
	esSink.flushData = func() error {
		return esSvc.FlushData()
	}

	glog.V(2).Info("ElasticSearch5 sink setup successfully")
	return &esSink, nil
}
