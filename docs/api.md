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
  - [compose](https://github.com/Dataman-Cloud/swan/tree/master/docs/compose.md)

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
  - [GET /v1/agents](#agents)
  - [GET /v1/agents/{agent_id}]
  - [DELETE /v1/agents/{agent_id}]

+ reset 
  - [POST /v1/apps/{app_id}/reset](#reset)

+ [deploy policy](https://github.com/Dataman-Cloud/swan/tree/master/docs/deploy.md)

+ [constraints](https://github.com/Dataman-Cloud/swan/tree/master/docs/constraints.md)

+ [scale](https://github.com/Dataman-Cloud/swan/tree/master/docs/scale.md)
 
+ [update policy](https://github.com/Dataman-Cloud/swan/tree/master/docs/update.md)

+ [port mapping](https://github.com/Dataman-Cloud/swan/tree/master/docs/port-mapping.md)
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
    "id": "nginx0r1.default.xcm.dataman",
    "name": "nginx0r1",
    "alias": "",
    "runAs": "xcm",
    "cluster": "dataman",
    "operationStatus": "noop",
    "tasks": [
      {
        "id": "e6404f0324d2.0.nginx0r1.default.xcm.dataman",
        "name": "0.nginx0r1.default.xcm.dataman",
        "ip": "192.168.1.102",
        "port": 31008,
        "healthy": "unset",
        "weight": 100,
        "agentId": "7a40294e-b16b-4ac3-bbe4-1865df4a4705-S6",
        "version": "1498029948754163146",
        "status": "TASK_RUNNING",
        "errmsg": ""
        "created": "2017-06-21T15:25:48.78944685+08:00",
        "updated": "2017-06-21T15:25:48.78944688+08:00"
      }
    ],
    "currentVersion": [
      "1498029948754163146"
    ],
    "versions": [
      {
        "id": "1498029948754163146",
        "name": "nginx0r1",
        "cmd": "",
        "cpus": 0.01,
        "mem": 32,
        "disk": 0,
        "instances": 1,
        "runAs": "xcm",
        "container": {
          "type": "DOCKER",
          "docker": {
            "image": "nginx",
            "network": "bridge",
            "parameters": [
              {
                "key": "label",
                "value": "APP_ID=wordpress"
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
          "volumes": [
            {
              "containerPath": "/data",
              "hostPath": "/home",
              "mode": "RW"
            }
          ]
        },
        "labels": null,
        "healthCheck": null,
        "env": {},
        "killPolicy": null,
        "updatPolicy": null,
        "constraints": [],
        "uris": [],
        "ips": null,
        "proxy": {
          "enabled": false,
          "alias": ""
        }
      }
    ],
    "status": "available",
    "health": {
      "healthy": 0,
      "unhealthy": 0,
      "unset": 1
    },
    "created": "2017-06-21T15:25:48.754164732+08:00",
    "updated": "2017-06-21T15:25:48.754164753+08:00"
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
            "alias": ""
  }
}
```
Example response:
```
  HTTP/1.1 201 Created
  Content-Type: application/json

  {
       "Id":"nginx0r1.default.xcm.dataman"
  }
```
#### Inspect a app
```
GET /v1/apps/{app_id}
```
Example request:
``
GET /v1/apps/nginx0r2.default.xcm.dataman
```
Example response:
```json
HTTP/1.1 200 OK
Content-Type: application/json

{
    "id":"nginx0r2.default.xcm.dataman",
    "name":"nginx0r2",
    "alias":"",
    "runAs":"xcm",
    "cluster":"dataman",
    "operationStatus":"noop",
    "tasks":[
        {
            "id":"731ea4512976.0.nginx0r2.default.xcm.dataman",
            "name":"0.nginx0r2.default.xcm.dataman",
            "ip":"192.168.1.102",
            "port":31010,
            "healthy":"unset",
            "weight":100,
            "agentId":"7a40294e-b16b-4ac3-bbe4-1865df4a4705-S6",
            "version":"1498030396211326306",
            "status":"TASK_RUNNING",
            "errmsg":"",
            "created":"2017-06-21T15:33:16.238348516+08:00",
            "updated":"2017-06-21T15:33:16.238348626+08:00"
        }
    ],
    "currentVersion":[
        "1498030396211326306"
    ],
    "versions":[
        {
            "id":"1498030396211326306",
            "name":"nginx0r2",
            "cmd":"",
            "cpus":0.01,
            "mem":32,
            "disk":0,
            "instances":1,
            "runAs":"xcm",
            "container":{
                "type":"DOCKER",
                "docker":{
                    "image":"nginx",
                    "network":"bridge",
                    "parameters":[
                        {
                            "key":"label",
                            "value":"APP_ID=wordpress"
                        }
                    ],
                    "portMappings":[
                        {
                            "containerPort":80,
                            "hostPort":80,
                            "name":"web",
                            "protocol":"tcp"
                        }
                    ],
                    "privileged":true
                },
                "volumes":[
                    {
                        "containerPath":"/data",
                        "hostPath":"/home",
                        "mode":"RW"
                    }
                ]
            },
            "labels":null,
            "healthCheck":null,
            "env":{

            },
            "killPolicy":null,
            "updatPolicy":null,
            "constraints":[

            ],
            "uris":[

            ],
            "ips":null,
            "proxy":{
                "enabled":false,
                "alias":""
				"listen": 99,
				"sticky": false
            }
        }
    ],
    "status":"available",
    "health":{
        "healthy":0,
        "unhealthy":0,
        "unset":1
    },
    "created":"2017-06-21T15:33:16.211327705+08:00",
    "updated":"2017-06-21T15:33:16.211327722+08:00"
}
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
ips(optional) : ip list for static ip(brige or host or scale down ignore)
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
			"listen": 99,
			"sticky": false
  },
  "update": {
      "delay": 5,
      "onfailure": "continue"
  }
}
```
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
    "id": "e6404f0324d2.0.nginx0r1.default.xcm.dataman",
    "name": "0.nginx0r1.default.xcm.dataman",
    "ip": "192.168.1.102",
    "port": 31008,
    "healthy": "unset",
    "weight": 100,
    "agentId": "7a40294e-b16b-4ac3-bbe4-1865df4a4705-S6",
    "version": "1498029948754163146",
    "status": "TASK_RUNNING",
    "errmsg": "",
    "container_id": "5dc2ae2bb5901c0f7aa5a24ffdba34166fb3f7730d88a93021c019c43c194b4d",
    "container_name": "/mesos-77cd3fe3-ead4-42e4-aff2-6b77f3697b1c-S0.088f44db-11d1-407a-9649-2811bf1b0d69",
    "created": "2017-06-21T15:25:48.78944685+08:00",
    "updated": "2017-06-21T15:25:48.78944688+08:00"
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
    "id": "1498029948754163146",
    "name": "nginx0r1",
    "cmd": "",
    "cpus": 0.01,
    "mem": 32,
    "disk": 0,
    "instances": 1,
    "runAs": "xcm",
    "container": {
      "type": "DOCKER",
      "docker": {
        "image": "nginx",
        "network": "bridge",
        "parameters": [
          {
            "key": "label",
            "value": "APP_ID=wordpress"
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
      "volumes": [
        {
          "containerPath": "/data",
          "hostPath": "/home",
          "mode": "RW"
        }
      ]
    },
    "labels": null,
    "healthCheck": null,
    "env": {},
    "killPolicy": null,
    "updatPolicy": null,
    "constraints": [],
    "uris": [],
    "ips": null,
    "proxy": {
      "enabled": false,
      "alias": ""
    }
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
  "id": "e6404f0324d2.0.nginx0r1.default.xcm.dataman",
  "name": "0.nginx0r1.default.xcm.dataman",
  "ip": "192.168.1.102",
  "port": 31008,
  "healthy": "unset",
  "weight": 100,
  "agentId": "7a40294e-b16b-4ac3-bbe4-1865df4a4705-S6",
  "version": "1498029948754163146",
  "status": "TASK_RUNNING",
  "errmsg": "",
  "container_id": "5dc2ae2bb5901c0f7aa5a24ffdba34166fb3f7730d88a93021c019c43c194b4d",
  "container_name": "/mesos-77cd3fe3-ead4-42e4-aff2-6b77f3697b1c-S0.088f44db-11d1-407a-9649-2811bf1b0d69",
  "created": "2017-06-21T15:25:48.78944685+08:00",
  "updated": "2017-06-21T15:25:48.78944688+08:00"
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
  "id": "1498029948754163146",
  "name": "nginx0r1",
  "cmd": "",
  "cpus": 0.01,
  "mem": 32,
  "disk": 0,
  "instances": 1,
  "runAs": "xcm",
  "container": {
    "type": "DOCKER",
    "docker": {
      "image": "nginx",
      "network": "bridge",
      "parameters": [
        {
          "key": "label",
          "value": "APP_ID=wordpress"
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
    "volumes": [
      {
        "containerPath": "/data",
        "hostPath": "/home",
        "mode": "RW"
      }
    ]
  },
  "labels": null,
  "healthCheck": null,
  "env": {},
  "killPolicy": null,
  "updatPolicy": null,
  "constraints": [],
  "uris": [],
  "ips": null,
  "proxy": {
    "enabled": false,
    "alias": ""
  }
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
      "Id":"1498791358276219465"
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
     "Value": 0.2
 }
```
Json parameters:
```
Value: Percentage of traffic. Indicated how much traffics switched to new version.
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


### agents

```
GET /v1/agents
GET /v1/agents/{agent_id}
```

```json
{
  "3264208446845635": {
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
  "3981314045636649": {
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
