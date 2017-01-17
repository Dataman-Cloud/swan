
<img src="docs/assets/img/swan.png" width="350">

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
+ Rolling upgrade
+ Version rollback
+ Health check
+ Auto failover
+ High Availability with Raft backend
+ Build in HTTP Proxy, Load Balance
+ Build in DNS

## Special features
+ The instance name is fixed during the application lifecycle.
+ The instance index is continuously incremented from zero.

## Installation
### From Source(swan only)
First get the swan:
```
go get github.com/Dataman-Cloud/swan
```
Then you can compile `swan` with:
```
make

or 

make docker-build
```
`swan` will be installed at $GOPATH/bin/swan, If `$GOPATH/bin` is in your `PATH`, you can invoke `swan` from the CLI.

## Run as standalone mode
```
goreman start

or 

make docker-run
```

### Run in HA mode
```
swan --mesos-master=zk://127.0.0.1:2181/mesos --cluster=127.0.0.1:9999,127.0.0.1:9998,127.0.0.1:9997 --raftid=1 --raft-cluster=http://127.0.0.1:2111,http://127.0.0.1:2112,http://127.0.0.1:2113 --work-dir=./data/ --mode=mixed
swan --mesos-master=zk://127.0.0.1:2181/mesos --cluster=127.0.0.1:9999,127.0.0.1:9998,127.0.0.1:9997 --raftid=2 --raft-cluster=http://127.0.0.1:2111,http://127.0.0.1:2112,http://127.0.0.1:2113 --work-dir=./data/ --mode=mixed
swan --mesos-master=zk://127.0.0.1:2181/mesos --cluster=127.0.0.1:9999,127.0.0.1:9998,127.0.0.1:9997 --raftid=3 --raft-cluster=http://127.0.0.1:2111,http://127.0.0.1:2112,http://127.0.0.1:2113 --work-dir=./data/ --mode=mixed
```
Use `swan --help` to see usage.

## Getting Started
### Use `curl` 

+ applicaiton deloyment
```
curl -X POST -H "Content-Type: application/json" -d@example/template-replicates.json http://localhost:9999/v_beta/apps
```

+ applications list
```
curl http://localhost:9999/v_beta/apps
```

+ application show
```
curl http://localhost:9999/v_beta/apps/nginx0003-xcm-unnamed
```

+ application delete
```
curl -X DELETE http://localhost:9999/v_beta/apps/nginx0003-xcm-unnamed
```

+ application scale up
```
curl -X PATCH -H "Content-Type: application/json" http://localhost:9999/v_beta/apps/nginx0003-xcm-unnamed/scale-up -d@example/scale.json
```

+ application scale down
```
curl -X PATCH -H "Content-Type: application/json" http://localhost:9999/v_beta/apps/nginx0003-xcm-unamed/scale-down -d@example/scale.json
```

+ application rolling update
```
curl -X POST -H "Content-Type: application/json" -d@new_verison.json http://localhost:9999/v_beta/apps/nginx0003-xcm-unnamed
```

`instances` -1 means updating all instances. other value means updating the specified instances at one time.

+ list application versions
```
curl http://localhost:9999/v_beta/apps/nginx0003-xcm-unnamed/versions
```

+ get application version
```
curl
http://localhost:9999/v_beta/apps/nginx0003-xcm-unnamed/versions/14012934223
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
