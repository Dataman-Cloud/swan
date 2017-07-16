#### Json Spec

```
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
  "restart": {
      "attempts": 3,
      "delay": 1,
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

