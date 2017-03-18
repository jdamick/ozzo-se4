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
	"bytes"
	"os/exec"
	"strings"
)

func Uname(buf *UnameInfo) (err error) {
	buf.Sysname, err = sysctl("kern.ostype")
	if err != nil {
		return
	}

	buf.Nodename, err = sysctl("kern.hostname")
	if err != nil {
		return
	}

	buf.Release, err = sysctl("kern.osrelease")
	if err != nil {
		return
	}

	buf.Version, err = sysctl("kern.version")
	if err != nil {
		return
	}

	buf.Machine, err = sysctl("hw.machine")
	return
}

func sysctl(kernVarName string) (string, error) {
	// not the fastest, but it works
	cmd := exec.Command("sysctl", "-n", kernVarName)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}
