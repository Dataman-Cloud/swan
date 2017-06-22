
<img src="docs/assets/img/swan.png" width="350">

##

[![Build Status](https://travis-ci.org/Dataman-Cloud/swan.svg?branch=master)](https://travis-ci.org/Dataman-Cloud/swan)
[![codecov](https://codecov.io/gh/Dataman-Cloud/swan/branch/master/graph/badge.svg)](https://codecov.io/gh/Dataman-Cloud/swan)
[![Go Report Card](https://goreportcard.com/badge/github.com/Dataman-Cloud/swan)](https://goreportcard.com/report/github.com/Dataman-Cloud/swan)
[![Join the chat at https://gitter.im/Dataman-Cloud/swan](https://badges.gitter.im/Dataman-Cloud/swan.svg)](https://gitter.im/Dataman-Cloud/swan?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

+ *Website:* https://dataman-cloud.github.io/swan-site  

## What is `Swan`?

`Swan` is a `Mesos` framework forcus on long running tasks, which inspired by `Marathon`, written with `Golang` and base on latest `Mesos` HTTP API.

## What does `Swan` do?

With `Swan` you can deploy long running `application` on mesos cluster, control lifecycle of the application, scale up or scale down any `instances` size you want, can also do rolling-update with new versions base on flexible `health checks` defined by you. `Swan` also designed for high avaliability which means any server crash wouldn't affect running applications. with `service discovery` and `API gateway` build-in, service discovery made easy.

#### `Swan` is created and maintained by [Dataman-Cloud](https://github.com/Dataman-Cloud), and licensed under the Apache License, Version 2.0.

#### for more documentation about `Swan` please refer to [swan-docs](https://dataman-cloud.github.io/swan-site/docs/)


## Features

+ Application Management 
+ ScaleUp/ScaleDown
+ Rolling update 
+ Rollback
+ Mesos-based health checks 
+ HA 
+ Event Subscription
+ Compose
+ Calico-based IP-Per-Task
+ Schedule stategy

## Installation

### From Source

clone `Swan` source code from github.com:

```
  git clone git@github.com:Dataman-Cloud/swan.git
```


Then you can compile `Swan` with:

```
  make
  
  # or build within docker
  
  make docker-build
```


## Run 
```
./bin/swan manager --mesos=zk://192.168.1.47:2181/mesos --zk=zk://192.168.1.47:2181/swan061904 --log-level=info --listen=0.0.0.0:9999
```
```
./bin/swan --help
```

### API
| Method        | API                        | 
| ------------- |:---------------------------|
| GET           | /v1/apps                   |
| POST          | /v1/apps                   |
| GET           | /v1/apps/{app_id}          |
| DELETE        | /v1/apps/{app_id}          |
| PATCH         | /v1/apps/{app_id}/scale    |
| PUT           | /v1/apps/{app_id}/update   |
| POST          | /v1/apps/{app_id}/rollback |
| GET           | /v1/apps/{app_id}/tasks    |
| GET           | /v1/apps/{app_id}/versions |
| GET           | /v1/apps/{app_id}/tasks/{task_id} |
| GET           | /v1/apps/{app_id}/versions/{version_id} |

| Method        | API
|---------------|:-------------------------|
| POST          | /v1/compose | 
| GET           | /v1/compose | 
| GET           | /v1/compose/{compose_id} |
| DELETE        | /v1/compose/{compose_id} | 

| Method        | API   |
|---------------|:-------------------------|
| GET           | /ping |
| GET           | /v1/events |
| GET           | /version |

[API](https://github.com/Dataman-Cloud/swan/tree/master/docs/api.md)

## Contributing
If you want to contribute to swan, make a PR or report a issue.
The goal of swan is to become the default and best scheduler for mesos, so let's do it!

## Licensing
Swan is licensed under the Apache License, Version 2.0. See
[LICENSE](https://github.com/Dataman-Cloud/swan/blob/master/LICENSE) for the full
license text.
