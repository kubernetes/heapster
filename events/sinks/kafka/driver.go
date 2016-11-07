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

package kafka

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/optiopay/kafka"
	"github.com/optiopay/kafka/proto"
	event_core "k8s.io/heapster/events/core"
	"k8s.io/heapster/metrics/core"
	kube_api "k8s.io/kubernetes/pkg/api"
)

const (
	partition              = 0
	brokerClientID         = "kafka-sink"
	brokerDialTimeout      = 10 * time.Second
	brokerDialRetryLimit   = 1
	brokerDialRetryWait    = 0
	brokerLeaderRetryLimit = 1
	brokerLeaderRetryWait  = 0
	dataTopic              = "heapster-events"
)

type KafkaSinkPoint struct {
	EventValue     interface{}
	EventTimestamp time.Time
	EventTags      map[string]string
}

type kafkaSink struct {
	producer  kafka.Producer
	dataTopic string
	sync.RWMutex
}

func getEventValue(event *kube_api.Event) (string, error) {
	// TODO: check whether indenting is required.
	bytes, err := json.MarshalIndent(event, "", " ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func eventToPoint(event *kube_api.Event) (*KafkaSinkPoint, error) {
	value, err := getEventValue(event)
	if err != nil {
		return nil, err
	}
	point := KafkaSinkPoint{
		EventTimestamp: event.LastTimestamp.Time.UTC(),
		EventValue:     value,
		EventTags: map[string]string{
			"eventID": string(event.UID),
		},
	}
	if event.InvolvedObject.Kind == "Pod" {
		point.EventTags[core.LabelPodId.Key] = string(event.InvolvedObject.UID)
		point.EventTags[core.LabelPodName.Key] = event.InvolvedObject.Name
	}
	point.EventTags[core.LabelHostname.Key] = event.Source.Host
	return &point, nil
}

func (sink *kafkaSink) ExportEvents(eventBatch *event_core.EventBatch) {
	sink.Lock()
	defer sink.Unlock()

	for _, event := range eventBatch.Events {
		point, err := eventToPoint(event)
		if err != nil {
			glog.Warningf("Failed to convert event to point: %v", err)
		}
		sink.produceKafkaMessage(*point, sink.dataTopic)
	}
}

func (sink *kafkaSink) produceKafkaMessage(dataPoint KafkaSinkPoint, topic string) error {
	start := time.Now()
	jsonItems, err := json.Marshal(dataPoint)
	if err != nil {
		return fmt.Errorf("failed to transform the items to json : %s", err)
	}
	message := &proto.Message{Value: []byte(string(jsonItems))}
	_, err = sink.producer.Produce(topic, partition, message)
	if err != nil {
		return fmt.Errorf("failed to produce message to %s:%d: %s", topic, partition, err)
	}
	end := time.Now()
	glog.V(4).Info("Exported %d data to kafka in %s", len([]byte(string(jsonItems))), end.Sub(start))
	return nil
}

func (sink *kafkaSink) Name() string {
	return "Apache Kafka Sink"
}

func (sink *kafkaSink) Stop() {
	// nothing needs to be done.
}

// setupProducer returns a producer of kafka server
func setupProducer(sinkBrokerHosts []string, brokerConf kafka.BrokerConf) (kafka.Producer, error) {
	glog.V(3).Infof("attempting to setup kafka sink")
	broker, err := kafka.Dial(sinkBrokerHosts, brokerConf)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to kafka cluster: %s", err)
	}
	defer broker.Close()

	//create kafka producer
	conf := kafka.NewProducerConf()
	conf.RequiredAcks = proto.RequiredAcksLocal
	sinkProducer := broker.Producer(conf)
	glog.V(3).Infof("kafka sink setup successfully")
	return sinkProducer, nil
}

func NewKafkaSink(uri *url.URL) (event_core.EventSink, error) {
	opts, err := url.ParseQuery(uri.RawQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to parser url's query string: %s", err)
	}

	var topic string = dataTopic
	if len(opts["eventstopic"]) > 0 {
		topic = opts["eventstopic"][0]
	}

	var kafkaBrokers []string
	if len(opts["brokers"]) < 1 {
		return nil, fmt.Errorf("There is no broker assigned for connecting kafka")
	}
	kafkaBrokers = append(kafkaBrokers, opts["brokers"]...)
	glog.V(2).Infof("initializing kafka sink with brokers - %v", kafkaBrokers)

	//structure the config of broker
	brokerConf := kafka.NewBrokerConf(brokerClientID)
	brokerConf.DialTimeout = brokerDialTimeout
	brokerConf.DialRetryLimit = brokerDialRetryLimit
	brokerConf.DialRetryWait = brokerDialRetryWait
	brokerConf.LeaderRetryLimit = brokerLeaderRetryLimit
	brokerConf.LeaderRetryWait = brokerLeaderRetryWait
	brokerConf.AllowTopicCreation = true

	// set up producer of kafka server.
	sinkProducer, err := setupProducer(kafkaBrokers, brokerConf)
	if err != nil {
		return nil, fmt.Errorf("Failed to setup Producer: - %v", err)
	}

	return &kafkaSink{
		producer:  sinkProducer,
		dataTopic: topic,
	}, nil
}
