# Depreciated, update needed

## `swan` 启动参数

```

➜  swan git:(config-refactor) ./bin/swan --help
NAME:
   swan - A general purpose mesos framework

USAGE:
   swan [global options] command [command options] [arguments...]

VERSION:
   0.01-e37b8e2

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --cluster-addrs value        address api server listen on, eg. 192.168.1.1:9999,192.168.1.2:9999 [$SWAN_CLUSTER]
   --zk-path value              zookeeper paths. eg. zk://host1:port1,host2:port2,.../path [$SWAN_ZKPATH]
   --log-level value, -l value  customize debug level [debug|info|error] [$SWAN_LOG_LEVEL]
   --raftid value               unique raft node id within the cluster. should be 1,2 or 3 within a 3-managers cluster (default: 0) [$SWAN_RAFT_ID]
   --raft-cluster value         raft cluster peers. eg. 192.168.1.1:1211,192.168.1.2.1211 [$SWAN_RAFT_CLUSTER]
   --mode value                 server mode, manager|agent|mixed [$SWAN_MODE]
   --data-dir value, -d value   swan data store dir [$SWAN_DATA_DIR]
   --domain value               domain which resolve to proxies. eg. access a slot by 0.appname.runas.clustername.domain [$SWAN_DOMAIN]
   --help, -h                   show help
   --version, -v                print the version

```


### --cluster-addrs `Swan` API监听的端口列表

对于单Manager节点，只需提供当前节点的监听地址；
对于多Manager节点，提供多个地址，以逗号分隔。其中第`raftId`个节点为当前Manager监听地址，其余为Peer Manager节点API的地址，
Raft Leader切换以后到当前Manager的请求会被Proxy到相应Leader上, 所有一个Manager需要知道其他Manager的地址。

```
 192.168.1.1:9999

  or

 192.168.1.1:9999,192.168.1.2:9999,192.168.1.3:9999
```

环境变量
```
SWAN_CLUSTER_ADDRS
```

### --zk-path value `Swan`所对应的Mesos再Zookeeper上的地址， 对应Mesos-Master的--zk配置。

```
zk://host1:port1,host2:port2,.../path
```

环境变量
```
SWAN_ZKPATH
```

### --log-level `Swan` 日志级别

debug日志级别对应开发环境， 日志内容会很多； info或者error级别对应生成环境

```
debug|info|error
```

环境变量
```
SWAN_LOG_LEVEL
```


### --raftid raft集群的节点ID，
用于区别不同的Raft节点和做`raft-cluster`以及`cluster-addrs`的Index。单Manager节点默认值为1；三Manager节点分别为1，2，3

环境变量
```
SWAN_RAFT_ID
```

### --raft-cluster  Raft协议通信地址，用于Raft节点直接日志同步和通信用。
```
0.0.0.0:1211 #单manager

or

192.168.1.1:1211,192.168.1.2:1211,192.168.1.3:1211
```

### --mode 启动节点的角色
```
manager # 当前节点作为Manager节点， 只做状态机和Raft同步作用
agent   # 当前节点作为Agent节点，做Proxy和DNS Proxy作用
mixed   # 当前节点同时作为Manager和Agent节点用
```

环境变量
```
SWAN_MODE
```


### --data-dir `Swan`数据存储地址 默认地址为./data

环境变量
```
SWAN_DATA_DIR
```

### --domain `Swan` Proxy和DNS功能对应的域名
对于Proxy功能，此域名的作用是帮助用户通过地址`0.app.user.cluster.domain`访问到Task；
对于DNS应用，DNS会把此域名解析到Proxy监听的IP，当前即`cluster-addrs`对应的IP

环境变量
```
SWAN_DOMAIN
```
