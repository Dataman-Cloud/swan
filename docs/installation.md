### Installation

clone Swan source code from github.com:
```
git clone git@github.com:Dataman-Cloud/swan.git
```
Then you can compile Swan with:
```
make
```

### Run

Single
```
./bin/swan manager --mesos=zk://192.168.1.92:2181/mesos \
                   --zk=zk://192.168.1.92:2181/swan \
                   --listen=192.168.1.193:9999
```

Cluster
```
./bin/swan manager --mesos=zk://192.168.1.92:2181/mesos \
                   --zk=zk://192.168.1.92:2181/swan \
                   --listen=192.168.1.193:9997
```
```
./bin/swan manager --mesos=zk://192.168.1.92:2181/mesos \
                   --zk=zk://192.168.1.92:2181/swan \
                   --listen=192.168.1.193:9998
```
```
./bin/swan manager --mesos=zk://192.168.1.92:2181/mesos \
                   --zk=zk://192.168.1.92:2181/swan \
                   --listen=192.168.1.193:9999
```

```
--mesos : mesos address.
--zk    : zk address.
--listen: listen address.
```

More comand line flags, see `./bin/swan --help`.

