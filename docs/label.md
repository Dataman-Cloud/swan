+ labels 
  - [GET /v1/mesos/agents](#list-all-mesos-agents) *List all mesos agents*
  - [POST /v1/mesos/agents/{agent_ip}](#create-new-label-on-mesos-agent) *Create new Label*
  - [GET /v1/mesos/agents/{agent_ip}](#inspect-a-mesos-agent) *Inspect a mesos agent*
  - [PATCH /v1/mesos/agents/{agent_ip}](#update-agent-label) *Update agent label*
  - [DELETE /v1/mesos/agents/{agent_ip}](#delete-mesos-agent-label) *Delete mesos agent label*

#### List all mesos agents 
```
GET /v1/mesos/agents
```
Example request:
```
GET /v1/mesos/agents
```
Example response:
```
[
    {
        "attrs": {
            "arch": "64",
            "disk": "ssd"
        },
        "cpus": 4,
        "disk": 39910,
        "gpus": 0,
        "id": "07239dfa-2982-4046-8f1c-f8f3c14d4649-S4",
        "ip": "172.16.1.18",
        "mem": 6799,
        "ports": 1001
    },
    {
        "attrs": {},
        "cpus": 4,
        "disk": 39910,
        "gpus": 0,
        "id": "07239dfa-2982-4046-8f1c-f8f3c14d4649-S5",
        "ip": "172.16.1.17",
        "mem": 6799,
        "ports": 1001
    }
]
```

#### Create new label on mesos agent
```
POST /v1/mesos/agents/{agent_ip}
```

Example request:
```
POST /v1/vclusters HTTP/1.1
Content-Type: application/json
{
        "key": "disk",
        "value": "ssd"
}
```

Example response:
```
HTTP/1.1 201 Created
Content-Type: application/json

{
    "attrs": {
        "disk": "ssd"
    },
    "cpus": 4,
    "disk": 39910,
    "gpus": 0,
    "id": "07239dfa-2982-4046-8f1c-f8f3c14d4649-S4",
    "ip": "172.16.1.18",
    "mem": 6799,
    "ports": 1001
}
```

#### Inspect a mesos agent 
```
GET /v1/mesos/agents/{agent_ip}
```
Example request:
```
GET /v1/mesos/agents/172.16.1.18 HTTP/1.1
```
Example response:
```
HTTP/1.1 200 OK
Content-Type: application/json

{
    "attrs": {
        "disk": "ssd"
    },
    "cpus": 4,
    "disk": 39910,
    "gpus": 0,
    "id": "07239dfa-2982-4046-8f1c-f8f3c14d4649-S4",
    "ip": "172.16.1.18",
    "mem": 6799,
    "ports": 1001
}
```

#### update agent label 
```
PATCH /v1/mesos/agents/{agent_ip}
```
Example request:
```
PATCH /v1/mesos/agents/172.16.1.18 HTTP/1.1
Content-Type: application/json
{
    "key": "disk",
    "value": "hdd"
}

```
Example response:
```
HTTP/1.1 200 OK
Content-Type: application/json

{
    "attrs": {
        "disk": "ssd1"
    },
    "cpus": 4,
    "disk": 39910,
    "gpus": 0,
    "id": "07239dfa-2982-4046-8f1c-f8f3c14d4649-S4",
    "ip": "172.16.1.18",
    "mem": 6799,
    "ports": 1001
}
```


#### Delete mesos agent label 
```
DELETE /v1/mesos/agents/{agent_ip}
```
Example request:
```
DELETE /v1/mesos/agents/{agent_ip} HTTP/1.1
Content-Type: application/json

{
    "key": "disk",
    "value": "",
}
```
Example response:
```
HTTP/1.1 204 No Content 
```
