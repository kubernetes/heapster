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

package elasticsearch

import (
	"net/url"
	"reflect"
	"testing"
	"time"

	"fmt"
	elastic2 "gopkg.in/olivere/elastic.v3"
	elastic5 "gopkg.in/olivere/elastic.v5"
	"github.com/davecgh/go-spew/spew"
)

func TestCreateElasticSearchServiceV2(t *testing.T) {
	clusterName := "sandbox"
	esURI := fmt.Sprintf("?nodes=https://foo.com:20468&nodes=https://bar.com:20468&"+
		"esUserName=test&esUserSecret=password&maxRetries=10&startupHealthcheckTimeout=30&"+
		"sniff=false&healthCheck=false&ver=2&cluster_name=%s", clusterName)

	url, err := url.Parse(esURI)
	if err != nil {
		t.Fatalf("Error when parsing URL: %s", err.Error())
	}

	esSvc, err := CreateElasticSearchService(url)
	if err != nil {
		t.Fatalf("Error when creating config: %s", err.Error())
	}

	expectedClient, err := elastic2.NewClient(
		elastic2.SetURL("https://foo.com:20468", "https://bar.com:20468"),
		elastic2.SetBasicAuth("test", "password"),
		elastic2.SetMaxRetries(10),
		elastic2.SetHealthcheckTimeoutStartup(30*time.Second),
		elastic2.SetSniff(false), elastic2.SetHealthcheck(false))

	if err != nil {
		t.Fatalf("Error when creating client: %s", err.Error())
	}

	actualClientRefl := reflect.ValueOf(esSvc.EsClient).Elem().FieldByName("clientV2").Elem()
	expectedClientRefl := reflect.ValueOf(expectedClient).Elem()

	if actualClientRefl.FieldByName("basicAuthUsername").String() != expectedClientRefl.FieldByName("basicAuthUsername").String() {
		t.Fatal("basicAuthUsername is not equal")
	}
	if actualClientRefl.FieldByName("basicAuthUsername").String() != expectedClientRefl.FieldByName("basicAuthUsername").String() {
		t.Fatal("basicAuthUsername is not equal")
	}
	if actualClientRefl.FieldByName("maxRetries").Int() != expectedClientRefl.FieldByName("maxRetries").Int() {
		t.Fatal("maxRetries is not equal")
	}
	if actualClientRefl.FieldByName("healthcheckTimeoutStartup").Int() != expectedClientRefl.FieldByName("healthcheckTimeoutStartup").Int() {
		t.Fatal("healthcheckTimeoutStartup is not equal")
	}
	if actualClientRefl.FieldByName("snifferEnabled").Bool() != expectedClientRefl.FieldByName("snifferEnabled").Bool() {
		t.Fatal("snifferEnabled is not equal")
	}
	if actualClientRefl.FieldByName("healthcheckEnabled").Bool() != expectedClientRefl.FieldByName("healthcheckEnabled").Bool() {
		t.Fatal("healthcheckEnabled is not equal")
	}
	if esSvc.ClusterName != clusterName {
		t.Fatal("cluster name is not equal")
	}
}
func TestCreateElasticSearchServiceV5(t *testing.T) {
	clusterName := "sandbox"
	esURI := fmt.Sprintf("?nodes=https://foo.com:20468&nodes=https://bar.com:20468&"+
		"esUserName=test&esUserSecret=password&maxRetries=10&startupHealthcheckTimeout=30&"+
		"sniff=false&healthCheck=false&ver=5&cluster_name=%s", clusterName)

	url, err := url.Parse(esURI)
	if err != nil {
		t.Fatalf("Error when parsing URL: %s", err.Error())
	}

	esSvc, err := CreateElasticSearchService(url)
	if err != nil {
		t.Fatalf("Error when creating config: %s", err.Error())
	}

	expectedClient, err := elastic5.NewClient(
		elastic5.SetURL("https://foo.com:20468", "https://bar.com:20468"),
		elastic5.SetBasicAuth("test", "password"),
		elastic5.SetMaxRetries(10),
		elastic5.SetHealthcheckTimeoutStartup(30*time.Second),
		elastic5.SetSniff(false), elastic5.SetHealthcheck(false))

	if err != nil {
		t.Fatalf("Error when creating client: %s", err.Error())
	}

	actualClientRefl := reflect.ValueOf(esSvc.EsClient).Elem().FieldByName("clientV5").Elem()
	expectedClientRefl := reflect.ValueOf(expectedClient).Elem()

	if actualClientRefl.FieldByName("basicAuthUsername").String() != expectedClientRefl.FieldByName("basicAuthUsername").String() {
		t.Fatal("basicAuthUsername is not equal")
	}
	if actualClientRefl.FieldByName("basicAuthUsername").String() != expectedClientRefl.FieldByName("basicAuthUsername").String() {
		t.Fatal("basicAuthUsername is not equal")
	}
	if actualClientRefl.FieldByName("healthcheckTimeoutStartup").Int() != expectedClientRefl.FieldByName("healthcheckTimeoutStartup").Int() {
		t.Fatal("healthcheckTimeoutStartup is not equal")
	}
	if actualClientRefl.FieldByName("snifferEnabled").Bool() != expectedClientRefl.FieldByName("snifferEnabled").Bool() {
		t.Fatal("snifferEnabled is not equal")
	}
	if actualClientRefl.FieldByName("healthcheckEnabled").Bool() != expectedClientRefl.FieldByName("healthcheckEnabled").Bool() {
		t.Fatal("healthcheckEnabled is not equal")
	}
	if esSvc.ClusterName != clusterName {
		t.Fatal("cluster name is not equal")
	}
}

