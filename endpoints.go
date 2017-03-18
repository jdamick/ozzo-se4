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
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"

	"time"

	sigar "github.com/cloudfoundry/gosigar"
	routing "github.com/go-ozzo/ozzo-routing"
	"github.com/go-ozzo/ozzo-routing/content"
)

const (
	MIME_TEXT_PLAIN = "text/plain"
)

//
// https://github.com/beamly/SE4/blob/master/SE4.md
//
// Name	         HTTP Verb	URI Path
// Status	         GET	/service/status
// Healthcheck	     GET	/service/healthcheck
// GTG (Good to Go)	 GET	/service/healthcheck/gtg
// Service Canary	 GET	/service/healthcheck/asg
//
// Optional
// Name	         HTTP Verb	URI Path
// Config	         GET	/service/config

type StandardEndpoints struct {
	Status       *Status
	healthReport HealthCheckReport
	locker       *sync.Mutex

	healthChecks     []HealthCheckFunc
	canaryCheck      ServiceCanaryFunc
	gtgCheck         GoodToGoFunc
	configSrc        ConfigSourceFunc
	healthCheckTimer *time.Timer
}

type ReportDuration time.Duration

type HealthCheckReport struct {
	Timestamp time.Time           `json:"report_as_of"`    // The time at which this report was generated (this may not be the current time)
	Duration  ReportDuration      `json:"report_duration"` // How long it took to generate the report
	Results   []HealthCheckResult `json:"tests"`           // array of test results
}

type HealthCheckResult struct {
	// To convert Duration use: DurationToMillis(..)
	DurationMillis float64   `json:"duration_millis"` // Number of milliseconds taken to run the test
	Name           string    `json:"test_name"`       // The name of the test, a name that is meaningful to supporting engineers
	Result         string    `json:"test_result"`     // The state of the test, may be "not_run", "running", "passed", "failed"
	Timestamp      time.Time `json:"tested_at"`       // The time at which this test was executed
}

func DurationToMillis(duration time.Duration) float64 {
	sec := duration / time.Millisecond
	nsec := duration % time.Millisecond
	return float64(sec) + float64(nsec)/1e6
}

// HealthResult
// "not_run", "running", "passed", "failed"
const (
	HealthResultNotRun  = "not_run"
	HealthResultRunning = "running"
	HealthResultPassed  = "passed"
	HealthResultFailed  = "failed"
)

// Healthcheck resource provides information about internal health and its perceived health of downstream dependencies.
type HealthCheckFunc func() HealthCheckResult

// The "Service Canary" (ASG) returns a successful response in the case that the service is in a healthy state.
type ServiceCanaryFunc func() bool

// The "Good To Go" (GTG) returns a successful response in the case that the service is in an operational state and is able to receive traffic.
type GoodToGoFunc func() bool

// ConfigSourceFunc returns the configuration that is used by the service
type ConfigSourceFunc func() interface{}

func ToConfigSourceFunc(config interface{}) ConfigSourceFunc {
	return func() interface{} { return config }
}

// these should be provided by the application
type BuildInfo struct {
	ArtifactID   string `json:"artifact_id"`
	BuildNumber  string `json:"build_number"`
	BuildMachine string `json:"build_machine"`
	BuiltBy      string `json:"built_by"`
	BuiltWhen    string `json:"built_when"`
	// VCS (baked in)
	GitSha1   string `json:"git_sha1"`
	GitBranch string `json:"git_branch"` // ADDITIONAL -
	GitRepo   string `json:"git_repo"`   // ADDITIONAL -
	// runbook (baked in)
	RunbookURI string `json:"runbook_uri"`
	// version number (baked in)
	Version string `json:"version"`
}

