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
	"strings"
	"syscall"
)

func Uname(buf *UnameInfo) (err error) {
	u := &syscall.Utsname{}
	err = syscall.Uname(u)
	if err != nil {
		return
	}
	buf.Sysname = byte2str(u.Sysname)
	buf.Nodename = byte2str(u.Nodename)
	buf.Release = byte2str(u.Release)
	buf.Version = byte2str(u.Version)
	buf.Machine = byte2str(u.Machine)
	return nil
}

func byte2str(in [65]int8) string {
	out := make([]byte, len(in))
	length := 0
	for i, v := range in {
		if v == 0 {
			break
		}
		out[i] = byte(v)
		length++
	}
	return strings.TrimSpace(string(out[0:length]))
}
