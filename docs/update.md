#### Update
##### API

```
PUT /v1/apps/{app_id}
```
Example request:
```
PUT /v1/apps/nginx0r2.default.xcm.dataman HTTP/1.1
Content-Type: application/json
{
    "instances": 5, 
    "canary": {
        "enabled": false,
        "value": 0.1,
    }
}
```
Json parameters:
```
instances: the number of tasks to be updated.
canary:
    enabled: Disable or Enabled canary publish(gray publish). If Disabled, value is ignore. Default is Disabled.
    value:   Percentage of traffic for new version. 0.1 means ten percent.
```
Example response:
```
HTTP/1.1 202 Accepted
```

#### How

First, create a new version for the app you plan to update.

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
  "priority": 100,
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
Second, process updating.

```
During the update process, swan will be automatic find the new version you created previous step, and updated the old tasks to the new version
according you strategy.

If canary is enabled, the traffic control will be enabled. you can control how much traffic the new version is divided.
```