type Status struct {
	// Build info (baked in)
	BuildInfo

	CompilerVersion string `json:"compiler_version"` // optional

	// per request
	CurrentTime string `json:"current_time"` // dynamic  - per req
	// maven group
	GroupID string `json:"group_id"` // N/A - maven

	// machine (baked in)
	MachineName    string `json:"machine_name"`     // dynamic
	OSArch         string `json:"os_arch"`          // dynamic
	OSAvgload      string `json:"os_avgload"`       // dynamic - per req
	OSName         string `json:"os_name"`          // dynamic
	OSNumProcessor string `json:"os_numprocessors"` // dynamic
	OSVersion      string `json:"os_version"`       // dynamic

	// dynamic / at startup
	UpDuration string `json:"up_duration"` // dynamic - per req
	UpSince    string `json:"up_since"`    // at startup

	// N/A
	VMName    string `json:"vm_name"`    // N/A
	VMVendor  string `json:"vm_vendor"`  // N/A
	VMVersion string `json:"vm_version"` // N/A
	// go additions
	GoMaxProcs    string `json:"go_maxprocs"`    // ADDITIONAL -
	GoNumRoutines string `json:"go_numroutines"` // ADDITIONAL - dynamic

	UpSinceTime time.Time `json:"-"`
}

func NewStandardEndpointsWithBuildInfo(buildInfo *BuildInfo) *StandardEndpoints {
	now := time.Now().UTC()
	u := &UnameInfo{}
	Uname(u)
	name, _ := os.Hostname()
	s := &Status{
		BuildInfo: *buildInfo,
		UpSince:   now.Format(time.RFC3339), UpSinceTime: now, MachineName: name,
		OSArch: u.Machine, OSVersion: u.Release, OSName: u.Sysname,
		CompilerVersion: runtime.Compiler,
		OSNumProcessor:  strconv.Itoa(runtime.NumCPU()),
		GoMaxProcs:      strconv.Itoa(runtime.GOMAXPROCS(-1))}

	return &StandardEndpoints{Status: s, locker: &sync.Mutex{}}
}

func NewStandardEndpoints() *StandardEndpoints {
	buildInfo := &BuildInfo{
		ArtifactID:   "undefined",
		BuildNumber:  "undefined",
		BuildMachine: "undefined",
		BuiltBy:      "undefined",
		BuiltWhen:    "undefined",
		GitSha1:      "undefined",
		GitBranch:    "undefined",
		GitRepo:      "undefined",
		RunbookURI:   "undefined",
		Version:      "dev",
	}
	return NewStandardEndpointsWithBuildInfo(buildInfo)
}

var (
	DefaultHealthCheckInterval = time.Duration(10) * time.Second
)

// Set the health check functions to run at a specified interval
func (s *StandardEndpoints) SetHealthCheckFuncs(interval time.Duration, healthchecks ...HealthCheckFunc) {
	s.locker.Lock()
	defer s.locker.Unlock()
	s.healthChecks = healthchecks
	if s.healthCheckTimer != nil {
		s.healthCheckTimer.Stop()
	}

	// stop any existing timers first, so you could pass in an empty array to stop it..
	if len(healthchecks) == 0 {
		return
	}

	// fire right away, then adjust to interval
	s.healthCheckTimer = time.AfterFunc(1, func() {
		report := HealthCheckReport{}
		start := time.Now()

		var results []HealthCheckResult
		for _, chk := range healthchecks {
			results = append(results, chk())
		}
		report.Results = results
		report.Duration = ReportDuration(time.Since(start))
		report.Timestamp = time.Now().UTC()

		s.locker.Lock()
		defer s.locker.Unlock()
		s.healthReport = report
		s.healthCheckTimer.Reset(interval)
	})
}

func (s *StandardEndpoints) SetServiceCanaryFunc(canaryCheck ServiceCanaryFunc) {
	s.locker.Lock()
	defer s.locker.Unlock()
	s.canaryCheck = canaryCheck
}

func (s *StandardEndpoints) SetGoodToGoFunc(gtgCheck GoodToGoFunc) {
	s.locker.Lock()
	defer s.locker.Unlock()
	s.gtgCheck = gtgCheck
}

