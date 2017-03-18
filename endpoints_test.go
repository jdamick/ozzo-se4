//
// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.
//
package se4

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-ozzo/ozzo-routing"
	. "github.com/smartystreets/goconvey/convey"
)

func contentType(resp *httptest.ResponseRecorder) string {
	v := resp.HeaderMap["Content-Type"]
	if len(v) > 0 {
		return v[0]
	}
	return ""
}

func TestBasicEndpoints(t *testing.T) {
	res := httptest.NewRecorder()
	var req *http.Request

	r := routing.New()

	se := NewStandardEndpoints()
	se.RegisterDefaultEndpoints(r)

	Convey("Get Status", t, func() {
		req, _ = http.NewRequest("GET", "/service/status", nil)
		r.ServeHTTP(res, req)
		So(contentType(res), ShouldStartWith, "application/json")
		So(res.Body.String(), ShouldContainSubstring, "artifact_id")
	})

	Convey("Get Healthcheck ASG", t, func() {
		res = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/service/healthcheck/asg", nil)
		r.ServeHTTP(res, req)
		So(res.Code, ShouldEqual, http.StatusOK)
		So(contentType(res), ShouldStartWith, "text/plain")
		So(res.Body.String(), ShouldEqual, "OK")
	})

	Convey("Get Healthcheck", t, func() {
		res = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/service/healthcheck", nil)
		r.ServeHTTP(res, req)
		So(contentType(res), ShouldStartWith, "application/json")
	})

	Convey("Get Healthcheck GTG", t, func() {
		res = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/service/healthcheck/gtg", nil)
		r.ServeHTTP(res, req)
		So(res.Code, ShouldEqual, http.StatusOK)
		So(contentType(res), ShouldStartWith, "text/plain")
		So(res.Body.String(), ShouldEqual, "OK")
	})

	Convey("Get Config", t, func() {
		res = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/service/config", nil)
		r.ServeHTTP(res, req)
		So(res.Code, ShouldEqual, http.StatusOK)
		So(contentType(res), ShouldStartWith, "application/json")
	})
}

func TestHealthcheckEndpoint(t *testing.T) {
	res := httptest.NewRecorder()
	var req *http.Request

	r := routing.New()

	se := NewStandardEndpoints()
	se.RegisterDefaultEndpoints(r)
	callCountHC1 := 0
	se.SetHealthCheckFuncs(time.Duration(1), func() HealthCheckResult {
		callCountHC1++
		result := HealthCheckResult{DurationMillis: DurationToMillis(time.Duration(20403) * time.Microsecond),
			Name: "some test", Result: HealthResultPassed, Timestamp: time.Now()}
		return result
	})

	time.Sleep(time.Duration(2) * time.Millisecond)

	Convey("Get Healthcheck", t, func() {
		res = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/service/healthcheck", nil)
		r.ServeHTTP(res, req)
		So(contentType(res), ShouldStartWith, "application/json")
		So(res.Body.String(), ShouldContainSubstring, "report_as_of")
		So(res.Body.String(), ShouldContainSubstring, "\"report_duration\":\"")
		So(res.Body.String(), ShouldContainSubstring, "\"duration_millis\":20.403")
		So(res.Body.String(), ShouldContainSubstring, "duration_millis")
		So(res.Body.String(), ShouldContainSubstring, `"test_name":"some test"`)
		So(callCountHC1, ShouldBeGreaterThan, 1)
	})

	callCountHC2 := 0
	// change healthcheck
	se.SetHealthCheckFuncs(time.Duration(1), func() HealthCheckResult {
		callCountHC2++
		result := HealthCheckResult{}
		time.Sleep(time.Duration(1) * time.Millisecond)
		return result
	})

	time.Sleep(time.Duration(1) * time.Millisecond)

	callCountHC1 = 0 // reset

	time.Sleep(time.Duration(2) * time.Millisecond)

	Convey("Get Empty Healthcheck", t, func() {
		So(callCountHC1, ShouldEqual, 0)
		So(callCountHC2, ShouldBeGreaterThan, 1)

		res = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/service/healthcheck", nil)
		r.ServeHTTP(res, req)
		So(contentType(res), ShouldStartWith, "application/json")
		So(res.Body.String(), ShouldContainSubstring, `report_as_of`)
		So(res.Body.String(), ShouldContainSubstring, `"test_name":""`)
		So(res.Body.String(), ShouldContainSubstring, "duration_millis")
	})
	se.SetHealthCheckFuncs(0)
}

func TestGtgAndCanaryEndpoint(t *testing.T) {
	res := httptest.NewRecorder()
	var req *http.Request

	r := routing.New()

	se := NewStandardEndpoints()
	se.RegisterDefaultEndpoints(r)

	se.SetGoodToGoFunc(func() bool {
		return false
	})

	Convey("Get Healthcheck ASG", t, func() {
		res = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/service/healthcheck/asg", nil)
		r.ServeHTTP(res, req)
		So(res.Code, ShouldEqual, http.StatusOK)
		So(contentType(res), ShouldStartWith, "text/plain")
		So(res.Body.String(), ShouldEqual, "OK")
	})

	Convey("Get Healthcheck GTG", t, func() {
		res = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/service/healthcheck/gtg", nil)
		r.ServeHTTP(res, req)
		So(res.Code, ShouldEqual, http.StatusServiceUnavailable)
		So(contentType(res), ShouldStartWith, "text/plain")
		So(res.Body.String(), ShouldNotEqual, "OK")
	})

	se.SetGoodToGoFunc(func() bool {
		return true
	})

	se.SetServiceCanaryFunc(func() bool {
		return false
	})

	Convey("Get Healthcheck ASG", t, func() {
		res = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/service/healthcheck/asg", nil)
		r.ServeHTTP(res, req)
		So(res.Code, ShouldEqual, http.StatusServiceUnavailable)
		So(contentType(res), ShouldStartWith, "text/plain")
		So(res.Body.String(), ShouldNotEqual, "OK")
	})

	Convey("Get Healthcheck GTG", t, func() {
		res = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/service/healthcheck/gtg", nil)
		r.ServeHTTP(res, req)
		So(res.Code, ShouldEqual, http.StatusOK)
		So(contentType(res), ShouldStartWith, "text/plain")
		So(res.Body.String(), ShouldEqual, "OK")
	})
}

func TestConfigEndpoint(t *testing.T) {
	res := httptest.NewRecorder()
	var req *http.Request

	r := routing.New()

	se := NewStandardEndpoints()
	se.RegisterDefaultEndpoints(r)

	se.SetConfigSourceFunc(ToConfigSourceFunc(map[string]string{"something": "a value"}))

	Convey("Get Config", t, func() {
		res = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/service/config", nil)
		r.ServeHTTP(res, req)
		So(res.Code, ShouldEqual, http.StatusOK)
		So(contentType(res), ShouldStartWith, "application/json")
		So(res.Body.String(), ShouldEqual, "{\"something\":\"a value\"}\n")
	})
}
