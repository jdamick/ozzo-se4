[![Build Status](https://travis-ci.org/jdamick/ozzo-routing.svg?branch=master)](https://travis-ci.org/jdamick/ozzo-routing)

# ozzo-se4

Simple Standard Service Endpoints for Ozzo

Implementation of [Simple Spec for Service Status and Health](https://github.com/beamly/SE4) targeted for the [ozzo-routing](https://github.com/go-ozzo/ozzo-routing) framework.


## Usage

See the [example](example/example.go)


Run example:

```
  go run --ldflags '-X main.Version=1.1' example/example.go  
```

Then you can curl the endpoints, for example:

``` 
curl -s http://localhost:8080/service/status|jq ''
{
  "artifact_id": "",
  "build_number": "",
  "build_machine": "",
  "built_by": "me",
  "built_when": "",
  "git_sha1": "",
  "git_branch": "",
  "git_repo": "",
  "runbook_uri": "",
  "version": "1.1",
  "compiler_version": "gc",
  "current_time": "2017-03-18T22:15:58Z",
  "group_id": "",
  "machine_name": "6c4008a5dd24",
  "os_arch": "x86_64",
  "os_avgload": "1.58",
  "os_name": "Darwin",
  "os_numprocessors": "8",
  "os_version": "16.4.0",
  "up_duration": "1m7.675266382s",
  "up_since": "2017-03-18T22:14:50Z",
  "vm_name": "",
  "vm_vendor": "",
  "vm_version": "",
  "go_maxprocs": "8",
  "go_numroutines": "5"
}
```
