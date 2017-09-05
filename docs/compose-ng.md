- [编排API](#compose)
  + [YAML解析](#parse)
  + [运行编排实例](#create)  
  + [列出编排实例（不包含应用列表）](#list)
  + [单个编排实例详情(包含应用列表)](#get)
  + [单个编排服务间依赖关系](#dependency)
  + [删除编排实例](#remove)
- [docker compose v3](#docker-compose-v3)
  + [支持的字段](#support)
  + [废弃的字段](#unsupport)
  + [扩展的字段](#extend)
  + [YAML样例（wordpress＋mariadb）](#example)

## compose
### parse
`POST` `/v1/compose-ng/parse`

Request:  

Content-Type: `text/plain`
```yaml
version: "3"

services:
  web:
    cap_add:
      - ALL
    cap_drop:
      - NET_ADMIN
      - SYS_ADMIN
    deploy:
      replicas: 3
      constraints:
        - attribute: "hostname"
          operator: "~="
          value: "130"
    devices:
      - "/dev/tty10:/dev/tty10"
    depends_on:
      - cache
      - dbmaster
      - dbslave
    dns_search:
      - swan.local 
    tmpfs:
      - /run
      - /tmp
    environment:
      - DEMO=true
      - PROD=false
    expose:
      - 80
      - 443
    extra_hosts:
      - "bbk:127.0.0.1"
      - "google-dns:8.8.8.8"
    healthcheck:
      test: ["CMD", "echo", "ok"]
      interval: 30s
      timeout: 10s
      retries: 3
    image: "nginx:latest"
    labels:
      - "name=bbklab"
      - "description=bbklab desc"
    network_mode: "bridge"
    pid: "host"
    ipc: "host"
    ports:
      - "80/tcp"
      - "443/tcp"
    security_opt:
      - label:user:USER
      - label:role:ROLE
    stop_grace_period: 10s
    stop_signal: SIGTERM
    ulimits:
      nproc: 65535
      nofile:
        soft: 20000
        hard: 40000
    volumes:
      - /tmp:/data:rw
      - /var/log:/log:ro
    restart: "$RESTART"
    user: "root"
    working_dir: "/"
    domainname: "foo.com"
    hostname: "${HOSTNAME}"
    privileged: true
    read_only: false
    stdin_open: true
    tty: true

  cache:
    image: "nginx:alpine"
    network_mode: "bridge"
    deploy:
      replicas: 1
      constraints:
        - attribute: "hostname"
          operator: "~="
          value: "130"
    depends_on:
      - dbmaster
      - dbslave
    ports:
      - "80/tcp"
      - "443/tcp"

  dbslave:
    image: "nginx:latest"
    network_mode: "bridge"
    deploy:
      replicas: 1
      wait_delay: 5
      constraints:
        - attribute: "hostname"
          operator: "~="
          value: "130"    
    ports:
      - "80/tcp"
      - "443/tcp"
    depends_on:
      - dbmaster

  dbmaster:
    image: "nginx:latest"
    network_mode: "bridge"
    deploy:
      replicas: 1
      wait_delay: 10
      constraints:
        - attribute: "hostname"
          operator: "~="
          value: "130"    
    ports:
      - "80/tcp"
      - "443/tcp"

  admin:
    image: "busybox:latest"
    command: "sleep 100d"
    network_mode: "bridge"
    deploy:
      replicas: 1
      constraints:
        - attribute: "hostname"
          operator: "~="
          value: "130"    
    resource:
      cpus: 0.01
      mem: 10
```

Response
```json
{
    "services": [
        "web",
        "admin",
        "cache",
        "dbmaster",
        "dbslave"
    ],
    "variables": [
        "RESTART:",
        "HOSTNAME:"
    ]
}
```

### create
`POST` `/v1/compose-ng?runas=bbk&name=demo&cluster=bbklab&labels=UID=1,GID=2&envs=PROD=true,DEMO=false`

Request:  

Query-Parameters:
  - name    编排名
  - runas   运行用户
  - cluster 集群名
  - labels  标签
  - envs    YAML文本中的变量替换所需的变量

Content-Type: `text/plain`
```yaml
version: "3"

services:
  wordpress:
    image: "wordpress"
    network_mode: "bridge"
    deploy:
      replicas: 1
      constraints:
        - attribute: "hostname"
          operator: "~="
          value: "130"
    environment:
      - WORDPRESS_DB_HOST=mariadb:3306
      - WORDPRESS_DB_PASSWORD=Password
    ports:
      - "80/tcp"
    depends_on:
      - mariadb
    proxy:
      enabled: true
      alias: "x.cn"
      listen: 8888
      sticky: false

  mariadb:
    image: "mariadb"
    network_mode: "bridge"
    deploy:
      replicas: 1
      wait_delay: 20
      constraints:
        - attribute: "hostname"
          operator: "~="
          value: "130"
    environment:
      - MYSQL_ROOT_PASSWORD=Password
    ports:
      - 3306
    proxy:
      enabled: true
      alias: "i.cn"
      listen: 3306
      sticky: false
```

### list
`GET` `/v1/compose-ng?GID=2&UID=1`

Query-Parameters:   
filter by any labels key-val

```json
[
    {
        "id": "170166598a6fee73",
        "name": "demo",
        "run_as": "bbk",
        "cluster": "bbklab",
        "display_name": "demo.bbk.bbklab",
        "desc": "",
        "op_status": "noop",
        "errmsg": "",
        "created_at": "2017-08-24T15:27:34.479455354+08:00",
        "updated_at": "2017-08-24T15:27:58.420563807+08:00",
        "labels": {
            "GID": "2",
            "UID": "1"
        },
        "yaml_raw": "version: \"3\"\n\nservices:\n  wordpress:\n    image: \"wordpress\"\n    network_mode: \"bridge\"\n    deploy:\n      replicas: 1\n      constraints:\n        - attribute: \"hostname\"\n          operator: \"~=\"\n          value: \"130\"\n    environment:\n      - WORDPRESS_DB_HOST=mariadb:3306\n      - WORDPRESS_DB_PASSWORD=Password\n    ports:\n      - \"80/tcp\"\n    depends_on:\n      - mariadb\n    dns:\n      - 192.168.1.196  # swan dns server\n    proxy:\n      enabled: true\n      alias: \"x.cn\"\n      listen: 8888\n      sticky: false\n\n  mariadb:\n    image: \"mariadb\"\n    network_mode: \"bridge\"\n    deploy:\n      replicas: 1\n      constraints:\n        - attribute: \"hostname\"\n          operator: \"~=\"\n          value: \"130\"\n    environment:\n      - MYSQL_ROOT_PASSWORD=Password\n    ports:\n      - 3306\n    dns:\n      - 192.168.1.196  # swan dns server\n    proxy:\n      enabled: true\n      alias: \"i.cn\"\n      listen: 3306\n      sticky: false",
        "yaml_env": {},
        "ComposeV3": {
            "Version": "3",
            "Variables": [],
            "Services": {
                "mariadb": {
                    "Name": "mariadb",
                    "CapAdd": null,
                    "CapDrop": null,
                    "CgroupParent": "",
                    "Command": null,
                    "ContainerName": "",
                    "DependsOn": null,
                    "Devices": null,
                    "Dns": [
                        "192.168.1.196"
                    ],
                    "DnsSearch": null,
                    "DomainName": "",
                    "Entrypoint": null,
                    "Environment": {
                        "MYSQL_ROOT_PASSWORD": "Password"
                    },
                    "Expose": null,
                    "ExternalLinks": null,
                    "ExtraHosts": null,
                    "Hostname": "",
                    "HealthCheck": null,
                    "Image": "mariadb",
                    "Ipc": "",
                    "Labels": null,
                    "Links": null,
                    "Logging": null,
                    "MacAddress": "",
                    "NetworkMode": "bridge",
                    "Pid": "",
                    "Ports": [
                        "3306"
                    ],
                    "Privileged": false,
                    "ReadOnly": false,
                    "Restart": "",
                    "SecurityOpt": null,
                    "StdinOpen": false,
                    "StopGracePeriod": "",
                    "StopSignal": "",
                    "ShmSize": "",
                    "Tmpfs": null,
                    "Tty": false,
                    "Ulimits": null,
                    "User": "",
                    "Volumes": null,
                    "WorkingDir": "",
                    "Deploy": {
                        "Replicas": 1,
                        "Constraints": [
                            {
                                "attribute": "hostname",
                                "operator": "~=",
                                "value": "130"
                            }
                        ]
                    },
                    "Resource": null,
                    "PullAlways": false,
                    "Proxy": {
                        "enabled": true,
                        "alias": "i.cn",
                        "listen": 3306,
                        "sticky": false
                    },
                    "URIs": null,
                    "IPs": null,
                    "WaitDelay": 20
                },
                "wordpress": {
                    "Name": "wordpress",
                    "CapAdd": null,
                    "CapDrop": null,
                    "CgroupParent": "",
                    "Command": null,
                    "ContainerName": "",
                    "DependsOn": [
                        "mariadb"
                    ],
                    "Devices": null,
                    "Dns": [
                        "192.168.1.196"
                    ],
                    "DnsSearch": null,
                    "DomainName": "",
                    "Entrypoint": null,
                    "Environment": {
                        "WORDPRESS_DB_HOST": "mariadb:3306",
                        "WORDPRESS_DB_PASSWORD": "Password"
                    },
                    "Expose": null,
                    "ExternalLinks": null,
                    "ExtraHosts": null,
                    "Hostname": "",
                    "HealthCheck": null,
                    "Image": "wordpress",
                    "Ipc": "",
                    "Labels": null,
                    "Links": null,
                    "Logging": null,
                    "MacAddress": "",
                    "NetworkMode": "bridge",
                    "Pid": "",
                    "Ports": [
                        "80/tcp"
                    ],
                    "Privileged": false,
                    "ReadOnly": false,
                    "Restart": "",
                    "SecurityOpt": null,
                    "StdinOpen": false,
                    "StopGracePeriod": "",
                    "StopSignal": "",
                    "ShmSize": "",
                    "Tmpfs": null,
                    "Tty": false,
                    "Ulimits": null,
                    "User": "",
                    "Volumes": null,
                    "WorkingDir": "",
                    "Deploy": {
                        "Replicas": 1,
                        "Constraints": [
                            {
                                "attribute": "hostname",
                                "operator": "~=",
                                "value": "130"
                            }
                        ]
                    },
                    "Resource": null,
                    "PullAlways": false,
                    "Proxy": {
                        "enabled": true,
                        "alias": "x.cn",
                        "listen": 8888,
                        "sticky": false
                    },
                    "URIs": null,
                    "IPs": null,
                    "WaitDelay": 0
                }
            }
        }
    }
]
```

### get
`GET` `/v1/compose-ng/demo`  
`GET` `/v1/compose-ng/170166598a6fee73`

Response
```json
{
    "id": "170166598a6fee73",
    "name": "demo",
    "run_as": "bbk",
    "cluster": "bbklab",
    "display_name": "demo.bbk.bbklab",
    "desc": "",
    "op_status": "noop",
    "errmsg": "",
    "created_at": "2017-08-24T15:27:34.479455354+08:00",
    "updated_at": "2017-08-24T15:27:58.420563807+08:00",
    "labels": {
        "GID": "2",
        "UID": "1"
    },
    "yaml_raw": "version: \"3\"\n\nservices:\n  wordpress:\n    image: \"wordpress\"\n    network_mode: \"bridge\"\n    deploy:\n      replicas: 1\n      constraints:\n        - attribute: \"hostname\"\n          operator: \"~=\"\n          value: \"130\"\n    environment:\n      - WORDPRESS_DB_HOST=mariadb:3306\n      - WORDPRESS_DB_PASSWORD=Password\n    ports:\n      - \"80/tcp\"\n    depends_on:\n      - mariadb\n    dns:\n      - 192.168.1.196  # swan dns server\n    proxy:\n      enabled: true\n      alias: \"x.cn\"\n      listen: 8888\n      sticky: false\n\n  mariadb:\n    image: \"mariadb\"\n    network_mode: \"bridge\"\n    deploy:\n      replicas: 1\n      constraints:\n        - attribute: \"hostname\"\n          operator: \"~=\"\n          value: \"130\"\n    environment:\n      - MYSQL_ROOT_PASSWORD=Password\n    ports:\n      - 3306\n    dns:\n      - 192.168.1.196  # swan dns server\n    proxy:\n      enabled: true\n      alias: \"i.cn\"\n      listen: 3306\n      sticky: false",
    "yaml_env": {},
    "ComposeV3": {
        "Version": "3",
        "Variables": [],
        "Services": {
            "mariadb": {
                "Name": "mariadb",
                "CapAdd": null,
                "CapDrop": null,
                "CgroupParent": "",
                "Command": null,
                "ContainerName": "",
                "DependsOn": null,
                "Devices": null,
                "Dns": [
                    "192.168.1.196"
                ],
                "DnsSearch": null,
                "DomainName": "",
                "Entrypoint": null,
                "Environment": {
                    "MYSQL_ROOT_PASSWORD": "Password"
                },
                "Expose": null,
                "ExternalLinks": null,
                "ExtraHosts": null,
                "Hostname": "",
                "HealthCheck": null,
                "Image": "mariadb",
                "Ipc": "",
                "Labels": null,
                "Links": null,
                "Logging": null,
                "MacAddress": "",
                "NetworkMode": "bridge",
                "Pid": "",
                "Ports": [
                    "3306"
                ],
                "Privileged": false,
                "ReadOnly": false,
                "Restart": "",
                "SecurityOpt": null,
                "StdinOpen": false,
                "StopGracePeriod": "",
                "StopSignal": "",
                "ShmSize": "",
                "Tmpfs": null,
                "Tty": false,
                "Ulimits": null,
                "User": "",
                "Volumes": null,
                "WorkingDir": "",
                "Deploy": {
                    "Replicas": 1,
                    "Constraints": [
                        {
                            "attribute": "hostname",
                            "operator": "~=",
                            "value": "130"
                        }
                    ]
                },
                "Resource": null,
                "PullAlways": false,
                "Proxy": {
                    "enabled": true,
                    "alias": "i.cn",
                    "listen": 3306,
                    "sticky": false
                },
                "URIs": null,
                "IPs": null,
                "WaitDelay": 20
            },
            "wordpress": {
                "Name": "wordpress",
                "CapAdd": null,
                "CapDrop": null,
                "CgroupParent": "",
                "Command": null,
                "ContainerName": "",
                "DependsOn": [
                    "mariadb"
                ],
                "Devices": null,
                "Dns": [
                    "192.168.1.196"
                ],
                "DnsSearch": null,
                "DomainName": "",
                "Entrypoint": null,
                "Environment": {
                    "WORDPRESS_DB_HOST": "mariadb:3306",
                    "WORDPRESS_DB_PASSWORD": "Password"
                },
                "Expose": null,
                "ExternalLinks": null,
                "ExtraHosts": null,
                "Hostname": "",
                "HealthCheck": null,
                "Image": "wordpress",
                "Ipc": "",
                "Labels": null,
                "Links": null,
                "Logging": null,
                "MacAddress": "",
                "NetworkMode": "bridge",
                "Pid": "",
                "Ports": [
                    "80/tcp"
                ],
                "Privileged": false,
                "ReadOnly": false,
                "Restart": "",
                "SecurityOpt": null,
                "StdinOpen": false,
                "StopGracePeriod": "",
                "StopSignal": "",
                "ShmSize": "",
                "Tmpfs": null,
                "Tty": false,
                "Ulimits": null,
                "User": "",
                "Volumes": null,
                "WorkingDir": "",
                "Deploy": {
                    "Replicas": 1,
                    "Constraints": [
                        {
                            "attribute": "hostname",
                            "operator": "~=",
                            "value": "130"
                        }
                    ]
                },
                "Resource": null,
                "PullAlways": false,
                "Proxy": {
                    "enabled": true,
                    "alias": "x.cn",
                    "listen": 8888,
                    "sticky": false
                },
                "URIs": null,
                "IPs": null,
                "WaitDelay": 0
            }
        }
    },
    "apps": [
        {
            "id": "wordpress.demo.bbk.bbklab",
            "name": "wordpress",
            "runAs": "bbk",
            "cluster": "bbklab",
            "operationStatus": "noop",
            "progress": -1,
            "progress_details": null,
            "task_count": 1,
            "currentVersion": [
                "1503559654480442429"
            ],
            "version_count": 1,
            "status": "available",
            "tasks_status": {
                "TASK_RUNNING": 1
            },
            "health": {
                "total": 1,
                "healthy": 0,
                "unhealthy": 0,
                "unset": 1
            },
            "errmsg": "",
            "created": "2017-08-24T15:27:56.560638396+08:00",
            "updated": "2017-08-24T15:27:58.41585524+08:00"
        },
        {
            "id": "mariadb.demo.bbk.bbklab",
            "name": "mariadb",
            "runAs": "bbk",
            "cluster": "bbklab",
            "operationStatus": "noop",
            "progress": -1,
            "progress_details": null,
            "task_count": 1,
            "currentVersion": [
                "1503559654480394180"
            ],
            "version_count": 1,
            "status": "available",
            "tasks_status": {
                "TASK_RUNNING": 1
            },
            "health": {
                "total": 1,
                "healthy": 0,
                "unhealthy": 0,
                "unset": 1
            },
            "errmsg": "",
            "created": "2017-08-24T15:27:34.498475803+08:00",
            "updated": "2017-08-24T15:27:56.557677806+08:00"
        }
    ]
}
```

### dependency
`GET` `/v1/compose-ng/demo/dependency`

Response
```json
{
    "admin": null,
    "cache": [
        "dbmaster",
        "dbslave"
    ],
    "dbmaster": null,
    "dbslave": [
        "dbmaster"
    ],
    "web": [
        "cache",
        "dbmaster",
        "dbslave"
    ]
}
```

### remove
`DELETE` `/v1/compose-ng/demo`  
`DELETE` `/v1/compose-ng/1653a0b51ae8f5ff`

Response:
204

## docker-compose-v3
[docker compose v3 reference](https://docs.docker.com/compose/compose-file/)  
[variable substitution](https://docs.docker.com/compose/compose-file/#variable-substitution)

### support
  - cap_add
  - cap_drop
  - command
  - cgroup_parent
  - devices
  - depends_on
  - dns
  - tmpfs
  - environment
  - expose
  - extra_hosts
  - healthcheck
  - image
  - labels
  - logging
  - network_mode
  - pid
  - ipc
  - ports
  - security_opt
  - stop_grace_period
  - stop_signal
  - ulimits
  - volumes
  - restart
  - user
  - working_dir
  - hostname`
  - mac_address
  - privileged
  - read_only
  - tty
  - dns_search

### unsupport
  - stdin_open     -   _rewrite by mesos agent_
  - container_name -  _rewrite by mesos agent_
  - entrypoint     -  _rewrite by mesos agent_
  - isolation      -  _fixed `default` under linux_
  - domainname
  - sysctls
  - external_links
  - links
  - build
  - shm_size
  - userns_mode
  - secrets
  - networks

### extend
#### deploy
```yaml
  deploy:
    replicas: 1      # 部署实例数量
    wait_delay: 20   # 部署该服务后等待时间，单位秒
    constraints:     # 选择部署节点的约束条件
      - attribute: "hostname"
        operator: "~="
        value: "130"
```
#### resource
```yaml
  resource
    cpus: 0.01
    mem: 10
    disk: 10
    gpus: 1
```
#### pull_always
```yaml
  pull_always: true
```
#### proxy  配置应用通过swan－proxy访问
```yaml
  proxy:
    enabled: true   # 开关
    alias: "x.cn"   # 7层访问别名
    listen: 8888    # 4层监听端口（将在所有节点上监听该端口）
    sticky: false   # 会话保持开关
```
#### uris
```yaml
  uris:
    - http://xxxx/abc.txt
    - http://yyyy/123.txt
```
#### ips
```yaml
  ips:
    - 192.168.1.101
    - 192.168.1.102
```



### example 
wordpress + mariadb
```yaml
version: "3"

services:
  wordpress:
    image: "wordpress"
    network_mode: "bridge"
    deploy:
      replicas: 1
      constraints:
        - attribute: "hostname"
          operator: "~="
          value: "192.*"
    environment:
      - WORDPRESS_DB_HOST=mariadb:3306
      - WORDPRESS_DB_PASSWORD=Password
    ports:
      - "80/tcp"
    depends_on:
      - mariadb
    proxy:
      enabled: true
      alias: "x.cn"
      listen: 8888
      sticky: false

  mariadb:
    image: "mariadb"
    network_mode: "bridge"
    deploy:
      replicas: 1
      wait_delay: 20
      constraints:
        - attribute: "hostname"
          operator: "~="
          value: "192.*"
    environment:
      - MYSQL_ROOT_PASSWORD=Password
    ports:
      - 3306
    proxy:
      enabled: true
      alias: "i.cn"
      listen: 3306
      sticky: false
```
