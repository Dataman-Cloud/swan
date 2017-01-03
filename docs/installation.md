## Setup Mesos Cluster

**Swan** acting  as an mesos framework, in order to experience each
feature of Swan, make sure mesos cluster setup properly. See [How to
setup mesos cluster](http://mesos.apache.org/documentation/latest/getting-started/)


## Build Swan from source code


### 1, clone latest release of [Swan](https://github.com/Dataman-Cloud/swan) from github into your $GOPATH

``` bash

  cd $GOPATH/src && git clone https://github.com/Dataman-Cloud/swan github.com/Dataman-Cloud/swan

```

### 2, build Golang source code

``` bash

  cd $GOPATH/src/github.com/Dataman-Cloud/swan/ && make

```

### 3, start Swan with standalone mode

``` bash

  sudo bin/swan --mesos-master=$MESOS_MASTER_PATH --log-level=debug --raftid=1 --raft-cluster=http://127.0.0.1:2111 --data-dir=./data --enable-dns --enable-proxy --cluster=0.0.0.0:9999

```

`Swan` require root permissions if you toggle enable-dns on, which will
listen on port UDP 53

### if want setup Swan clusters with 3 managers

``` bash

sudo bin/swan --mesos-master=$MESOS_MASTER_PATH --log-level=debug --raftid=1 --raft-cluster=http://127.0.0.1:2111,http://127.0.0.1:2112,http://127.0.0.1:2113 --data-dir=./data --enable-dns --cluster=0.0.0.0:9999,0.0.0.0:9998,0.0.0.0:9997
sudo bin/swan --mesos-master=$MESOS_MASTER_PATH --log-level=debug --raftid=2 --raft-cluster=http://127.0.0.1:2111,http://127.0.0.1:2112,http://127.0.0.1:2113 --data-dir=./data --enable-dns --cluster=0.0.0.0:9999,0.0.0.0:9998,0.0.0.0:9997
sudo bin/swan --mesos-master=$MESOS_MASTER_PATH --log-level=debug --raftid=3 --raft-cluster=http://127.0.0.1:2111,http://127.0.0.1:2112,http://127.0.0.1:2113 --data-dir=./data --enable-dns --cluster=0.0.0.0:9999,0.0.0.0:9998,0.0.0.0:9997

```

### quick setup with `Goreman`

``` bash

  cp Procfile.exmaple Procfile
  # make sure change the flags within Procfile
  goreman start

```

### feature 1 ip per container require have Macvlan or Linux Bridge
setup on each mesos agent, make sure name the driver as `swan`.

see [macvlan setup](https://github.com/alfredhuang211/study-docker-doc/blob/master/docker%E8%B7%A8%E4%B8%BB%E6%9C%BAmacvlan%E7%BD%91%E7%BB%9C%E9%85%8D%E7%BD%AE.md),

open an issue if you have diffculties setting up a macvlan driver.
