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
package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-ozzo/ozzo-routing"
	se4 "github.com/jdamick/ozzo-se4"
)

var (
	Version = "0"
)

// for example, to run this: go run --ldflags '-X main.Version=1.1' example/example.go
func main() {
	router := routing.New()

	bi := &se4.BuildInfo{Version: Version, BuiltBy: "me"}

	se := se4.NewStandardEndpointsWithBuildInfo(bi)
	se.RegisterDefaultEndpoints(router)

	se.SetHealthCheckFuncs(time.Duration(10)*time.Second,
		func() se4.HealthCheckResult {
			ts := time.Now()
			//
			// run some check..
			//
			duration := time.Since(ts)

			result := se4.HealthCheckResult{
				DurationMillis: se4.DurationToMillis(duration),
				Name:           "healtcheck #1",
				Result:         se4.HealthResultPassed,
				Timestamp:      time.Now()}
			return result
		},
		// another health check..
		func() se4.HealthCheckResult {
			ts := time.Now()
			//
			// run some check..
			//
			duration := time.Since(ts)

			result := se4.HealthCheckResult{
				DurationMillis: se4.DurationToMillis(duration),
				Name:           "healtcheck #2",
				Result:         se4.HealthResultFailed,
				Timestamp:      time.Now()}
			return result
		},
	)

	se.SetGoodToGoFunc(func() bool {
		// run some good to go check..
		return true
	})

	se.SetServiceCanaryFunc(func() bool {
		// run some canary check..
		return true
	})

	port := "8080"
	se.SetConfigSourceFunc(se4.ToConfigSourceFunc(map[string]string{
		"port": port,
	}))

	http.Handle("/", router)
	fmt.Printf("listening on: %v\n", port)
	fmt.Printf("now you can run: \n")
	fmt.Printf("curl -v http://localhost:8080/service/status\n")
	fmt.Printf("curl -v http://localhost:8080/service/config\n")
	fmt.Printf("curl -v http://localhost:8080/service/healthcheck\n")
	fmt.Printf("curl -v http://localhost:8080/service/healthcheck/gtg\n")
	fmt.Printf("curl -v http://localhost:8080/service/healthcheck/asg\n")
	http.ListenAndServe(":"+port, nil)
}
