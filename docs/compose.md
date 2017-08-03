
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
		"listen": ":99",
		"sticky": false
	  }
    },
    "cache": {
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

Response:
```json
201:

400:
yaml convert: yaml: line 104: did not find expected key
missing dependency: cache -> dbmaster
dependency circled: [cache dbmaster cache]
at least one of ServiceGroup or YamlRaw required
...

409: conflict

500:
```

### list
`GET` `/v1/compose`

### get
`GET` `/v1/compose/b`  
`GET` `/v1/compose/643a4c42-9f7b-4a65-82f5-7d89d1d7c766`

Response
```json
{
  "id": "643a4c42-9f7b-4a65-82f5-7d89d1d7c766",
  "name": "b",
  "desc": "demo instance",
  "version_id": "xxx",
  "status": "ready",      // creating, ready, failed
  "errmsg": "",
  "created_at": "2017-05-09T10:39:06.406874508+08:00",
  "updated_at": "2017-05-09T10:39:27.066117222+08:00",
  "service_group": {   // same as above
  },
  "yaml_raw": "",      // same as above
  "yaml_extra": {      // same as above
  },
  "apps": [            // related apps
    {
      "id": "web-b-bbk-datamanmesos",
      "name": "web",
      "instances": 3,
      "updatedInstances": 0,
      "runningInstances": 3,
      "runAs": "bbk",
      "priority": 0,
      "clusterID": "datamanmesos",
      "created": "2017-05-09T10:39:17.543846121+08:00",
      "updated": "2017-05-09T10:39:17.543846336+08:00",
      "mode": "replicates",
      "state": "normal",
      "currentVersion": {
        "id": "1494297557",
        "appName": "web",
        "appVersion": "xxx",
        "cmd": "sleep 100d",
        "cpus": 0.01,
        "mem": 50,
        "disk": 100,
        "instances": 3,
        "runAs": "bbk",
        "priority": 0,
        "container": {
          "type": "docker",
          "docker": {
            "image": "nginx:latest",
            "network": "bridge",
            "parameters": [
              {
                "key": "ipc",
                "value": "host"
              },
              {
                "key": "mac-address",
                "value": "02:42:ac:11:65:43"
              },
              {
                "key": "user",
                "value": "root"
              },
              {
                "key": "tty",
                "value": "true"
              },
              {
                "key": "name",
                "value": "my-web-container"
              },
              {
                "key": "pid",
                "value": "host"
              },
              {
                "key": "stop-signal",
                "value": "SIGTERM"
              },
              {
                "key": "restart",
                "value": "no"
              },
              {
                "key": "workdir",
                "value": "/"
              },
              {
                "key": "read-only",
                "value": "true"
              },
              {
                "key": "log-driver",
                "value": "syslog"
              },
              {
                "key": "hostname",
                "value": "foo"
              },
              {
                "key": "tmpfs",
                "value": "/run"
              },
              {
                "key": "tmpfs",
                "value": "/tmp"
              },
              {
                "key": "label",
                "value": "name=bbklab"
              },
              {
                "key": "label",
                "value": "description=bbklab desc"
              },
              {
                "key": "cap-drop",
                "value": "NET_ADMIN"
              },
              {
                "key": "cap-drop",
                "value": "SYS_ADMIN"
              },
              {
                "key": "dns",
                "value": "114.114.114.114"
              },
              {
                "key": "dns",
                "value": "8.8.8.8"
              },
              {
                "key": "env",
                "value": "DEMO=true"
              },
              {
                "key": "env",
                "value": "PROD=false"
              },
              {
                "key": "expose",
                "value": "80"
              },
              {
                "key": "expose",
                "value": "443"
              },
              {
                "key": "security-opt",
                "value": "label:user:USER"
              },
              {
                "key": "security-opt",
                "value": "label:role:ROLE"
              },
              {
                "key": "ulimit",
                "value": "nproc=65535:65535"
              },
              {
                "key": "ulimit",
                "value": "nofile=20000:40000"
              },
              {
                "key": "cap-add",
                "value": "ALL"
              },
              {
                "key": "device",
                "value": "/dev/tty10:/dev/tty10"
              },
              {
                "key": "dns-search",
                "value": "swan.local"
              },
              {
                "key": "add-host",
                "value": "bbk:127.0.0.1"
              },
              {
                "key": "add-host",
                "value": "google-dns:8.8.8.8"
              },
              {
                "key": "volume",
                "value": "/tmp:/data:rw"
              },
              {
                "key": "volume",
                "value": "/var/log:/log:ro"
              }
            ],
            "portMappings": [
              {
                "containerPort": 3000,
                "hostPort": 3000,
                "name": "3000",
                "protocol": "udp"
              },
              {
                "containerPort": 3001,
                "hostPort": 3001,
                "name": "3001",
                "protocol": "udp"
              },
              {
                "containerPort": 3002,
                "hostPort": 3002,
                "name": "3002",
                "protocol": "udp"
              },
              {
                "containerPort": 3003,
                "hostPort": 3003,
                "name": "3003",
                "protocol": "udp"
              },
              {
                "containerPort": 800,
                "hostPort": 8080,
                "name": "8080",
                "protocol": "tcp"
              },
              {
                "containerPort": 443,
                "hostPort": 8090,
                "name": "8090",
                "protocol": "tcp"
              }
            ],
            "privileged": true
          }
        },
        "labels": {
          "DM_COMPOSE_NAME": "b",
          "description": "bbklab desc",
          "name": "bbklab"
        },
        "healthCheck": {
          "protocol": "cmd",
          "value": "echo ok",
          "consecutiveFailures": 3,
          "intervalSeconds": 30,
          "timeoutSeconds": 10
        },
        "env": {
          "DEMO": "true",
          "PROD": "false"
        },
        "killPolicy": {
          "duration": 10
        }
      },
      "labels": {
        "DM_COMPOSE_NAME": "b",
        "description": "bbklab desc",
        "name": "bbklab"
      },
      "env": {
        "DEMO": "true",
        "PROD": "false"
      }
    },
    {
      "id": "cache-b-bbk-datamanmesos",
      "name": "cache",
      "instances": 1,
      "updatedInstances": 0,
      "runningInstances": 1,
      "runAs": "bbk",
      "priority": 0,
      "clusterID": "datamanmesos",
      "created": "2017-05-09T10:39:13.469840343+08:00",
      "updated": "2017-05-09T10:39:13.469840415+08:00",
      "mode": "replicates",
      "state": "normal",
      "currentVersion": {
        "id": "1494297553",
        "appName": "cache",
        "appVersion": "xxx",
        "cpus": 0.02,
        "mem": 100,
        "disk": 33,
        "instances": 1,
        "runAs": "bbk",
        "priority": 0,
        "container": {
          "type": "docker",
          "docker": {
            "image": "redis:alpine",
            "network": "bridge",
            "parameters": [
              {
                "key": "read-only",
                "value": "false"
              },
              {
                "key": "tty",
                "value": "false"
              }
            ]
          }
        },
        "labels": {
          "DM_COMPOSE_NAME": "b"
        }
      },
      "labels": {
        "DM_COMPOSE_NAME": "b"
      }
    },
    {
      "id": "dbslave-b-bbk-datamanmesos",
      "name": "dbslave",
      "instances": 1,
      "updatedInstances": 0,
      "runningInstances": 1,
      "runAs": "bbk",
      "priority": 0,
      "clusterID": "datamanmesos",
      "created": "2017-05-09T10:39:09.38498159+08:00",
      "updated": "2017-05-09T10:39:09.384981688+08:00",
      "mode": "replicates",
      "state": "normal",
      "currentVersion": {
        "id": "1494297549",
        "appName": "dbslave",
        "appVersion": "xxx",
        "cmd": "sleep 100d",
        "cpus": 0.03,
        "mem": 100,
        "disk": 0,
        "instances": 1,
        "runAs": "bbk",
        "priority": 0,
        "container": {
          "type": "docker",
          "docker": {
            "image": "busybox:latest",
            "network": "host",
            "parameters": [
              {
                "key": "tty",
                "value": "false"
              },
              {
                "key": "read-only",
                "value": "false"
              }
            ]
          }
        },
        "labels": {
          "DM_COMPOSE_NAME": "b"
        }
      },
      "labels": {
        "DM_COMPOSE_NAME": "b"
      }
    },
    {
      "id": "dbmaster-b-bbk-datamanmesos",
      "name": "dbmaster",
      "instances": 1,
      "updatedInstances": 0,
      "runningInstances": 1,
      "runAs": "bbk",
      "priority": 0,
      "clusterID": "datamanmesos",
      "created": "2017-05-09T10:39:06.511004581+08:00",
      "updated": "2017-05-09T10:39:06.511004689+08:00",
      "mode": "replicates",
      "state": "normal",
      "currentVersion": {
        "id": "1494297546",
        "appName": "dbmaster",
        "appVersion": "xxx",
        "cmd": "sleep 100d",
        "cpus": 0.03,
        "mem": 100,
        "disk": 0,
        "instances": 1,
        "runAs": "bbk",
        "priority": 0,
        "container": {
          "type": "docker",
          "docker": {
            "image": "busybox:latest",
            "network": "host",
            "parameters": [
              {
                "key": "read-only",
                "value": "false"
              },
              {
                "key": "tty",
                "value": "false"
              }
            ]
          }
        },
        "labels": {
          "DM_COMPOSE_NAME": "b"
        }
      },
      "labels": {
        "DM_COMPOSE_NAME": "b"
      }
    }
  ]
}
```

### remove
`DELETE` `/v1/compose/b`  
`DELETE` `/v1/compose/643a4c42-9f7b-4a65-82f5-7d89d1d7c766`

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
