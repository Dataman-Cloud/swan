
- [compose api](#compose-API)
  + [parse](#parse)
  + [create](#create)
  + [list](#list)
  + [get](#get)
  + [remove](#remove)
- [docker compose v3](#docker-compose-v3)
  + [support](#support)
  + [example](#example)

## compose-API
### parse
`POST` `/v1/compose/parse`

Request
```json
{
	"yaml": "docker composer v3 yaml text here ..."
}
```

Response
```json
{
  "services": [
    "dbmaster",
    "web",
    "cache",
    "dbslave"
  ],
  "variables": [
    "RESTART",
    "WORKDIR",
    "HOSTNAME"
  ]
}
```

### create
`POST` `/v1/compose`

Request:
```json
{
  "name": "b",
  "desc": "demo instance",
  "service_group": {},
  "yaml_raw": "version: \"3\"\n\nservices:\n  web:\n    cap_add:\n      - ALL\n    cap_drop:\n      - NET_ADMIN\n      - SYS_ADMIN\n    command: \"sleep 100d\"\n    # cgroup_parent: \"/system.slice\"\n    container_name: \"my-web-container\"\n    deploy:\n      mode: replicated\n      replicas: 3\n    devices:\n        - \"/dev/tty10:/dev/tty10\"  \n    depends_on:\n      - cache\n      - dbmaster\n      - dbslave\n    dns:\n      - 114.114.114.114\n      - 8.8.8.8\n    dns_search:\n      - swan.local \n    tmpfs:\n      - /run\n      - /tmp\n    environment:\n      - DEMO=true\n      - PROD=false\n    expose:\n      - 80\n      - 443\n    extra_hosts:\n      - \"bbk:127.0.0.1\"\n      - \"google-dns:8.8.8.8\"\n    healthcheck:\n      test: [\"CMD\", \"echo\", \"ok\"]\n      interval: 30s\n      timeout: 10s\n      retries: 3\n    image: \"nginx:latest\"\n    labels:\n      - \"name=bbklab\"\n      - \"description=bbklab desc\"\n    logging:\n      driver: syslog\n      # options:\n      # syslog-address: \"tcp://127.0.0.1:123\"\n    network_mode: \"bridge\"\n    pid: \"host\"\n    ipc: \"host\"\n    ports:\n      - \"3000-3003/udp\"\n      - \"8080:800/tcp\"\n      - \"8090:443\"\n    security_opt:\n      - label:user:USER\n      - label:role:ROLE\n    stop_grace_period: 10s\n    stop_signal: SIGTERM\n    ulimits:\n      nproc: 65535\n      nofile:\n        soft: 20000\n        hard: 40000\n    volumes:\n      - /tmp:/data:rw\n      - /var/log:/log:ro\n    restart: \"no\"\n    user: \"root\"\n    working_dir: \"/\"\n    domainname: \"foo.com\"\n    hostname: \"foo\"\n    mac_address: 02:42:ac:11:65:43\n    privileged: true\n    read_only: true\n    stdin_open: true\n    tty: true\n\n  cache:\n    image: \"redis:alpine\"\n    network_mode: \"bridge\"\n    deploy:\n      mode: replicated\n      replicas: 1\n    depends_on:\n      - dbmaster\n      - dbslave\n\n  dbslave:\n    image: \"busybox:latest\"\n    command: \"sleep 100d\"\n    network_mode: \"host\"\n    depends_on:\n      - dbmaster\n\n  dbmaster:\n    image: \"busybox:latest\"\n    command: \"sleep 100d\"\n    network_mode: \"host\"\n",
  "yaml_env": {
	"WORKDIR": "/bbklab",
	"HOSTNAME": "damn"
  },
  "yaml_extra": {
    "web": {
      "cluster": "xxx",
      "runas": "bbk",
      "wait_delay": 1,
      "pull_always": false,
      "constraints": "",
      "uris": null,
      "ips": null,
      "resource": {
        "cpus": 0.01,
        "gpus": 0,
        "mem": 50,
        "disk": 100
      },
	  "proxy": {
		"enabled": true,
		"alias": "g.cn",
		"listen": 99,
		"sticky": false
	  }
    },
    "cache": {
      "cluster": "xxx",
      "runas": "bbk",
      "wait_delay": 1,
      "pull_always": false,
      "constraints": "",
      "uris": null,
      "ips": null,
      "resource": {
        "cpus": 0.02,
        "gpus": 0,
        "mem": 100,
        "disk": 33
      },
	  "proxy": {
		"enabled": false,
	  }
    },
    "dbmaster": {
      "cluster": "xxx",
      "runas": "bbk",
      "wait_delay": 1,
      "pull_always": false,
      "constraints": "",
      "uris": null,
      "ips": null,
      "resource": {
        "cpus": 0.03,
        "gpus": 0,
        "mem": 100
      },
	  "proxy": {
		"enabled": false,
	  }
    },
    "dbslave": {
      "cluster": "xxx",
      "runas": "bbk",
      "wait_delay": 1,
      "pull_always": false,
      "constraints": "",
      "uris": null,
      "ips": null,
      "resource": {
        "cpus": 0.03,
        "gpus": 0,
        "mem": 100
      },
	  "proxy": {
		"enabled": false,
	  }
    }
  }
}
```

### list
`GET` `/v1/compose`

### get
`GET` `/v1/compose/a`  
`GET` `/v1/compose/1653a0b51ae8f5ff`

Response
```json
{
  "id": "1653a0b51ae8f5ff",
  "name": "a",
  "display_name": "a.bbk.dataman-mesos",
  "desc": "demo instance",
  "op_status": "noop",   // creating, deleting, noop
  "errmsg": "",
  "created_at": "2017-08-04T15:40:45.209451241+08:00",
  "updated_at": "2017-08-04T15:40:56.251093817+08:00",
  "yaml_raw": "", // same as above
  "yaml_env": {
    "HOSTNAME": "damnhost",
    "WORKDIR": "/bbklab"
  },
  "yaml_extra": {}, // same as above
  "service_group": {}, // same as above
  "apps": [
    {
      "id": "dbmaster.a.bbk.dataman-mesos",
      "name": "dbmaster",
      "runAs": "bbk",
      "cluster": "dataman-mesos",
      "operationStatus": "noop",
      "progress": 0,
      "task_count": 1,
      "currentVersion": [
        "1501832445215187780"
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
      "created": "2017-08-04T15:40:45.215218542+08:00",
      "updated": "2017-08-04T15:40:49.051227394+08:00"
    },
    {
      "id": "cache.a.bbk.dataman-mesos",
      "name": "cache",
      "runAs": "bbk",
      "cluster": "dataman-mesos",
      "operationStatus": "noop",
      "progress": 0,
      "task_count": 1,
      "currentVersion": [
        "1501832450869902669"
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
      "created": "2017-08-04T15:40:50.869925765+08:00",
      "updated": "2017-08-04T15:40:52.68310734+08:00"
    },
    {
      "id": "dbslave.a.bbk.dataman-mesos",
      "name": "dbslave",
      "runAs": "bbk",
      "cluster": "dataman-mesos",
      "operationStatus": "noop",
      "progress": 0,
      "task_count": 1,
      "currentVersion": [
        "1501832449054682772"
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
      "created": "2017-08-04T15:40:49.054705667+08:00",
      "updated": "2017-08-04T15:40:50.866596828+08:00"
    },
    {
      "id": "web.a.bbk.dataman-mesos",
      "name": "web",
      "runAs": "bbk",
      "cluster": "dataman-mesos",
      "operationStatus": "noop",
      "progress": 0,
      "task_count": 3,
      "currentVersion": [
        "1501832452686746879"
      ],
      "version_count": 1,
      "status": "available",
      "tasks_status": {
        "TASK_RUNNING": 3
      },
      "health": {
        "total": 3,
        "healthy": 3,
        "unhealthy": 0,
        "unset": 0
      },
      "errmsg": "",
      "created": "2017-08-04T15:40:52.68680238+08:00",
      "updated": "2017-08-04T15:40:56.247136433+08:00"
    }
  ]
}
```

### remove
`DELETE` `/v1/compose/a`  
`DELETE` `/v1/compose/1653a0b51ae8f5ff`

Response:
204

## docker-compose-v3
[docker compose v3 reference](https://docs.docker.com/compose/compose-file/)  
[variable substitution](https://docs.docker.com/compose/compose-file/#variable-substitution)

### support
| Entry  | - | Example | 
|------|---------|------|
| `cap_add` | OK | - |
| `cap_drop` | OK | - |
| `command` | OK | - | 
| `cgroup_parent` | OK | - | 
| `deploy` | PART | - |
| `devices` | OK | - | 
| `depends_on` | OK  | - |
| `dns` | OK | - |
| `tmpfs` | OK | - |
| `environment` | OK | - |
| `expose` | OK | - | 
| `extra_hosts` | OK | - |
| `healthcheck` | OK | - |
| `image` | OK | - |
| `labels` | OK  | - |
| `logging` | OK | - |
| `network_mode` | OK | - |
| `pid` | OK | - |
| `ipc` | OK | - |
| `ports` | OK | - |
| `security_opt` | OK | - |
| `stop_grace_period` | OK | - |
| `stop_signal` | OK | - |
| `ulimits` | OK | - |
| `volumes` | PART | - |
| `restart` | OK | - |
| `user` | OK | - |
| `working_dir` | OK | - |
| `hostname` | OK | - |
| `mac_address` | OK | - |
| `privileged` | OK | - |
| `read_only` | OK | - |
| `tty` | OK  | - |
| `dns_search` | OK | |
| `stdin_open` | NO, rewrite by mesos agent  | - |
| `container_name` | NO, rewrite by mesos agent | - |
| `entrypoint` | NO, rewrite by mesos agent | - |
| `isolation` | NO, fixed `default` under linux  | - |
| `domainname` | NO | - |
| `shm_size` | NO | - |
| `sysctls` | NO  | - |
| `external_links` | NO  | - |
| `links` | NO  | - |
| `build` | NO  | - |
| `userns_mode` | NO  | - |
| `secrets` | NO  | - |
| `networks` | NO  | - |


### example 
```yaml
version: "3"

services:
  web:
    cap_add:
      - ALL
    cap_drop:
      - NET_ADMIN
      - SYS_ADMIN
    command: "sleep 100d"
    # cgroup_parent: "/system.slice"
    container_name: "my-web-container"
    deploy:
      mode: replicated
      replicas: 3
    devices:
        - "/dev/tty10:/dev/tty10"  
    depends_on:
      - cache
      - dbmaster
      - dbslave
    dns:
      - 114.114.114.114
      - 8.8.8.8
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
    logging:
      driver: syslog
      # options:
      # syslog-address: "tcp://127.0.0.1:123"
    network_mode: "bridge"
    pid: "host"
    ipc: "host"
    ports:
      - "3000-3003/udp"
      - "8080:800/tcp"
      - "8090:443"
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
    restart: "no"
    user: "root"
    working_dir: "${WORKDIR-/}"             # env WORKDIR, default: /
    domainname: "foo.com"
    hostname: "${HOSTNAME-defaulthostname}" # env HOSTNAME, default: defaulthostname
    mac_address: 02:42:ac:11:65:43
    privileged: true
    read_only: true
    stdin_open: true
    tty: true

  cache:
    image: "redis:alpine"
    network_mode: "bridge"
    deploy:
      mode: replicated
      replicas: 1
    depends_on:
      - dbmaster
      - dbslave

  dbslave:
    image: "busybox:latest"
    command: "sleep 100d"
    network_mode: "host"
    depends_on:
      - dbmaster

  dbmaster:
    image: "busybox:latest"
    command: "sleep 100d"
    network_mode: "host"
```