func (s *StandardEndpoints) SetConfigSourceFunc(configSrc ConfigSourceFunc) {
	s.locker.Lock()
	defer s.locker.Unlock()
	s.configSrc = configSrc
}

func (s *StandardEndpoints) generateStatus() *Status {
	s.locker.Lock()
	defer s.locker.Unlock()
	concreteSigar := sigar.ConcreteSigar{}
	avg, err := concreteSigar.GetLoadAverage()
	s.Status.OSAvgload = "0.0"
	if err == nil {
		s.Status.OSAvgload = fmt.Sprintf("%.2f", avg.One)
	}
	now := time.Now().UTC()
	s.Status.UpDuration = now.Sub(s.Status.UpSinceTime).String()
	s.Status.CurrentTime = now.Format(time.RFC3339)
	s.Status.GoNumRoutines = strconv.Itoa(runtime.NumGoroutine())

	return s.Status
}

func (s *StandardEndpoints) RegisterDefaultEndpoints(router *routing.Router) {
	s.RegisterEndpoints(CreateDefaultRouterGroup(router))
}

func CreateDefaultRouterGroup(router *routing.Router) *routing.RouteGroup {
	return router.Group("/service")
}

func (s *StandardEndpoints) RegisterEndpoints(group *routing.RouteGroup) {
	group.Use(
		content.TypeNegotiator(routing.MIME_JSON),
	)

	group.Get("/status", func(c *routing.Context) error {
		c.Write(s.generateStatus())
		return nil
	})

	group.Get("/config", func(c *routing.Context) error {
		if s.configSrc != nil {
			c.Write(s.configSrc())
		}
		return nil
	})

	group.Get("/healthcheck", func(c *routing.Context) error {
		s.locker.Lock()
		defer s.locker.Unlock()
		c.Response.WriteHeader(http.StatusOK)
		c.Write(s.healthReport)
		return nil
	})

	textDataWriter := &TextPlainDataWriter{}

	// successful response is a 200 OK with a content of the text "OK" (including quotes) and a media type of "plain/text"
	// failed response is a 5XX reponse with either a 500 or 503 response preferred.
	group.Get("/healthcheck/gtg", func(c *routing.Context) error {
		result := true
		if s.gtgCheck != nil {
			result = s.gtgCheck()
		}

		textDataWriter.SetHeader(c.Response)
		c.SetDataWriter(textDataWriter)
		if result {
			c.Response.WriteHeader(http.StatusOK)
			c.Write("OK") // if we're here, we're good (for now)
		} else {
			c.Response.WriteHeader(http.StatusServiceUnavailable)
		}
		return nil
	})

	// successful response is a 200 OK with a content of the text "OK" (including quotes) and a media type of "plain/text"
	// failed response is a 5XX reponse with either a 500 or 503 response preferred.
	group.Get("/healthcheck/asg", func(c *routing.Context) error {
		c.SetDataWriter(&content.HTMLDataWriter{})

		result := true
		if s.canaryCheck != nil {
			result = s.canaryCheck()
		}

		textDataWriter.SetHeader(c.Response)
		c.SetDataWriter(textDataWriter)
		if result {
			c.Response.WriteHeader(http.StatusOK)
			c.Write("OK") // if we're here, we're good (for now)
		} else {
			c.Response.WriteHeader(http.StatusServiceUnavailable)
		}
		return nil
	})
}

type TextPlainDataWriter struct{}

func (w *TextPlainDataWriter) SetHeader(res http.ResponseWriter) {
	res.Header().Set("Content-Type", MIME_TEXT_PLAIN+"; charset=UTF-8")
}

func (w *TextPlainDataWriter) Write(res http.ResponseWriter, data interface{}) error {
	return routing.DefaultDataWriter.Write(res, fmt.Sprintf("%v", data))
}

func (r ReportDuration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(r).String())
}
