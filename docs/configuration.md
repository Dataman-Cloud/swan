## swan help

```bash

NAME:
   swan - A general purpose mesos framework

USAGE:
   swan [global options] command [command options] [arguments...]

VERSION:
   0.01-49d0b06

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --config-file value, -c value   specify config file path (default: "./config.json")
   --cluster value                 API Server address <ip:port> [$SWAN_CLUSTER]
   --mesos-master value, -m value  mesos master address host1:port1,host2:port2,... or zk://host1:port1,host2:port2,.../path [$SWAN_MESOS_MASTER]
   --log-level value, -l value     customize debug level [debug|info|error]
   --raftid value                  raft node id (default: 0) [$SWAN_RAFT_ID]
   --raft-cluster value            raft cluster peers addr [$SWAN_RAFT_CLUSTER]
   --mode value                    Server mode, manager|agent|mixed [$SWAN_MODE]
   --data-dir value, -d value      swan data store dir [$SWAN_DATA_DIR]
   --enable-proxy                  enable proxy or not [$SWAN_ENABLE_PROXY]
   --enable-dns                    enable dns resolver or not [$SWAN_ENABLE_DNS]
   --no-recover                    do not retry recover from previous crush [$SWAN_NO_RECOVER]
   --help, -h                      show help
   --version, -v                   print the version

```


