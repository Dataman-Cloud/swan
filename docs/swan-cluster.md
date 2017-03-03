#swan-cluster

  a swan-cluster contains one or more swan-node, and one swan-node can display one of the following two roles, manager, agent. 
The proxy and DNS server was run in the agent, the manager provide service of manage application, and persist all app and cluster data.
So there must have one manger and one agent in a swan-cluster.

  we can start a swan-node as manager by the following command:
```
./swan manager init --zk-path=zk://192.168.59.104:2181/mesos 
                    --listen-addr=0.0.0.0:9999 --raft-listen-addr=http://127.0.0.1:2111 
                    --data-dir=./data
```

 as the same we start a swan-node with agent mode
```
sudo ./swan agent join --listen-addr=0.0.0.0:9999 --join-addrs=0.0.0.0:9999
                       --data-dir=./data
```
### NOTICE
* swan agent need to listen 80 and 53/udp port so need **sudo** authority
* event there is nor persist data in agent the **-data-dir** is need to set, beacuse the agent ID is store on disk.

 if the advertise address if not the same as the listen address of the swan-node. The start parameter **advertise-addr** and
**raft-advertise-addr** need to be provided. The start command :
```
./swan manager init --zk-path=zk://192.168.59.104:2181/mesos 
                    --listen-addr=0.0.0.0:9999 --advertise-addr=192.168.1.111
                    --raft-listen-addr=http://127.0.0.1:2111 raft-advertise-addr=http://192.168.1.111:2111
                    --data-dir=./data/1/
```

 a new swan-node aslo can join to an existing swan-cluster with the following command:
```
./swan manager join --zk-path=zk://192.168.59.104:2181/mesos 
                    --listen-addr=0.0.0.0:9997 --raft-listen-addr=http://127.0.0.1:2113
                    --data-dir=./data/3/
                    --join-addrs=0.0.0.0:9999
```
 the **join-addrs** contains one of more managers advertise-addrs which already in swan-cluster

  now we can run a swan-cluster with 3 managers and 2 agents as follow :
```
./swan manager init --zk-path=zk://192.168.59.104:2181/mesos --listen-addr=0.0.0.0:9999 --raft-listen-addr=http://127.0.0.1:2111 --data-dir=./data/1/
./swan manager join --zk-path=zk://192.168.59.104:2181/mesos --listen-addr=0.0.0.0:9998 --raft-listen-addr=http://127.0.0.1:2112 --data-dir=./data/2/ --join-addrs=0.0.0.0:9999
./swan manager join --zk-path=zk://192.168.59.104:2181/mesos --listen-addr=0.0.0.0:9997 --raft-listen-addr=http://127.0.0.1:2113 --data-dir=./data/3/ --join-addrs=0.0.0.0:9999
 sudo ./swan agent join --listen-addr=0.0.0.0:9996 --join-addrs=0.0.0.0:9999,0.0.0.0:9998 --data-dir=./data/4
 sudo ./swan agent join --listen-addr=0.0.0.0:9995 --join-addrs=0.0.0.0:9999,0.0.0.0:9997 --data-dir=./data/5
```

## Description of persist data
  all swan persist data was in the manager node data dir, the default data dir is **./data/**, and there has an file named ID if this file was found swan will start with history data.
