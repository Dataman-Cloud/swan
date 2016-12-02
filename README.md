
<img src="assets/img/swan.png" width="350">

##

[![Build Status](https://travis-ci.org/Dataman-Cloud/swan.svg?branch=master)](https://travis-ci.org/Dataman-Cloud/swan)
[![codecov](https://codecov.io/gh/Dataman-Cloud/swan/branch/master/graph/badge.svg)](https://codecov.io/gh/Dataman-Cloud/swan)
[![Go Report Card](https://goreportcard.com/badge/github.com/Dataman-Cloud/swan)](https://goreportcard.com/report/github.com/Dataman-Cloud/swan)
[![Join the chat at https://gitter.im/Dataman-Cloud/swan](https://badges.gitter.im/Dataman-Cloud/swan.svg)](https://gitter.im/Dataman-Cloud/swan?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

#### swan is a mesos scheduling framework written in golang based on mesos new HTTP API.

#### you can use swan to deployment application on mesos cluster, and manage the entire lifecycle of the application. you can do rolling-update with new version, you can scale application, and you can do health check for your applications and auto failover when applications or services are not available.

#### swan is maintained by [dataman-cloud](https://github.com/Dataman-Cloud), and  licensed under the Apache License, Version 2.0. 

## Features
+ Application deployment
+ Application scaling
+ Rolling update
+ Version rollback
+ Health check
+ Auto failover

## Special features
+ The instance name is fixed during the application lifecycle. 
+ The instance index is continuously incremented from zero.

## Installation 
### Use Docker Compose(all-in-one, include mesos)
```
docker-compose up -d
```
### From Source(swan only)
First get the swan:
```
go get github.com/Dataman-Cloud/swan
```
Then you can compile `swan` with:
```
go install github.com/Dataman-Cloud/swan
```
`swan` will be installed at $GOPATH/bin/swan, If `$GOPATH/bin` is in your `PATH`, you can invoke `swan` from the CLI.

### Latest Release(swan only)
To get started with the latest release, run the following commands:
```
wget https://github.com/Dataman-Cloud/swan/releases/download/v0.1/swan
chmod +x swan
```
### Run as standalone mode
```
swan --masters=192.168.1.50:5050 --addr=0.0.0.0:9999 --raftid=1 --cluster=http://127.0.0.1:2111 --sock=./data/swam.sock
```

### Run as HA mode
```
swan --masters=192.168.1.50:5050 --addr=0.0.0.0:9999 --raftid=1 --cluster=http://127.0.0.1:2111,http://127.0.0.1:2112,http://127.0.0.1:2113 --sock=./data/swam1.sock
swan --masters=192.168.1.50:5050 --addr=0.0.0.0:9998 --raftid=2 --cluster=http://127.0.0.1:2111,http://127.0.0.1:2112,http://127.0.0.1:2113 --sock=./data/swam2.sock
swan --masters=192.168.1.50:5050 --addr=0.0.0.0:9997 --raftid=3 --cluster=http://127.0.0.1:2111,http://127.0.0.1:2112,http://127.0.0.1:2113 --sock=./data/swam3.sock
```
Use `swan --help` to see usage.

## Getting Started
### Use `curl` 

+ applicaiton deloyment
```
curl -X POST -H "Content-Type: application/json" -d@example.json http://localhost:9999/v1/apps
```
+ application delete
```
curl -X DELETE http://localhost:9999/v1/apps/nginx0003
```
+ application show
```
curl http://localhost:9999/v1/apps/nginx0003
```
+ applications list
```
curl http://localhost:9999/v1/apps
```
+ application scaling
```
curl -X POST http://localhost:9999/v1/apps/nginx0003/scale?instances=10
```
`instances` is used to specified the instances count you wanna scale to.
  
+ application rolling update
```
curl -X POST -H "Content-Type: application/json" -d@new_verison.json http://localhost:9999/v1/apps/nginx0003/update\?instances\=-1
```  
`instances` -1 means updating all instances. other value means updating the specified instances at one time.
  
+ application versions
```
curl http://localhost:9999/v1/apps/nginx0003/versions
```

### Use command line client `swancfg`
```
cd cli
make && make install
```
`swancfg --help` for usage.

## Roadmap
See [ROADMAP](https://github.com/Dataman-Cloud/swan/blob/master/ROADMAP.md) for the full roadmap.

## Contributing
If you want to contribute to swan, make a PR or report a issue. 
The goal of swan is to become the default and best scheduler for mesos, so let's do it!

## Licensing
Swan is licensed under the Apache License, Version 2.0. See 
[LICENSE](https://github.com/Dataman-Cloud/swan/blob/master/LICENSE) for the full
license text.
