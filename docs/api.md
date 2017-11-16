+ apps
  - [GET  /v1/apps](#list-all-apps)  *List all apps*
  - [POST /v1/apps](#create-a-app)   *Create a app*
  - [GET /v1/apps/{app_id}](#inspect-a-app) *Inspect a app*
  - [DELETE /v1/apps/{app_id}](#delete-a-app) *Delete a app*
  - [POST /v1/apps/{app_id}/scale](#scale-up-down) *Scale up-down*
  - [PUT /v1/apps/{app_id}](#rolling-update) *Rolling update a app*
  - [POST /v1/apps/{app_id}/rollback](#roll-back) *Roll back a app*
  - [PUT /v1/apps/{app_id}/canary](#canary-update-a-app) *Canary update a app*
  - [PUT /v1/apps/{app_id}/weights](#update-weights) *Update tasks's weights*

+ tasks
  - [GET /v1/apps/{app_id}/tasks](#list-all-tasks-for-a-app) *List all tasks for a app*
  - [GET /v1/apps/{app_id}/tasks/{task_id}](#inspect-a-app) *Inspect a task*
  - [PUT /v1/apps/{app_id}/tasks/{task_id}/weight](#update-weight) *Update task's weight*

+ versions
  - [GET /v1/apps/{app_id}/versions](#list-all-versions-for-a-app) *List all versions for a app*
  - [GET /v1/apps/{app_id}/versions/{version_id}](#inspect-a-version) *Inpect a version*

+ dns/proxy/traffics
  - [GET /v1/apps/{app_id}/dns](#list-all-dns-for-a-app) *List all dns records for a app*
  - [GET /v1/apps/{app_id}/dns/traffics](#list-dns-traffics-for-a-app) *List dns traffics for a app*
  - [GET /v1/apps/{app_id}/proxy](#list-all-proxy-for-a-app) *List all proxy records for a app*
  - [GET /v1/apps/{app_id}/proxy/traffics](#list-proxy-traffics-for-a-app) *List proxy traffics for a app*

+ compose
  - ~[compose](https://github.com/Dataman-Cloud/swan/tree/master/docs/compose.md)~ **Deprecated**
  - [compose-ng](https://github.com/Dataman-Cloud/swan/tree/master/docs/compose-ng.md)

+ kvm
  - [kvm](https://github.com/Dataman-Cloud/swan/tree/master/docs/kvm.md)

+ framework
  - [GET /v1/framework](#framework) *Framework Info*

+ events
  - [GET /v1/events](#) *Event Subscription*

+ health
  - [GET /ping](#ping) *Health check*
 
+ leader
  - [GET /v1/leader](#leader) *Inspect leader info*

+ version
  - [GET /version](#version) *Version information*

+ debug
  - [GET /v1/debug/dump](#dump)
  - [GET /v1/debug/load](#load)

+ agents
  - [GET /v1/agents](#list-agents) *List all agents*
  - [GET /v1/agents/query_id](#query-agent-id) *Query mesos slave id by ip addresses (internal use)*
  - [GET /v1/agents/{agent_id}](#get-agent) *Get specified agent*
  - [DELETE /v1/agents/{agent_id}](#close-agent) *Disconnect specified agent*
  - [GET /v1/agents/{agent_id}/dns](#get-agent-dns) *Get dns records on specified agent*
  - [GET /v1/agents/{agent_id}/dns/stats](#get-agent-dns-stats) *Get dns traffics stats on specified agent*
  - [GET /v1/agents/{agent_id}/proxy](#get-agent-proxy) *Get proxy records on specified agent*
  - [GET /v1/agents/{agent_id}/proxy/stats](#get-agent-proxy-stats) *Get proxy traffics stats on specified agent*

+ ipam
  - [PUT /v1/agents/{agent_id}/ipam/subnets](#set-ipam-pool-range) *Set ip pool range*
  - [GET /v1/agents/{agent_id}/ipam/subnets](#get-ipam-pool-usage) *List ipam pool usage*

+ networks
  - [GET /v1/agents/networks](#swan-driven-networks) *List swan ipam driven docker networks*

+ task-over-docker-remote-api
  - [ANY /v1/agents/{agent_id}/docker/{standard_docker_remote_api}](#take-over-agent-docker-remote-api) *Redirect agent docker remote API through swan api*

+ reset 
  - [POST /v1/apps/{app_id}/reset](#reset)

+ [deploy policy](https://github.com/Dataman-Cloud/swan/tree/master/docs/deploy.md)

+ [constraints](https://github.com/Dataman-Cloud/swan/tree/master/docs/constraints.md)

+ [scale](https://github.com/Dataman-Cloud/swan/tree/master/docs/scale.md)
 
+ [update policy](https://github.com/Dataman-Cloud/swan/tree/master/docs/update.md)

+ [port mapping](https://github.com/Dataman-Cloud/swan/tree/master/docs/port-mapping.md)

+ [status](https://github.com/Dataman-Cloud/swan/tree/master/docs/status.md)

#### List all apps
```
GET /v1/apps 
```
Example request:
```
GET /v1/apps
```
Example response:
```
[
    {
        "cluster": "dataman2",
        "created": "2017-09-06T10:39:17.784738187+08:00",
        "currentVersion": [
            "1504665557784509256"
        ],
        "errmsg": "",
        "health": {
            "healthy": 0,
            "total": 10,
            "unhealthy": 0,
            "unset": 10
        },
        "id": "nginx15.default.xcm.dataman2",
        "name": "nginx15",
        "operationStatus": "noop",
        "progress": -1,
        "progress_details": null,
        "runAs": "xcm",
        "status": "available",
        "task_count": 10,
        "tasks_status": {
            "TASK_RUNNING": 10
        },
        "updated": "2017-09-06T10:39:24.049790305+08:00",
        "version_count": 1
    }
]
```
Query parameters:
```
```

#### Create a app
```
POST /v1/apps
```
Example request:
```
 POST /v1/apps HTTP/1.1
 Content-Type: application/json
 
 {
  "name": "nginx002",
  "cmd": null,
  "args": null,
  "cpus": 0.01,
  "mem": 32,
  "disk": 0,
  "runAs": "xcm",
  "cluster": "cctv",
  "instances": 10,
  "constraints": [],
  "container": {
    "docker": {
      "image": "nginx",
      "network": "bridge",
      "forcePullImage": false,
      "privileged": true,
      "parameters": [
        {
            "key": "label",
            "value": "APP_ID=wordpress"
        }
      ],
      "portMappings": [
              {
                      "name": "web",
                      "protocol": "tcp",
                      "containerPort": 80,
                      "hostPort": 80
              }
      ]
    },
    "type": "DOCKER",
    "volumes": [
      {
        "hostPath": "/home",
        "containerPath": "/data",
        "mode": "RW"
      }
  ]
  },
  "env": {
    "WORDPRESS_DB_HOST": "192.168.1.210",
    "WORDPRESS_DB_PASSWORD": "root"
  },
  "uris": [
  ],
  "labels": {
    "USER_ID": "1"
  },
  "healthCheck":
    {
      "protocol": "http",
      "path": "/",
      "delaySeconds": 2,
      "gracePeriodSeconds": 5,
      "intervalSeconds": 1,
      "portName": "web",
      "timeoutSeconds": 1,
      "consecutiveFailures": 5
    },
  "proxy": {
      "enabled": false,
      "alias": "www.example.com",
      "listen": 9999,
      "sticky": false
  }
}
```
Json Parameters:

+ **name**(required): the name of the appliation.
+ **runAs**(required): the user or group the app belong to.
+ **cluster**(optional): the virtaul cluster name app run in, if not set, the real physic mesos cluster name will be used. 
+ **cpus**(required): the cpus used for each container.
+ **mem**(required): the memory used for each container.
+ **disk**(optional): the disk space allcated to each container.
+ **instances**(required): the container count for this application.
+ **cmd**(optional): the command to be launched.
+ **args**(optional): the arguments for the command.
+ **container**(required): the container related configuration.
+ **image**(required): the image to used for run container.
+ **network**(required): the network mode used for container. possible values are:
```
bridge
host
custom network name, eg. swan
```
+ **forcePullImage**(optional): whether to pull image force or not event it exists. default is false.
+ **privileged**(optional): whether to give extended privileges to container. default is false.
+ **parameters**(optional): docker parameters inject in container at runtime.
+ **portMappings**(required): the mapping between host port and container port. see https://github.com/Dataman-Cloud/swan/tree/master/docs/port-mapping.md
+ **type**(required): the containerizer used for container. possible values are `DOCKER` and `MESOS`, currently support `DOCKER`. 
+ **volume**(optional): the volume mounted to container. values for `mode` are `RW` or `RO`.
+ **env**(optional): the enviroment inject in container at runtime.
+ **uris**(optional): the resource willed be downloaded in container sandbox befor run.
+ **labels**(optinal): the container labels
+ **healthCheck**(optional): the health check configuration for container. see https://github.com/Dataman-Cloud/swan/tree/master/docs/health-check.md.
+ **proxy**(optional): the proxy configuration for app. see https://github.com/Dataman-Cloud/swan/tree/master/docs/proxy.md

Example response:
```
  HTTP/1.1 201 Created
  Content-Type: application/json

  {
       "id":"nginx0r1.default.xcm.dataman"
  }
```
#### Inspect a app
```
GET /v1/apps/{app_id}
```
Example request:
```
GET /v1/apps/nginx0r2.default.xcm.dataman
```
Example response:
```
HTTP/1.1 200 OK
Content-Type: application/json

{
    "cluster": "dataman2",
    "created": "2017-09-06T10:39:17.784738187+08:00",
    "currentVersion": [
        "1504665557784509256"
    ],
    "errmsg": "",
    "health": {
        "healthy": 0,
        "total": 10,
        "unhealthy": 0,
        "unset": 10
    },
    "id": "nginx15.default.xcm.dataman2",
    "name": "nginx15",
    "operationStatus": "noop",
    "progress": -1,
    "progress_details": null,
    "runAs": "xcm",
    "status": "available",
    "task_count": 10,
    "tasks_status": {
        "TASK_RUNNING": 10
    },
    "updated": "2017-09-06T10:39:24.049790305+08:00",
    "version_count": 1
}

```
Json Parameters:

+ **currentVersion**: the current version for application. If app is in updating, this field will be has mutiple value. 
+ **errmsg**: the error message while application deployment.
+ **health**: the health status summary of all the tasks for the application
+ **operationStatus**: current operation for app. possible values are:
```
noop
creating
scaling_up
scaling_down
updating
canary_updating
canary_unfinished
weight_updating
starting
stopping
deleting
rollbacking
```
+ **progress**: the tasks count has been updated. this field only meaningful in application updating.
+ **progress_details**: indicated the task has been updated or not. this field only meaningful in application updating. 
+ **task_status**: the status summary of all the tasks for the application. possiable values are:
```
pending
TASK_STAGING
TASK_STARTING
TASK_RUNNING
TASK_KILLING
TASK_FINISHED
TASK_FAILED
TASK_KILLED
TASK_ERROR
TASK_LOST
TASK_DROPPED
TASK_UNREACHABLE
TASK_GONE
TASK_GONE_BY_OPERATOR
TASK_UNKNOWN
```
see https://github.com/Dataman-Cloud/swan/tree/master/docs/status.md for detail. 

+ **status**: application status, possible values are:
```
available
unavailable
```

#### Delete a app
```
DELETE /v1/apps/{app_id}
```
Example request:
```
DELETE /v1/apps/nginx0r2.default.xcm.dataman
```
Example response:
```
HTTP/1.1 204 No Content
```
##### Scale up down
```
POST /v1/apps/{app_id}/scale
```
Example request:
```
POST /v1/apps/nginx0r2.default.xcm.dataman/scale HTTP/1.1
Content-Type: application/json
{
    "instances": 5, 
    "ips": [],
}
```
Json parameters:
```
instances     : the goal to scale up/down
ips(optional) : ip list for static ip(brige or host or scale down ignore). if this field is not set or empty, the ip address will be auto-allcated from ipam.
```
Example response:
```
HTTP/1.1 202 Accepted
```

#### Rolling update 

```
PUT /v1/apps/{app_id}
```
Example request:
```
 PUT /v1/apps/nginx004.default.testuser.dataman HTTP/1.1
 Content-Type: application/json
 
 {
  "name": "nginx002",
  "cmd": null,
  "args": null,
  "cpus": 0.01,
  "mem": 32,
  "disk": 0,
  "runAs": "xcm",
  "instances": 10,
  "constraints": [],
  "container": {
    "docker": {
      "image": "nginx",
      "network": "bridge",
      "forcePullImage": false,
      "privileged": true,
      "parameters": [
        {
            "key": "label",
            "value": "APP_ID=wordpress"
        }
      ],
      "portMappings": [
              {
                      "name": "web",
                      "protocol": "tcp",
                      "containerPort": 80,
                      "hostPort": 80
              }
      ]
    },
    "type": "DOCKER",
    "volumes": [
      {
        "hostPath": "/home",
        "containerPath": "/data",
        "mode": "RW"
      }
  ]
  },
  "env": {
    "WORDPRESS_DB_HOST": "192.168.1.210",
    "WORDPRESS_DB_PASSWORD": "root"
  },
  "uris": [
  ],
  "labels": {
    "USER_ID": "1"
  },
  "healthCheck": {
      "protocol": "http",
      "path": "/",
      "delaySeconds": 2,
      "gracePeriodSeconds": 5,
      "intervalSeconds": 1,
      "portName": "web",
      "timeoutSeconds": 1,
      "consecutiveFailures": 5
  },
  "proxy": {
       "enabled": false,
       "alias": ""
       "listen": 99,
       "sticky": false
  },
  "update": {
      "delay": 5,
      "onfailure": "continue"
  }
}
```
Json Parameters:
+ **delay**: the delay seconds between two updates.
+ **onfailure**: the action while updating failure, possible values are:
```
stop
continue
```
more details see https://github.com/Dataman-Cloud/swan/tree/master/docs/update.md

Example response:
```
  HTTP/1.1 202 Accepted 
```

#### Roll back
```
POST /v1/apps/{app_id}/rollback
```
Example request:
```
POST /v1/apps/nginx0r2.default.xcm.dataman/rollback
```
```
Rollback will update app to the previous version.
```
Example response:
```
HTTP/1.1 202 Accepted
```

#### List all tasks for a app

```
GET /v1/apps/{app_id}/tasks
```
Example request:
```
GET /v1/apps/nginx0r1.default.xcm.dataman/tasks
```
Example response:
```json
HTTP/1.1 200 OK
Content-Type: application/json
[
    {
        "agentId": "07239dfa-2982-4046-8f1c-f8f3c14d4649-S4",
        "container_id": "69a08b075328ce380b4f9b3668294d0af0a52ca02606ae1026479425e7668257",
        "container_name": "/mesos-07239dfa-2982-4046-8f1c-f8f3c14d4649-S4.5b9bf6ea-4f7b-4aef-b08b-fd980ebf7e06",
        "created": "2017-09-06T10:39:17.872672561+08:00",
        "errmsg": "",
        "healthy": "unset",
        "histories": null,
        "id": "31defd3cf9f2.4.nginx15.default.xcm.dataman2",
        "ip": "172.16.1.18",
        "maxRetries": 0,
        "name": "4.nginx15.default.xcm.dataman2",
        "opstatus": "",
        "ports": [
            31004
        ],
        "retries": 0,
        "status": "TASK_RUNNING",
        "updated": "2017-09-06T10:39:17.872672595+08:00",
        "version": "1504665557784509256",
        "weight": 100
    }
]

```

#### Canary update a app
```
PUT /v1/apps/{app_id}/canary
```

Example request:
```
PUT /v1/apps/nginx0r1.default.xcm.dataman/canary
Content-Type: application/json
 
 {
    "version": {
        "name": "nginx002",
        "cmd": null,
        "args": null,
        "cpus": 0.01,
        "mem": 32,
        "disk": 0,
        "runAs": "xcm",
        "instances": 10,
        "constraints": [],
        "container": {
          "docker": {
            "image": "nginx",
            "network": "bridge",
            "forcePullImage": false,
            "privileged": true,
            "parameters": [
              {
                  "key": "label",
                  "value": "APP_ID=wordpress"
              }
            ],
            "portMappings": [
                    {
                            "name": "web",
                            "protocol": "tcp",
                            "containerPort": 80,
                            "hostPort": 80
                    }
            ]
          },
          "type": "DOCKER",
          "volumes": [
            {
              "hostPath": "/home",
              "containerPath": "/data",
              "mode": "RW"
            }
        ]
        },
        "env": {
          "WORDPRESS_DB_HOST": "192.168.1.210",
          "WORDPRESS_DB_PASSWORD": "root"
        },
        "uris": [
        ],
        "label": {
          "USER_ID": "1"
        },
        "healthCheck":
          {
            "protocol": "http",
            "path": "/",
            "delaySeconds": 2,
            "gracePeriodSeconds": 5,
            "intervalSeconds": 1,
            "portName": "web",
            "timeoutSeconds": 1,
            "consecutiveFailures": 5
          },
        "proxy": {
                  "enabled": false,
                  "alias": ""
        },
        "update": {
            "delay": 5,
            "onfailure": "continue"
        }
    },
    "instances": 5,
    "value": 0.1,
    "delay": 5,
    "onFailure": "stop"
}
```
Json Parameters:
```
version: (types.Version) the new version to be updated to, can be empty(null).

instances: (int) the task count to be updated to new version.

delay: (float) the delay seconds between two updates.

onFailure:(string) the action when update failed.
```
Example response:
```
  HTTP/1.1 202 Accepted 
```

```json
{
  "id": "demo.default.bbk.dataman-mesos",
  "name": "demo",
  "runAs": "bbk",
  "cluster": "dataman-mesos",
  "operationStatus": "canary_unfinished", // canary_updating, canary_unfinished, noop
  "progress": 3,						  // task progress count
  "progress_details": {					  // task progress details
    "106af6135cf0.0.demo.default.bbk.dataman-mesos": true,
    "1a0959454228.2.demo.default.bbk.dataman-mesos": true,
    "641a0ad32e5b.4.demo.default.bbk.dataman-mesos": false,
    "6dd5300ad5e5.3.demo.default.bbk.dataman-mesos": false,
    "a33c06d46e3e.1.demo.default.bbk.dataman-mesos": true
  },
  "task_count": 5,
  "currentVersion": [
    "1502183593944171595",
    "1502183799949971040"
  ],
  "version_count": 3,
  "status": "available",
  "tasks_status": {
    "TASK_RUNNING": 5
  },
  "health": {
    "total": 5,
    "healthy": 5,
    "unhealthy": 0,
    "unset": 0
  },
  "errmsg": "",
  "created": "2017-08-08T17:12:25.979196132+08:00",
  "updated": "2017-08-08T17:16:58.704153869+08:00"
}
```


#### List all versions for a app

```
GET /v1/apps/{app_id}/versions
```
Example request:
```
GET /v1/apps/nginx0r1.default.xcm.dataman/versions
```
Example response:
```json
HTTP/1.1 200 OK
Content-Type: application/json

[
    {
        "cluster": "dataman2",
        "cmd": "",
        "constraints": [
            {
                "attribute": "disk",
                "operator": "==",
                "value": "ssd"
            }
        ],
        "container": {
            "docker": {
                "image": "nginx",
                "network": "bridge",
                "parameters": [
                    {
                        "key": "ipc",
                        "value": "host"
                    },
                    {
                        "key": "workdir",
                        "value": "/data"
                    }
                ],
                "portMappings": [
                    {
                        "containerPort": 80,
                        "hostPort": 80,
                        "name": "web",
                        "protocol": "tcp"
                    }
                ],
                "privileged": true
            },
            "type": "DOCKER",
            "volumes": [
                {
                    "containerPath": "/data",
                    "hostPath": "/home",
                    "mode": "RW"
                }
            ]
        },
        "cpus": 0.01,
        "disk": 0,
        "env": {
            "WORDPRESS_DB_HOST": "dbhost",
            "WORDPRESS_DB_PASSWORD": "password"
        },
        "gpus": 0,
        "healthCheck": null,
        "id": "1504665557784509256",
        "instances": 10,
        "ips": null,
        "kill": null,
        "labels": null,
        "mem": 10,
        "name": "nginx15",
        "proxy": {
            "alias": "",
            "enabled": false,
            "listen": 0,
            "sticky": false
        },
        "restart": {
            "retries": 0
        },
        "runAs": "xcm",
        "update": null,
        "uris": []
    }
]
```

#### Inspect a task

```
GET /v1/apps/{app_id}/tasks/{task_id}
```
Example request:
```
GET /v1/apps/nginx0r1.default.xcm.dataman/tasks/e6404f0324d2.0.nginx0r1.default.xcm.dataman
```
Example response:
```json
HTTP/1.1 200 OK
Content-Type: application/json

{
    "agentId": "07239dfa-2982-4046-8f1c-f8f3c14d4649-S4",
    "container_id": "7d3c4d4e62b0696f5ee0ed7ccabdc17f83716c6eaedd9e5cc51a3eca3e6a8b05",
    "container_name": "/mesos-07239dfa-2982-4046-8f1c-f8f3c14d4649-S4.7978406a-bb81-4a0c-a182-31abc4f483f9",
    "created": "2017-09-06T10:39:17.848361457+08:00",
    "errmsg": "",
    "healthy": "unset",
    "histories": null,
    "id": "9a3adf023c4c.2.nginx15.default.xcm.dataman2",
    "ip": "172.16.1.18",
    "maxRetries": 0,
    "name": "2.nginx15.default.xcm.dataman2",
    "opstatus": "",
    "ports": [
        31002
    ],
    "retries": 0,
    "status": "TASK_RUNNING",
    "updated": "2017-09-06T10:39:17.848361519+08:00",
    "version": "1504665557784509256",
    "weight": 100
}

```

#### Inspect a version 
```
GET /v1/apps/{app_id}/versions/{version_id}
```
Example request:
```
GET /v1/apps/nginx0r1.default.xcm.dataman/versions/1498029948754163146
```
Example response:
```json
HTTP/1.1 200 OK
Content-Type: application/json

{
    "cluster": "dataman2",
    "cmd": "",
    "constraints": [
        {
            "attribute": "disk",
            "operator": "==",
            "value": "ssd"
        }
    ],
    "container": {
        "docker": {
            "image": "nginx",
            "network": "bridge",
            "parameters": [
                {
                    "key": "ipc",
                    "value": "host"
                },
                {
                    "key": "workdir",
                    "value": "/data"
                }
            ],
            "portMappings": [
                {
                    "containerPort": 80,
                    "hostPort": 80,
                    "name": "web",
                    "protocol": "tcp"
                }
            ],
            "privileged": true
        },
        "type": "DOCKER",
        "volumes": [
            {
                "containerPath": "/data",
                "hostPath": "/home",
                "mode": "RW"
            }
        ]
    },
    "cpus": 0.01,
    "disk": 0,
    "env": {
        "WORDPRESS_DB_HOST": "dbhost",
        "WORDPRESS_DB_PASSWORD": "password"
    },
    "gpus": 0,
    "healthCheck": null,
    "id": "1504665557784509256",
    "instances": 10,
    "ips": null,
    "kill": null,
    "labels": null,
    "mem": 10,
    "name": "nginx15",
    "proxy": {
        "alias": "",
        "enabled": false,
        "listen": 0,
        "sticky": false
    },
    "restart": {
        "retries": 0
    },
    "runAs": "xcm",
    "update": null,
    "uris": []
}

```
#### Create a version 
```
POST /v1/apps/{app_id}/versions
```
Example request:
```
 POST /v1/apps/{nginx004.default.testuser.dataman}/versions HTTP/1.1
 Content-Type: application/json
 
 {
  "name": "nginx002",
  "cmd": null,
  "args": null,
  "cpus": 0.01,
  "mem": 32,
  "disk": 0,
  "runAs": "xcm",
  "instances": 10,
  "constraints": [],
  "container": {
    "docker": {
      "image": "nginx",
      "network": "bridge",
      "forcePullImage": false,
      "privileged": true,
      "parameters": [
        {
            "key": "label",
            "value": "APP_ID=wordpress"
        }
      ],
      "portMappings": [
              {
                      "name": "web",
                      "protocol": "tcp",
                      "containerPort": 80,
                      "hostPort": 80
              }
      ]
    },
    "type": "DOCKER",
    "volumes": [
      {
        "hostPath": "/home",
        "containerPath": "/data",
        "mode": "RW"
      }
  ]
  },
  "env": {
    "WORDPRESS_DB_HOST": "192.168.1.210",
    "WORDPRESS_DB_PASSWORD": "root"
  },
  "uris": [
  ],
  "labels": {
    "USER_ID": "1"
  },
  "healthCheck":
    {
      "protocol": "http",
      "path": "/",
      "delaySeconds": 2,
      "gracePeriodSeconds": 5,
      "intervalSeconds": 1,
      "portName": "web",
      "timeoutSeconds": 1,
      "consecutiveFailures": 5
    },
  "proxy": {
            "enabled": false,
            "alias": ""
  }
}
```
Example response:
```
  HTTP/1.1 201 Created
  Content-Type: application/json

  {
      "id":"1498791358276219465"
  }
```
#### Update weights 
```
PUT /v1/apps/{app_id}/weights
```
Example request:
```
 PUT /v1/apps/nginx004.default.testuser.dataman/weights HTTP/1.1
 Content-Type: application/json
 
 {
	"value": 0.1
 }
```
Json parameters:
```
Value(float) : traffic for new version relative to all versions. 
```
Example response:
```
HTTP/1.1 202 Accepted
```

#### Update weight
```
PUT /v1/apps/{app_id}/tasks/{task_id}/weight
```
Example request:
```
PUT /v1/apps/nginx004.default.testuser.dataman}/tasks/0.nginx004.default.testuser.dataman/weight HTTP/1.1
Content-Type: application/json
{
    Weight: 50,
}
```
Json parameters:
```
Weight: the task weight
```
Example response:
```
HTTP/1.1 202 Accepted
```

#### List all dns for a app
```
GET /v1/apps/{app_id}/dns
```

```json
{
  "3264208446845635": [ // agent id
    {
      "clean_name": "2.demo.default.bbk.dataman-mesos.bbklab.net.",
      "id": "872e5253d705.2.demo.default.bbk.dataman-mesos",
      "ip": "192.168.1.130",
      "parent": "demo.default.bbk.dataman-mesos",
      "port": "31004",
      "proxy_record": false,
      "weight": 100
    },
    {
      "clean_name": "1.demo.default.bbk.dataman-mesos.bbklab.net.",
      "id": "df98d3f720f2.1.demo.default.bbk.dataman-mesos",
      "ip": "192.168.1.130",
      "parent": "demo.default.bbk.dataman-mesos",
      "port": "31002",
      "proxy_record": false,
      "weight": 100
    },
    {
      "clean_name": "0.demo.default.bbk.dataman-mesos.bbklab.net.",
      "id": "b1a86b5f2b3f.0.demo.default.bbk.dataman-mesos",
      "ip": "192.168.1.130",
      "parent": "demo.default.bbk.dataman-mesos",
      "port": "31000",
      "proxy_record": false,
      "weight": 100
    },
    {
      "clean_name": "4.demo.default.bbk.dataman-mesos.bbklab.net.",
      "id": "35ff224db20c.4.demo.default.bbk.dataman-mesos",
      "ip": "192.168.1.130",
      "parent": "demo.default.bbk.dataman-mesos",
      "port": "31008",
      "proxy_record": false,
      "weight": 100
    },
    {
      "clean_name": "3.demo.default.bbk.dataman-mesos.bbklab.net.",
      "id": "7377d5c1cd9e.3.demo.default.bbk.dataman-mesos",
      "ip": "192.168.1.130",
      "parent": "demo.default.bbk.dataman-mesos",
      "port": "31006",
      "proxy_record": false,
      "weight": 100
    }
  ],
  "3981314045636649": [ // agent id
    {
      "clean_name": "2.demo.default.bbk.dataman-mesos.bbklab.net.",
      "id": "872e5253d705.2.demo.default.bbk.dataman-mesos",
      "ip": "192.168.1.130",
      "parent": "demo.default.bbk.dataman-mesos",
      "port": "31004",
      "proxy_record": false,
      "weight": 100
    },
    {
      "clean_name": "1.demo.default.bbk.dataman-mesos.bbklab.net.",
      "id": "df98d3f720f2.1.demo.default.bbk.dataman-mesos",
      "ip": "192.168.1.130",
      "parent": "demo.default.bbk.dataman-mesos",
      "port": "31002",
      "proxy_record": false,
      "weight": 100
    },
    {
      "clean_name": "0.demo.default.bbk.dataman-mesos.bbklab.net.",
      "id": "b1a86b5f2b3f.0.demo.default.bbk.dataman-mesos",
      "ip": "192.168.1.130",
      "parent": "demo.default.bbk.dataman-mesos",
      "port": "31000",
      "proxy_record": false,
      "weight": 100
    },
    {
      "clean_name": "4.demo.default.bbk.dataman-mesos.bbklab.net.",
      "id": "35ff224db20c.4.demo.default.bbk.dataman-mesos",
      "ip": "192.168.1.130",
      "parent": "demo.default.bbk.dataman-mesos",
      "port": "31008",
      "proxy_record": false,
      "weight": 100
    },
    {
      "clean_name": "3.demo.default.bbk.dataman-mesos.bbklab.net.",
      "id": "7377d5c1cd9e.3.demo.default.bbk.dataman-mesos",
      "ip": "192.168.1.130",
      "parent": "demo.default.bbk.dataman-mesos",
      "port": "31006",
      "proxy_record": false,
      "weight": 100
    }
  ]
}
```

#### List all proxy for a app
```
GET /v1/apps/{app_id}/proxy
```

```json
{
  "3264208446845635": { // node id
    "alias": "g.cn",
    "listen": ":99",
    "name": "demo.default.bbk.dataman-mesos",
    "sticky": false
    "backends": [
      {
        "clean_name": "2.demo.default.bbk.dataman-mesos",
        "id": "872e5253d705.2.demo.default.bbk.dataman-mesos",
        "ip": "192.168.1.130",
        "port": 31004,
        "scheme": "",
        "version": "",
        "weihgt": 100
      },
      {
        "clean_name": "1.demo.default.bbk.dataman-mesos",
        "id": "df98d3f720f2.1.demo.default.bbk.dataman-mesos",
        "ip": "192.168.1.130",
        "port": 31002,
        "scheme": "",
        "version": "",
        "weihgt": 100
      },
      {
        "clean_name": "0.demo.default.bbk.dataman-mesos",
        "id": "b1a86b5f2b3f.0.demo.default.bbk.dataman-mesos",
        "ip": "192.168.1.130",
        "port": 31000,
        "scheme": "",
        "version": "",
        "weihgt": 100
      },
      {
        "clean_name": "4.demo.default.bbk.dataman-mesos",
        "id": "35ff224db20c.4.demo.default.bbk.dataman-mesos",
        "ip": "192.168.1.130",
        "port": 31008,
        "scheme": "",
        "version": "",
        "weihgt": 100
      },
      {
        "clean_name": "3.demo.default.bbk.dataman-mesos",
        "id": "7377d5c1cd9e.3.demo.default.bbk.dataman-mesos",
        "ip": "192.168.1.130",
        "port": 31006,
        "scheme": "",
        "version": "",
        "weihgt": 100
      }
    ]
  },
  "3981314045636649": { // node id
    "alias": "g.cn",
    "listen": ":99",
    "name": "demo.default.bbk.dataman-mesos",
    "sticky": false
    "backends": [
      {
        "clean_name": "2.demo.default.bbk.dataman-mesos",
        "id": "872e5253d705.2.demo.default.bbk.dataman-mesos",
        "ip": "192.168.1.130",
        "port": 31004,
        "scheme": "",
        "version": "",
        "weihgt": 100
      },
      {
        "clean_name": "1.demo.default.bbk.dataman-mesos",
        "id": "df98d3f720f2.1.demo.default.bbk.dataman-mesos",
        "ip": "192.168.1.130",
        "port": 31002,
        "scheme": "",
        "version": "",
        "weihgt": 100
      },
      {
        "clean_name": "0.demo.default.bbk.dataman-mesos",
        "id": "b1a86b5f2b3f.0.demo.default.bbk.dataman-mesos",
        "ip": "192.168.1.130",
        "port": 31000,
        "scheme": "",
        "version": "",
        "weihgt": 100
      },
      {
        "clean_name": "4.demo.default.bbk.dataman-mesos",
        "id": "35ff224db20c.4.demo.default.bbk.dataman-mesos",
        "ip": "192.168.1.130",
        "port": 31008,
        "scheme": "",
        "version": "",
        "weihgt": 100
      },
      {
        "clean_name": "3.demo.default.bbk.dataman-mesos",
        "id": "7377d5c1cd9e.3.demo.default.bbk.dataman-mesos",
        "ip": "192.168.1.130",
        "port": 31006,
        "scheme": "",
        "version": "",
        "weihgt": 100
      }
    ]
  }
}
```
#### List dns traffics for a app
```
GET /v1/apps/{app_id}/dns/traffics
```

```json
{
  "3264208446845635": {
    "authority": 9,
    "fails": 0,
    "forward": 0,
    "requests": 9,
    "type_a": 5,
    "type_srv": 4
  },
  "3981314045636649": {
    "authority": 17,
    "fails": 0,
    "forward": 0,
    "requests": 17,
    "type_a": 2,
    "type_srv": 15
  }
}
```

#### List proxy traffics for a app
```
GET /v1/apps/{app_id}/proxy/traffics
```

```json
{
  "3264208446845635": { // node id
    "35ff224db20c.4.demo.default.bbk.dataman-mesos": {
      "active_clients": 0,
      "requests": 1,
      "requests_rate": 0,
      "rx_bytes": 80,
      "rx_rate": 0,
      "tx_bytes": 851,
      "tx_rate": 0,
      "uptime": "11m15.219579111s"
    },
    "872e5253d705.2.demo.default.bbk.dataman-mesos": {
      "active_clients": 0,
      "requests": 1,
      "requests_rate": 0,
      "rx_bytes": 80,
      "rx_rate": 0,
      "tx_bytes": 851,
      "tx_rate": 0,
      "uptime": "11m15.909388496s"
    }
  },
  "3981314045636649": { // node id
    "35ff224db20c.4.demo.default.bbk.dataman-mesos": {
      "active_clients": 0,
      "requests": 5,
      "requests_rate": 0,
      "rx_bytes": 400,
      "rx_rate": 0,
      "tx_bytes": 4255,
      "tx_rate": 0,
      "uptime": "12m21.369988696s"
    },
    "7377d5c1cd9e.3.demo.default.bbk.dataman-mesos": {
      "active_clients": 0,
      "requests": 1,
      "requests_rate": 0,
      "rx_bytes": 80,
      "rx_rate": 0,
      "tx_bytes": 851,
      "tx_rate": 0,
      "uptime": "11m27.356938773s"
    },
    "872e5253d705.2.demo.default.bbk.dataman-mesos": {
      "active_clients": 0,
      "requests": 6,
      "requests_rate": 0,
      "rx_bytes": 480,
      "rx_rate": 0,
      "tx_bytes": 5106,
      "tx_rate": 0,
      "uptime": "12m20.339854227s"
    },
    "b1a86b5f2b3f.0.demo.default.bbk.dataman-mesos": {
      "active_clients": 0,
      "requests": 4,
      "requests_rate": 0,
      "rx_bytes": 320,
      "rx_rate": 0,
      "tx_bytes": 3404,
      "tx_rate": 0,
      "uptime": "12m20.029829649s"
    },
    "df98d3f720f2.1.demo.default.bbk.dataman-mesos": {
      "active_clients": 0,
      "requests": 2,
      "requests_rate": 0,
      "rx_bytes": 160,
      "rx_rate": 0,
      "tx_bytes": 1702,
      "tx_rate": 0,
      "uptime": "12m19.142783872s"
    }
  }
}
```

### IPAM

#### set ipam pool range
```
# by CLI
./swan agent ipam  --ip-start=192.168.1.199/24 --ip-end=192.168.1.209/24

# by HTTP API
agent_id 为任意一个节点
PUT /v1/agents/{agent_id}/ipam/subnets
```

```json
{
  "ip_start": "192.168.1.199/24",
  "ip_end": "192.168.1.209/24"
}
```

#### get ipam pool usage
```
agent_id 为任意一个节点
GET /v1/agents/{agent_id}/ipam/subnets
```

```json
{
  "192.168.1.0": {
    "ips": [
      [
        "192.168.1.208",
        "false"
      ],
      [
        "192.168.1.209",
        "false"
      ],
      [
        "192.168.1.206",
        "false"
      ],
      [
        "192.168.1.199",
        "true"
      ],
      [
        "192.168.1.207",
        "false"
      ],
      [
        "192.168.1.204",
        "false"
      ],
      [
        "192.168.1.205",
        "false"
      ],
      [
        "192.168.1.202",
        "false"
      ],
      [
        "192.168.1.203",
        "false"
      ],
      [
        "192.168.1.200",
        "true"
      ],
      [
        "192.168.1.201",
        "false"
      ]
    ],
    "subnet": {
      "id": "192.168.1.0",
      "cidr": "192.168.1.0/24",
      "ipnet": "192.168.1.0/24",
      "ip_start": "192.168.1.0",
      "ip_end": "192.168.1.255",
      "mask": 24
    }
  }
}
```

### Networks

#### swan driven networks
```
GET /v1/agents/networks
```

```json
{
  "swan": "192.168.1.0/24",
  "test": "172.16.0.0/16"
}
```

Debug: 列出节点上所有的容器网络，包括docker内置网络
```
GET /v1/agents/networks?debug=true
```

```json
{
  "1c086812-09c0-4c16-811c-85da2c903016-S0": [
    {
      "Name": "swan",
      "Id": "399ff028420252a2050dcd951f68ecc9a8f45ea8e27844d156786490d0126644",
      "Scope": "local",
      "Driver": "macvlan",
      "EnableIPv6": false,
      "IPAM": {
        "Driver": "swan",
        "Options": {
          
        },
        "Config": [
          {
            "Subnet": "192.168.1.0/24",
            "Gateway": "192.168.1.1"
          }
        ]
      },
      "Internal": false,
      "Attachable": false,
      "Containers": {
        
      },
      "Options": {
        "parent": "enp0s3"
      },
      "Labels": {
        
      }
    },
    {
      "Name": "bridge",
      "Id": "cc8153b8865ff8985708943b2075a55e8038512b1b86c6bba4fdfbbcec1ee67d",
      "Scope": "local",
      "Driver": "bridge",
      "EnableIPv6": false,
      "IPAM": {
        "Driver": "default",
        "Options": null,
        "Config": [
          {
            "Subnet": "172.17.0.0/16",
            "Gateway": "172.17.0.1"
          }
        ]
      },
      "Internal": false,
      "Attachable": false,
      "Containers": {
        
      },
      "Options": {
        "com.docker.network.bridge.default_bridge": "true",
        "com.docker.network.bridge.enable_icc": "true",
        "com.docker.network.bridge.enable_ip_masquerade": "true",
        "com.docker.network.bridge.host_binding_ipv4": "0.0.0.0",
        "com.docker.network.bridge.name": "docker0",
        "com.docker.network.driver.mtu": "1500"
      },
      "Labels": {
        
      }
    },
    {
      "Name": "host",
      "Id": "fce292794513294b4f8244f9133c584852961f4b8ec74130837e8bb8e619bb1b",
      "Scope": "local",
      "Driver": "host",
      "EnableIPv6": false,
      "IPAM": {
        "Driver": "default",
        "Options": null,
        "Config": [
          
        ]
      },
      "Internal": false,
      "Attachable": false,
      "Containers": {
        "1c4b4b05852d9b04ac100f8c6f126c6a932d56004426d350cec7e7496432c855": {
          "Name": "mesos-slave",
          "EndpointID": "811ec0cbe973ad2cee774a7ca382098f967cc3616066e7f4a5c236d132fe9c0b",
          "MacAddress": "",
          "IPv4Address": "",
          "IPv6Address": ""
        },
        "22193beac2d0c322c242d4e0344ba25b352947e675c0a2122bd9e04591c16f46": {
          "Name": "swan-manager-2",
          "EndpointID": "878aed535b0ff2763f08420e69b64d50b47c26d5f38375e56226b49af70f0484",
          "MacAddress": "",
          "IPv4Address": "",
          "IPv6Address": ""
        },
        "4d79ef21dc164107d50f4396e7fa560be6c771e5b0601d0ce7cb7c9085ee4534": {
          "Name": "etcd",
          "EndpointID": "4e546c99cd7add3826ce995480c26743787c645f641d09e5db156d4223748ef8",
          "MacAddress": "",
          "IPv4Address": "",
          "IPv6Address": ""
        },
        "77ebef1b711eb278134f93ca06a41a155139f0ee0e7aa1588c7650f7f30fee1a": {
          "Name": "swan-agent",
          "EndpointID": "37b37dfbe9da7d39d70f754ff7ba51f3de60223c76d693f3698885f0b90dcb6f",
          "MacAddress": "",
          "IPv4Address": "",
          "IPv6Address": ""
        },
        "8a44de714be1331f6ad351cb3ee0db30fa78a3efd7bd7d67f3a0caa0d211b85a": {
          "Name": "mesos-master",
          "EndpointID": "7a83118125735271ad987c8e71292f8c32ec789b6892315a28abd4cb737786ca",
          "MacAddress": "",
          "IPv4Address": "",
          "IPv6Address": ""
        },
        "91b983e66a6f3d6593af106f5db00c2419da2cf8d1d75d37e59465d8f45d0632": {
          "Name": "zkui",
          "EndpointID": "1308daaf9b9dc9f4ef8fc930cf8affe2a4798b942d162091a99b90cc1cc3311d",
          "MacAddress": "",
          "IPv4Address": "",
          "IPv6Address": ""
        },
        "a58aa0c61463805a37b829cc9acf9d1ae6bc30ad1f075841feab15f0e12ec544": {
          "Name": "swan-manager-1",
          "EndpointID": "29b7e6f00cc936cc5a4a8838650b0075ac1c78bc9e3cae1af26d3bf351553066",
          "MacAddress": "",
          "IPv4Address": "",
          "IPv6Address": ""
        },
        "ddfa3c29c0fa2550fd9d2c7885798b52784504e071dba30688e5383d26f1e230": {
          "Name": "zookeeper",
          "EndpointID": "bc4e0c83101d70045ee5c4cf0511f5d8d9c7f07816c49c9a6d3e9c971242c411",
          "MacAddress": "",
          "IPv4Address": "",
          "IPv6Address": ""
        },
        "f3ca0793745d7f10cb28b3f3f6b44c775c65aed591d4d167371335eba9cd8204": {
          "Name": "swan-manager-3",
          "EndpointID": "e99b89442567a312fba1e0f22be6203f131896ac8ff89ebeb6f4122bde9b3d16",
          "MacAddress": "",
          "IPv4Address": "",
          "IPv6Address": ""
        }
      },
      "Options": {
        
      },
      "Labels": {
        
      }
    },
    {
      "Name": "none",
      "Id": "32c5e63659f780200bdc170f294839c694b112fcf6835c0f29a484ca49c13544",
      "Scope": "local",
      "Driver": "null",
      "EnableIPv6": false,
      "IPAM": {
        "Driver": "default",
        "Options": null,
        "Config": [
          
        ]
      },
      "Internal": false,
      "Attachable": false,
      "Containers": {
        
      },
      "Options": {
        
      },
      "Labels": {
        
      }
    }
  ],
  "212c92eb-f594-43d5-89da-7820a56e8570-S1": [
    {
      "Name": "none",
      "Id": "351fc1de17d0a8105c5414e3443cd304028e897cd655744f9885380a7c2521c9",
      "Scope": "local",
      "Driver": "null",
      "EnableIPv6": false,
      "IPAM": {
        "Driver": "default",
        "Options": null,
        "Config": [
          
        ]
      },
      "Internal": false,
      "Attachable": false,
      "Containers": {
        
      },
      "Options": {
        
      },
      "Labels": {
        
      }
    },
    {
      "Name": "host",
      "Id": "856419edd231111b462fe8eb6bcc34f455fff1c0cee48eb9e6b8c2add47219fd",
      "Scope": "local",
      "Driver": "host",
      "EnableIPv6": false,
      "IPAM": {
        "Driver": "default",
        "Options": null,
        "Config": [
          
        ]
      },
      "Internal": false,
      "Attachable": false,
      "Containers": {
        "2f1a874daac54a39b03acc907f6f9c58e4848aca86803f70da1109d267a8f254": {
          "Name": "swan-agent",
          "EndpointID": "ae78e4416c153958cbc21459a420bbd58198463fc2fb0bcc33b176143778632b",
          "MacAddress": "",
          "IPv4Address": "",
          "IPv6Address": ""
        },
        "d26b7f6e9937586d072e328b90b328567c32f357b97d3057376e12835743626e": {
          "Name": "mesos-slave",
          "EndpointID": "b4c7e8d69922b7b7bfbbad3031efae3e439f2ad515a1e022b9600b6d8129f597",
          "MacAddress": "",
          "IPv4Address": "",
          "IPv6Address": ""
        }
      },
      "Options": {
        
      },
      "Labels": {
        
      }
    },
    {
      "Name": "swan",
      "Id": "901cf0b20c96fef77631bd7af46f8a8d63286c3b9e672e2ba7c2bd9412cda842",
      "Scope": "local",
      "Driver": "macvlan",
      "EnableIPv6": false,
      "IPAM": {
        "Driver": "swan",
        "Options": {
          
        },
        "Config": [
          {
            "Subnet": "192.168.1.0/24",
            "Gateway": "192.168.1.1"
          }
        ]
      },
      "Internal": false,
      "Attachable": false,
      "Containers": {
        
      },
      "Options": {
        "parent": "enp0s3"
      },
      "Labels": {
        
      }
    },
    {
      "Name": "bridge",
      "Id": "e4bf109efd8440fd6a077cfea8486b70b9355471d728ab4a74e5b4cb2ebee6cf",
      "Scope": "local",
      "Driver": "bridge",
      "EnableIPv6": false,
      "IPAM": {
        "Driver": "default",
        "Options": null,
        "Config": [
          {
            "Subnet": "172.17.0.0/16",
            "Gateway": "172.17.0.1"
          }
        ]
      },
      "Internal": false,
      "Attachable": false,
      "Containers": {
        
      },
      "Options": {
        "com.docker.network.bridge.default_bridge": "true",
        "com.docker.network.bridge.enable_icc": "true",
        "com.docker.network.bridge.enable_ip_masquerade": "true",
        "com.docker.network.bridge.host_binding_ipv4": "0.0.0.0",
        "com.docker.network.bridge.name": "docker0",
        "com.docker.network.driver.mtu": "1500"
      },
      "Labels": {
        
      }
    }
  ]
}
```

#### framework
```
GET /v1/framework
```

```json
{
	"ID": "fe9f9429-e17c-4aad-9689-3ba8f5a11e30-0000"
}
```

#### Ping
```
GET /ping
```
Example request:
```
GET /ping
```
Example response:
```
"pong"
```

#### Version
```
GET /version
```

Example response:
```
{
  "version": "v0.1.11",
  "commit": "v0.1.11-153-g397f477-dirty",
  "build_time": "2017-05-27:11-27-19",
  "go_version": "go1.8.1",
  "os": "linux",
  "arch": "amd64"
}
```

#### Leader
```
GET /v1/leader
```

Example response:
```
{
    "leader": "192.168.1.92:5016"
}
```

#### take over agent docker remote api
swan上可以直接请求 节点上的 [Docker Remote API 1.21](https://docs.docker.com/engine/api/v1.21/)  
而不需要节点上的docker 监听tcp端口, 不过需要swan agent部署的时候挂载/var/run/docker.sock:/var/run/docker.sock:rw  

Example:
```
/v1/agents/212c92eb-f594-43d5-89da-7820a56e8570-S0/docker/containers/json

/v1/agents/212c92eb-f594-43d5-89da-7820a56e8570-S0/docker/containers/54b701f325fe1f229dd9174fb90123057e933d2f564d4e5f90e0a69ee6461770/json

/v1/agents/212c92eb-f594-43d5-89da-7820a56e8570-S0/docker/images/json

/v1/agents/212c92eb-f594-43d5-89da-7820a56e8570-S0/docker/networks
```

#### Reset
`Reset` is used for manually update app's op-status to noop in some situation so that you can continue.
```
POST /v1/apps/{app_id}/reset
```

Example request:
```
POST /v1/apps/nginx004.default.testuser.dataman/reset
```

Example response:
```
{
    "previous": "scaling",
    "current": "noop"
}
```

#### list agents
```
GET /v1/agents             // list normal agents
GET /v1/agents?debug=true  // list all of agents including offlines
```

```json
{
  "212c92eb-f594-43d5-89da-7820a56e8570-S0": {
    "hostname": "master",
    "os": "NAME=\"CentOS Linux\"\nVERSION=\"7 (Core)\"\nID=\"centos\"\nID_LIKE=\"rhel fedora\"\nVERSION_ID=\"7\"\nPRETTY_NAME=\"CentOS Linux 7 (Core)\"\nANSI_COLOR=\"0;31\"\nCPE_NAME=\"cpe:/o:centos:centos:7\"\nHOME_URL=\"https://www.centos.org/\"\nBUG_REPORT_URL=\"https://bugs.centos.org/\"\n\nCENTOS_MANTISBT_PROJECT=\"CentOS-7\"\nCENTOS_MANTISBT_PROJECT_VERSION=\"7\"\nREDHAT_SUPPORT_PRODUCT=\"centos\"\nREDHAT_SUPPORT_PRODUCT_VERSION=\"7\"\n\n",
    "uptime": "34909.000000",
    "unixtime": 1502193274,
    "loadavg": 0.31,
    "cpu": {
      "processor": 4,
      "physical": 4,
      "used": 2.174071913836869
    },
    "memory": {
      "total": 3975307264,
      "used": 1428484096
    },
    "containers": {
      "total": 0,
      "running": 0,
      "stopped": 0,
      "killed": 0,
      "paused": 0
    },
    "ips": {
      "docker0": [
        "172.17.0.1"
      ],
      "enp0s3": [
        "192.168.1.117",
        "192.168.1.196"
      ]
    },
    "listenings": [
      2379,
      2380,
      9999,
      80,
      10000,
      10001,
      57489,
      47220,
      22,
      52921,
      5050,
      5051,
      34237,
      514,
      60326,
      9900,
      22,
      31000,
      31001,
      443,
      31004,
      31005,
      31006,
      31007,
      31008,
      31009,
      31010,
      9090,
      514,
      31011,
      38883,
      2181
    ]
  },
  "212c92eb-f594-43d5-89da-7820a56e8570-S1": {
    "hostname": "node1",
    "os": "NAME=\"CentOS Linux\"\nVERSION=\"7 (Core)\"\nID=\"centos\"\nID_LIKE=\"rhel fedora\"\nVERSION_ID=\"7\"\nPRETTY_NAME=\"CentOS Linux 7 (Core)\"\nANSI_COLOR=\"0;31\"\nCPE_NAME=\"cpe:/o:centos:centos:7\"\nHOME_URL=\"https://www.centos.org/\"\nBUG_REPORT_URL=\"https://bugs.centos.org/\"\n\nCENTOS_MANTISBT_PROJECT=\"CentOS-7\"\nCENTOS_MANTISBT_PROJECT_VERSION=\"7\"\nREDHAT_SUPPORT_PRODUCT=\"centos\"\nREDHAT_SUPPORT_PRODUCT_VERSION=\"7\"\n\n",
    "uptime": "34906.000000",
    "unixtime": 1502193275,
    "loadavg": 0,
    "cpu": {
      "processor": 4,
      "physical": 4,
      "used": 0.11925134971008333
    },
    "memory": {
      "total": 3975307264,
      "used": 186720256
    },
    "containers": {
      "total": 0,
      "running": 0,
      "stopped": 0,
      "killed": 0,
      "paused": 0
    },
    "ips": {
      "docker0": [
        "172.17.0.1"
      ],
      "enp0s3": [
        "192.168.1.130"
      ]
    },
    "listenings": [
      80,
      22,
      5051,
      9900,
      22,
      443
    ]
  }
}
```

#### query agent id
```
GET /v1/agents/query_id?ips=192.168.1.196,192.168.1.130,xxx
```

```plain
212c92eb-f594-43d5-89da-7820a56e8570-S0
```

#### get agent
```
GET /v1/agents/{agent_id}
```

```json
{
  "hostname": "master",
  "os": "NAME=\"CentOS Linux\"\nVERSION=\"7 (Core)\"\nID=\"centos\"\nID_LIKE=\"rhel fedora\"\nVERSION_ID=\"7\"\nPRETTY_NAME=\"CentOS Linux 7 (Core)\"\nANSI_COLOR=\"0;31\"\nCPE_NAME=\"cpe:/o:centos:centos:7\"\nHOME_URL=\"https://www.centos.org/\"\nBUG_REPORT_URL=\"https://bugs.centos.org/\"\n\nCENTOS_MANTISBT_PROJECT=\"CentOS-7\"\nCENTOS_MANTISBT_PROJECT_VERSION=\"7\"\nREDHAT_SUPPORT_PRODUCT=\"centos\"\nREDHAT_SUPPORT_PRODUCT_VERSION=\"7\"\n\n",
  "uptime": "25042.000000",
  "unixtime": 1502877459,
  "loadavg": 0.11,
  "cpu": {
    "processor": 4,
    "physical": 4,
    "used": 1.026625907152237
  },
  "memory": {
    "total": 3975307264,
    "used": 1297035264
  },
  "containers": {
    "total": 0,
    "running": 0,
    "stopped": 0,
    "killed": 0,
    "paused": 0
  },
  "ips": {
    "docker0": [
      "172.17.0.1"
    ],
    "enp0s3": [
      "192.168.1.117",
      "192.168.1.196"
    ]
  },
  "listenings": [
    2379,
    2380,
    9999,
    80,
    10000,
    10001,
    22,
    5050,
    5051,
    514,
    9900,
    22,
    443,
    35612,
    9090,
    514,
    99,
    2181
  ]
}
```

#### close agent
```
DELETE /v1/agents/{agent_id}
```

#### get agent dns
```
GET /v1/agents/{agent_id}/dns
```

```json
{
  "PROXY": [
    {
      "id": "local_proxy",
      "parent": "PROXY",
      "ip": "192.168.1.196",
      "port": "80",
      "weight": 0,
      "proxy_record": true,
      "clean_name": ""
    }
  ],
  "demo.default.bbk.dataman-mesos": [
    {
      "id": "be0410cb4573.3.demo.default.bbk.dataman-mesos",
      "parent": "demo.default.bbk.dataman-mesos",
      "ip": "192.168.1.130",
      "port": "31006",
      "weight": 100,
      "proxy_record": false,
      "clean_name": "3.demo.default.bbk.dataman-mesos.bbklab.net."
    },
    {
      "id": "93af82d0ef07.2.demo.default.bbk.dataman-mesos",
      "parent": "demo.default.bbk.dataman-mesos",
      "ip": "192.168.1.130",
      "port": "31004",
      "weight": 100,
      "proxy_record": false,
      "clean_name": "2.demo.default.bbk.dataman-mesos.bbklab.net."
    },
    {
      "id": "0c97e734c2f4.4.demo.default.bbk.dataman-mesos",
      "parent": "demo.default.bbk.dataman-mesos",
      "ip": "192.168.1.130",
      "port": "31008",
      "weight": 100,
      "proxy_record": false,
      "clean_name": "4.demo.default.bbk.dataman-mesos.bbklab.net."
    },
    {
      "id": "94be53aa2c26.1.demo.default.bbk.dataman-mesos",
      "parent": "demo.default.bbk.dataman-mesos",
      "ip": "192.168.1.130",
      "port": "31002",
      "weight": 100,
      "proxy_record": false,
      "clean_name": "1.demo.default.bbk.dataman-mesos.bbklab.net."
    },
    {
      "id": "b8d3d1d49fe9.0.demo.default.bbk.dataman-mesos",
      "parent": "demo.default.bbk.dataman-mesos",
      "ip": "192.168.1.130",
      "port": "31000",
      "weight": 100,
      "proxy_record": false,
      "clean_name": "0.demo.default.bbk.dataman-mesos.bbklab.net."
    }
  ],
  "static.default.bbk.dataman-mesos": [
    {
      "id": "83182b005e01.0.static.default.bbk.dataman-mesos",
      "parent": "static.default.bbk.dataman-mesos",
      "ip": "192.168.1.199",
      "port": "0",
      "weight": 100,
      "proxy_record": false,
      "clean_name": "0.static.default.bbk.dataman-mesos.bbklab.net."
    },
    {
      "id": "881454606c0d.1.static.default.bbk.dataman-mesos",
      "parent": "static.default.bbk.dataman-mesos",
      "ip": "192.168.1.200",
      "port": "0",
      "weight": 100,
      "proxy_record": false,
      "clean_name": "1.static.default.bbk.dataman-mesos.bbklab.net."
    }
  ]
}
```

#### get agent dns stats
```
GET /v1/agents/{agent_id}/dns/stats
```

```json
{
  "global": {
    "requests": 19,
    "fails": 0,
    "authority": 18,
    "forward": 1,
    "type_a": 18,
    "type_srv": 0
  },
  "parents": {
    "demo.default.bbk.dataman-mesos": {
      "requests": 8,
      "fails": 0,
      "authority": 8,
      "forward": 0,
      "type_a": 8,
      "type_srv": 0
    }
  },
  "uptime": "1h58m19.683359556s"
}
```

#### get agent proxy
```
GET /v1/agents/{agent_id}/proxy
```

```json
[
  {
    "name": "demo.default.bbk.dataman-mesos",
    "alias": "g.cn",
    "listen": ":99",
    "sticky": false,
    "backends": [
      {
        "id": "be0410cb4573.3.demo.default.bbk.dataman-mesos",
        "ip": "192.168.1.130",
        "port": 31006,
        "scheme": "",
        "version": "",
        "weihgt": 100,
        "clean_name": "3.demo.default.bbk.dataman-mesos"
      },
      {
        "id": "93af82d0ef07.2.demo.default.bbk.dataman-mesos",
        "ip": "192.168.1.130",
        "port": 31004,
        "scheme": "",
        "version": "",
        "weihgt": 100,
        "clean_name": "2.demo.default.bbk.dataman-mesos"
      },
      {
        "id": "0c97e734c2f4.4.demo.default.bbk.dataman-mesos",
        "ip": "192.168.1.130",
        "port": 31008,
        "scheme": "",
        "version": "",
        "weihgt": 100,
        "clean_name": "4.demo.default.bbk.dataman-mesos"
      },
      {
        "id": "94be53aa2c26.1.demo.default.bbk.dataman-mesos",
        "ip": "192.168.1.130",
        "port": 31002,
        "scheme": "",
        "version": "",
        "weihgt": 100,
        "clean_name": "1.demo.default.bbk.dataman-mesos"
      },
      {
        "id": "b8d3d1d49fe9.0.demo.default.bbk.dataman-mesos",
        "ip": "192.168.1.130",
        "port": 31000,
        "scheme": "",
        "version": "",
        "weihgt": 100,
        "clean_name": "0.demo.default.bbk.dataman-mesos"
      }
    ]
  }
]
```

#### get agent proxy stats
```
GET /v1/agents/{agent_id}/proxy/stats
```

```json
{
  "counter": {
    "global": {
      "rx_bytes": 1200,
      "tx_bytes": 12765,
      "requests": 15,
      "fails": 0,
      "rx_rate": 0,
      "tx_rate": 0,
      "requests_rate": 0,
      "fails_rate": 0,
      "uptime": "2h0m32.073628892s"
    },
    "upstream": {
      "demo.default.bbk.dataman-mesos": {
        "0c97e734c2f4.4.demo.default.bbk.dataman-mesos": {
          "active_clients": 0,
          "rx_bytes": 480,
          "tx_bytes": 5106,
          "requests": 6,
          "rx_rate": 0,
          "tx_rate": 0,
          "requests_rate": 0,
          "uptime": "53.861101038s"
        },
        "93af82d0ef07.2.demo.default.bbk.dataman-mesos": {
          "active_clients": 0,
          "rx_bytes": 80,
          "tx_bytes": 851,
          "requests": 1,
          "rx_rate": 0,
          "tx_rate": 0,
          "requests_rate": 0,
          "uptime": "36.96047377s"
        },
        "94be53aa2c26.1.demo.default.bbk.dataman-mesos": {
          "active_clients": 0,
          "rx_bytes": 160,
          "tx_bytes": 1702,
          "requests": 2,
          "rx_rate": 0,
          "tx_rate": 0,
          "requests_rate": 0,
          "uptime": "35.760768071s"
        },
        "b8d3d1d49fe9.0.demo.default.bbk.dataman-mesos": {
          "active_clients": 0,
          "rx_bytes": 80,
          "tx_bytes": 851,
          "requests": 1,
          "rx_rate": 0,
          "tx_rate": 0,
          "requests_rate": 0,
          "uptime": "36.148380693s"
        },
        "be0410cb4573.3.demo.default.bbk.dataman-mesos": {
          "active_clients": 0,
          "rx_bytes": 400,
          "tx_bytes": 4255,
          "requests": 5,
          "rx_rate": 0,
          "tx_rate": 0,
          "requests_rate": 0,
          "uptime": "52.731013333s"
        }
      }
    }
  },
  "httpd": "192.168.1.130:80",
  "httpdTLS": ":443",
  "tcpd": {
    ":99": {
      "active_clients": 0,
      "listen": ":99",
      "serving": true,
      "uptime": "11m55.39149628s"
    }
  }
}
```
