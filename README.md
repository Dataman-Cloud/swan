
<img src="docs/assets/img/swan.png" width="350">

##

[![Build Status](https://travis-ci.org/Dataman-Cloud/swan.svg?branch=master)](https://travis-ci.org/Dataman-Cloud/swan)
[![codecov](https://codecov.io/gh/Dataman-Cloud/swan/branch/master/graph/badge.svg)](https://codecov.io/gh/Dataman-Cloud/swan)
[![Go Report Card](https://goreportcard.com/badge/github.com/Dataman-Cloud/swan)](https://goreportcard.com/report/github.com/Dataman-Cloud/swan)
[![Join the chat at https://gitter.im/Dataman-Cloud/swan](https://badges.gitter.im/Dataman-Cloud/swan.svg)](https://gitter.im/Dataman-Cloud/swan?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

#### swan is a mesos scheduling framework written in golang based on mesos new HTTP API.

#### you can use swan to deployment application on mesos cluster, and manage the entire lifecycle of the application. You can do rolling-update with new versions, you can scale applications, and you can do health checks for your applications and auto failover when applications or services are not available.

#### swan is maintained by [dataman-cloud](https://github.com/Dataman-Cloud), and licensed under the Apache License, Version 2.0. 

#### for more documentation about Swan please refer to [swan-docs](https://github.com/Dataman-Cloud/swan/docs/)


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

## Installation

### From Source

clone `Swan` source code from github.com:
```
git clone git@github.com:Dataman-Cloud/swan.git
```
Then you can compile `Wwan` with:
```
make

# or build within docker

make docker-build

## Run as standalone mode
```
goreman start

or run swan in docker

make docker-run-mixed
```

Use `swan --help` to see usage.

## Getting Started
### Use `curl`

+ application deployment
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

+ application rolling upgrade
```
curl -X PUT -H "Content-Type: application/json" -d@new_verison.json http://localhost:9999/v_beta/apps/nginx0003-xcm-unnamed
```

+ proceed upgrade process
```
curl -X PATCH -H "Content-Type: application/json" http://localhost:9999/v_beta/apps/nginx0003-xcm-unnamed/proceed-update -d'{"instances": $NUM}'
```

`instances` -1 means upgrading all instances at once. Any other value specifies the number of instances to be updated at the same time.

+ cancel upgrade process
```
curl -X PATCH -H "Content-Type: application/json" http://localhost:9999/v_beta/apps/nginx0003-xcm-unnamed/cancel-update
```

+ list application versions
```
curl http://localhost:9999/v_beta/apps/nginx0003-xcm-unnamed/versions
```

+ get application version
```
curl
http://localhost:9999/v_beta/apps/nginx0003-xcm-unnamed/versions/14012934223
```


## Roadmap
See [ROADMAP](https://github.com/Dataman-Cloud/swan/blob/master/ROADMAP.md) for the full roadmap.

## Contributing
If you want to contribute to swan, make a PR or report a issue. 
The goal of swan is to become the default and best scheduler for mesos, so let's do it!

## Licensing
Swan is licensed under the Apache License, Version 2.0. See 
[LICENSE](https://github.com/Dataman-Cloud/swan/blob/master/LICENSE) for the full
license text.
