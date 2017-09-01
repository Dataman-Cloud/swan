+ vclusters
  - [GET /v1/vclusters](#list-all-vclusters) *List all vclusters*
  - [POST /v1/vclusters](#create-new-vcluster) *Create new vcluster*
  - [GET /v1/vclusters/{vcluster_name}](#inspect-a-cluster) *Inspect a vcluster*
  - [DELETE /v1/vclusters/{vcluster_name}](#delete-a-vcluster) *Delete a vcluster*
  - [POST /v1/vclusters/{vcluster_name}/nodes](#add-node) *Add host to vcluster*
  - [PATCH /v1/vclusters/{vcluster_name}/nodes/{node_ip}](#update-node-attribute) *Update node attribute*
  - [GET /v1/vclusters/{vcluster_name}/nodes](#list-all-nodes) *List all nodes*
  - [GET /v1/vclusters/{vcluster_name}/nodes/{node_ip}](#inspect-a-node) *Inspect a node*
  - [DELETE /v1/vclusters/{vcluster_name}/nodes/{node_ip}](#delete-a-node) *Delete a node*

#### List all vclusters
```
GET /v1/vclusters
```
Example request:
```
GET /v1/vclusters
```
Example response:
```
[
    {
        "created": "2017-09-01T11:43:42.639081779+08:00",
        "id": "dbfedad1ebe4e6a2c105066055493003",
        "name": "dataman2",
        "nodes": [
            {
                "attrs": {
                    "cluster": "dataman2"
                },
                "id": "07239dfa-2982-4046-8f1c-f8f3c14d4649-S4",
                "ip": "172.16.1.18"
            }
        ],
        "updated": "2017-09-01T11:43:42.639081875+08:00"
    }
]
```

#### Create new vcluster
```
POST /v1/vclusters
```

Example request:
```
POST /v1/vclusters HTTP/1.1
Content-Type: application/json
{
        "name": "dataman2"
}
```

Example response:
```
HTTP/1.1 201 Created
Content-Type: application/json

{
    "created": "2017-09-01T15:49:27.325008363+08:00",
    "id": "c074a0cae43d753fb8d6a36dc5da1452",
    "name": "dataman",
    "nodes": null,
    "updated": "2017-09-01T15:49:27.325008397+08:00"
}
```

#### Inspect a vluster
```
GET /v1/vclusters/{vcluster_name}
```
Example request:
```
GET /v1/vclusters/dataman HTTP/1.1
```
Example response:
```
HTTP/1.1 200 OK
Content-Type: application/json
{
    "created": "2017-09-01T15:49:27.325008363+08:00",
    "id": "c074a0cae43d753fb8d6a36dc5da1452",
    "name": "dataman",
    "nodes": [
        {
            "attrs": {
                "cluster": "dataman"
            },
            "id": "07239dfa-2982-4046-8f1c-f8f3c14d4649-S4",
            "ip": "172.16.1.18"
        },
        {
            "attrs": {
                "cluster": "dataman",
                "disk": "ssd"
            },
            "id": "f40cf79498eb08ba610409a1c418847b",
            "ip": "172.16.1.17"
        }
    ],
    "updated": "2017-09-01T15:49:27.325008397+08:00"
}
```

#### Delete a vcluster
```
DELETE /v1/vclusters/{vcluster_name}
```
Example request:
```
DELETE /v1/vclusters/dataman
```
Example response:
```
HTTP/1.1 204 No Content
```


#### Add node
```
POST /v1/vclusters/{vcluster_name}/nodes
```
Example request:
```
POST /v1/vclusters/{vcluster_name}/nodes HTTP/1.1
Content-Type: application/json

{
    "IP": "172.16.1.18"
}
```
Example response:
```
HTTP/1.1 201 Created
Content-Type: application/json
{
    "attrs": {
        "cluster": "dataman"
    },
    "id": "f40cf79498eb08ba610409a1c418847b",
    "ip": "172.16.1.17"
}
```
#### Update node attribute
```
PATCH /v1/vclusters/{vcluster_name}/nodes/{node_ip}
```
Example request:
```
PATCH /v1/vclusters/{vcluster_name}/nodes/{node_ip} HTTP/1.1
Content-Type: application/json
{
	"key": "disk",
	"value": "ssd"
}
```
Example response:
```
HTTP/1.1 200 OK 
Content-Type: application/json
{
    "attrs": {
        "cluster": "dataman",
        "disk": "ssd"
    },
    "id": "f40cf79498eb08ba610409a1c418847b",
    "ip": "172.16.1.17"
}
```
#### List all nodes
```
GET /v1/vclusters/{vcluster_name}/nodes
```
Example request:
```
GET /v1/vclusters/dataman/nodes HTTP/1.1
```
Example response:
```
HTTP/1.1 200 OK
Content-Type: application/json

[
    {
        "attrs": {
            "cluster": "dataman"
        },
        "id": "07239dfa-2982-4046-8f1c-f8f3c14d4649-S4",
        "ip": "172.16.1.18"
    },
    {
        "attrs": {
            "cluster": "dataman",
            "disk": "ssd"
        },
        "id": "f40cf79498eb08ba610409a1c418847b",
        "ip": "172.16.1.17"
    }
]
```

#### Inspect a node
```
GET /v1/vclusters/{vcluster_name}/nodes/{node_ip}
```
Example request:
```
GET /v1/vclusters/dataman/nodes/172.16.1.18 HTTP/1.1
```
Example response:
```
HTTP/1.1 200 OK
Content-Type: application/json
{
    "attrs": {
        "cluster": "dataman"
    },
    "id": "07239dfa-2982-4046-8f1c-f8f3c14d4649-S4",
    "ip": "172.16.1.18"
}
```

#### Delele a node
```
DELETE /v1/vclusters/{vcluster_name}/nodes/{node_ip}
```
Example request:
```
DELETE /v1/vclusters/dataman/nodes/172.16.1.18 HTTP/1.1
```
Example response:
```
HTTP/1.1 204 No Content
```




