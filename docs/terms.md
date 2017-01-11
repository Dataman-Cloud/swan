### Task
等同一个Mesos
Task，根据用户提供的应用模板构造出Mesos要执行的Task描述，Task状态为Mesos反馈的状态，Task在Mesos里消失之后会存入Task历史中。

### Slot
`Swan`运行时概念，Slot运行时最多有一个Running的Task，应用创建或者Scale-Up时先创建出Pending-Offer的Slot，得到Offer之后构造TaskInfo，之后把TaskInfo Dispatch到Mesos; 对于Fixed类型应用，每个Slot都有唯一的IP地址；
Slot的Index为从0开始顺序增大的数字，可通过slotIndex.app.user.cluster定位一个唯一的Slot，比如0.appname.username.clustername。
Scale-Up过程为从最后一个Slot依次增加过程，Scale-Down为从最后一个Slot依次减少过程，rolling-update为从第0个slot依次更新到新version过程。


### App
标识一个运行中应用概念， 一个应用由多个Slot构成，一个应用同时可以有两个version, proposedVersion和currentVersion 。寻址一个应用办法为appname.username.clustername。


### Version
版本为当前应用使用的描述模板，定义了应用的健康监测策略，容器，应用类型，IP等。 应用Normal状态下CurrentVersion代表当前Slot发Task所使用的Version; 当Rolling—Update开始时，更新的版本为proposedVersion, proceed update完成以后currentVersion放入version history当中， currentVersion指向proposedVersion; proposedVersion变为空。 

### Proxy
这里Proxy指七层HTTP代理，
`Swan`内部实现了简单的HTTP代理功能，目的是配合服务发现将用户的应用暴露出去，更多的比如分流，限流，灰度发布等功能依赖此功能。类似请求如
0.app.user.cluster.domain.com的请求到Proxy上之后，
Proxy会根据Header当中的Host将请求分发到后面不同的Slot之上；
对于请求如app.user.cluster.domain.com，Proxy会将负载以一定的算法分发到不同的Slot之上，达到负载均衡目的。

### DNS Proxy
`Swan` 内部实现了的DNS Proxy功能，为每一个Replicates 类型Slot增加了一条SRV记录，
对于Fixed类型应用， 每个Slot都有一个A记录对应;
同时，domain的A记录IP地址为Proxy的IP。

### 一容器一IP
根据大部分客户需求，需要为容器指定一个Vlan的IP地址，`Swan`允许发布应用时指定当前应用为fixed类型应用，并且提供数量足够的唯一可访问的IP地址。
`Swan`在Task下发过程中会将Mesos
Task为Task指定好Ip，并且使用`用户自定义`类型的docker driver,
driver固定为`Swan`；
如果想使用此功能，需要事先建立起名称为`swan`的docker
driver，这里我们推荐使用macvlan的driver，后面会有详细介绍如何搭建一个macvlan的driver。（理论上Swan对docker network driver没有固定需求，Linux Bridge同样可行，如需部署Linux Bridge可联系我们）

### App Mode
由于`Swan`兼容的主要两种应用类型区别很大，这里我们将应用分为主要两大类，并区别对待

* fixed  此类应用指代对弹性需求不大， 偏四层应用，使用一容器一IP，
  DNS增加了A记录， 不支持健康检查和端口指定。
* replicates 此类应用对弹性依赖较大，偏七层应用，不用一容器一IP， Docker
  Bridge Driver，DNS有对应的人SRV记录。

### Manager

主要作用为维护分布式状态机，比如应用状态和Slot状态。多Manager之间通过Raft协议做Log
Replication和Leader
Election。对状态机的修改只会发生在Manager中的Leader上，再同步到Follower上。 推荐生产环境部署为3个Manager。

### Agent

做Proxy和DNS Proxy之用。







