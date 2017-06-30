##### Health Check

Spec

```
HTTP

    {
      "protocol": "http",
      "path": "/",
      "portName": "web",
      "delaySeconds": 2,
      "gracePeriodSeconds": 5,
      "intervalSeconds": 1,
      "timeoutSeconds": 1,
      "consecutiveFailures": 5
    }
```

````
TCP 
    {
      "protocol": "tcp",
      "portName": "web",
      "delaySeconds": 2,
      "gracePeriodSeconds": 5,
      "intervalSeconds": 1,
      "timeoutSeconds": 1,
      "consecutiveFailures": 5
    }
```

```
CMD
    {
      "protocol": "cmd",
      "command": "",
      "delaySeconds": 2,
      "gracePeriodSeconds": 5,
      "intervalSeconds": 1,
      "timeoutSeconds": 1,
      "consecutiveFailures": 5
    }
```