func TestCreateElasticSearchServiceForDefaultClusterName(t *testing.T) {
	esURI := "?nodes=https://foo.com:20468&nodes=https://bar.com:20468&" +
		"esUserName=test&esUserSecret=password&maxRetries=10&startupHealthcheckTimeout=30&" +
		"sniff=false&healthCheck=false"

	url, err := url.Parse(esURI)
	if err != nil {
		t.Fatalf("Error when parsing URL: %s", err.Error())
	}

	esSvc, err := CreateElasticSearchService(url)
	if err != nil {
		t.Fatalf("Error when creating config: %s", err.Error())
	}

	if esSvc.ClusterName != ESClusterName {
		t.Fatalf("cluster name is not equal. Expected: %s, Got: %s", ESClusterName, esSvc.ClusterName)
	}
}

func TestCreateElasticSearchServiceSingleDnsEntrypoint(t *testing.T) {
	clusterName := "sandbox"
	esURI := fmt.Sprintf("https://foo.com:9200?"+
		"esUserName=test&esUserSecret=password&maxRetries=10&startupHealthcheckTimeout=30&"+
		"sniff=false&healthCheck=false&cluster_name=%s", clusterName)

	url, err := url.Parse(esURI)
	if err != nil {
		t.Fatalf("Error when parsing URL: %s", err.Error())
	}

	esSvc, err := CreateElasticSearchService(url)
	if err != nil {
		t.Fatalf("Error when creating config: %s", err.Error())
	}

	expectedClient, err := elastic5.NewClient(
		elastic5.SetURL("https://foo.com:9200"),
		elastic5.SetBasicAuth("test", "password"),
		elastic5.SetMaxRetries(10),
		elastic5.SetHealthcheckTimeoutStartup(30*time.Second),
		elastic5.SetSniff(false), elastic5.SetHealthcheck(false))

	if err != nil {
		t.Fatalf("Error when creating client: %s", err.Error())
	}

	actualClientRefl := reflect.ValueOf(esSvc.EsClient).Elem().FieldByName("clientV5").Elem()
	expectedClientRefl := reflect.ValueOf(expectedClient).Elem()

	if actualClientRefl.FieldByName("basicAuthUsername").String() != expectedClientRefl.FieldByName("basicAuthUsername").String() {
		spew.Dump(actualClientRefl.FieldByName("basicAuthUsername"))
		t.Fatal("basicAuthUsername is not equal")
	}
	if actualClientRefl.FieldByName("healthcheckTimeoutStartup").Int() != expectedClientRefl.FieldByName("healthcheckTimeoutStartup").Int() {
		t.Fatal("healthcheckTimeoutStartup is not equal")
	}
	if actualClientRefl.FieldByName("snifferEnabled").Bool() != expectedClientRefl.FieldByName("snifferEnabled").Bool() {
		t.Fatal("snifferEnabled is not equal")
	}
	if actualClientRefl.FieldByName("healthcheckEnabled").Bool() != expectedClientRefl.FieldByName("healthcheckEnabled").Bool() {
		t.Fatal("healthcheckEnabled is not equal")
	}
	if esSvc.ClusterName != clusterName {
		t.Fatal("cluster name is not equal")
	}
}
