#swan-cluster

  a swan-cluster contains one or more swan-node, and one swan-node can display one of the following three roles, manager, agent, mixed. 
The proxy and DNS server was run in the agent, the manager provide service of manage application, and persist all app and cluster data.
So there must have one manger and one agent in a swan-cluster.

  we can start a swan-node as manager by the follower cmd:
```
sudo ./swan --zk-path=zk://192.168.59.104:2181/mesos 
            --listen-addr=0.0.0.0:9999 --raft-listen-addr=http://127.0.0.1:2111 
            --data-dir=./data/1/ --mode=manager
```

 as the same we start a swan-node with mixed mode
```
sudo ./swan --zk-path=zk://192.168.59.104:2181/mesos 
            --listen-addr=0.0.0.0:9999 --raft-listen-addr=http://127.0.0.1:2111 
            --data-dir=./data/1/ --mode=mixed
```
 if swan-node is started as mixed mode, it contains a manager and an agent.


 if the advertise address if not the same as the listen address of the swan-node. The start parameter **advertise-addr** and
**raft-advertise-addr** need to fill out. The start cmd :
```
sudo ./swan --zk-path=zk://192.168.59.104:2181/mesos 
            --listen-addr=0.0.0.0:9999 --advertise-addr=192.168.1.111
            --raft-listen-addr=http://127.0.0.1:2111 raft-advertise-add=http://192.168.1.111:2111
            --data-dir=./data/1/ --mode=manager
```

 a new swan-node aslo can join to an exist swan-cluster with the followe cmd:
```
sudo ./swan --zk-path=zk://192.168.59.104:2181/mesos 
            --listen-addr=0.0.0.0:9997 --raft-listen-addr=http://127.0.0.1:2113
            --data-dir=./data/3/ --mode=manager 
            --join-addrs=0.0.0.0:9999
```
 the **join-addrs** contains one of more manager advertise-addr which already in swan-cluster

  now we can run a swan-cluster with 3 manager and 1 agent like this:
```
 sudo ./swan --zk-path=zk://192.168.59.104:2181/mesos --listen-addr=0.0.0.0:9999 --raft-listen-addr=http://127.0.0.1:2111 --data-dir=./data/1/ --mode=manager
 sudo ./swan --zk-path=zk://192.168.59.104:2181/mesos --listen-addr=0.0.0.0:9998 --raft-listen-addr=http://127.0.0.1:2112 --data-dir=./data/2/ --mode=manager --join-addrs=0.0.0.0:9999
 sudo ./swan --zk-path=zk://192.168.59.104:2181/mesos --listen-addr=0.0.0.0:9997 --raft-listen-addr=http://127.0.0.1:2113 --data-dir=./data/3/ --mode=manager --join-addrs=0.0.0.0:9999
 sudo ./swan --listen-addr=0.0.0.0:9997 --mode=agent --join-addrs=0.0.0.0:9999
```
