- [环境要求](#requirement)
- [Executor](#executor)
- [KVM API](#kvm)
  + [创建KVM App](#create)  
  + [列出KVM App](#list)
  + [列出KVM App Tasks](#list-tasks)
  + [查看KVM App](#get)
  + [删除KVM App](#remove)
  + [启动KVM App](#start)
  + [停止KVM App](#stop)
  + [挂起KVM App](#suspend)
  + [恢复KVM App](#resume)


## requirement

### 宿主机
```bash
# install base packages
yum -y install qemu-kvm qemu-img qemu-kvm-tools libvirt libvirt-daemon-kvm virt-top virt-install virt-what

# optional, require GTK+ to show vm graphical console
yum -y install virt-manager-common virt-manager virt-viewer 

# ensure the host support vm
virt-host-validate

# enable & start libvirtd
systemctl enable libvirtd
systemctl start  libvirtd

# optional
chmod 777 /var/lib/libvirt/qemu
```

### mesos-slave:

```liquid
Dockerfile:
RUN yum -y install qemu-kvm qemu-img qemu-kvm-tools libvirt virt-top virt-install

Start:
        -v /etc/libvirt:/etc/libvirt/ \
        -v /var/lib/libvirt:/var/lib/libvirt \
        -v /var/run/libvirt:/var/run/libvirt \
        -v /var/log/libvirt:/var/log/libvirt \
        -v /data/iso:/data/iso
```

## executor

### 构建：
```bash
cd path-to-swan-src/executor/
make
```

### 参数：
```bash
swan-executor kvm			// KVM调度
swan-executor pod			// POD调度(尚未支持)
swan-executor example		// 开发示例
```

### 使用
结合下面创建KVM App API的`executor`字段使用

## kvm 

### create
`POST` `/v1/kvm-apps?name=demo&runas=bbk&cluster=bj`

Request:  

Content-Type: `application/json`
```json
{
  "count": 3,  // instance count
  "cpus": 1,
  "mems": 500, // by MiB
  "disks": 10, // by GiB
  "image": {
    "type": "iso",
    "uri": "CentOS-7-x86_64-Minimal-1708.iso"
  },
  "vnc": {
    "password": "password"
  },
  "executor": {
    "url": "http://192.168.0.104:81/swan-executor",
    "command": "./swan-executor kvm"
  },
  "constraints": [
    {
      "attribute": "hostname",
      "operator": "~=",
      "value": "107"
    }
  ]
}
```

Response
```json
201: Created
{
    "id": "demo.bbk.bj"
}
```


### list
`GET` `/v1/kvm-apps` 

Response:  
```json

// TaskStatus
// Creating: IsoFetching, ImageCreating, XmlCreating, KvmDefining, KvmStarting, 
// Stopping: KvmStopping, KvmStopped, KvmStopFailed, 
// Starting: KvmStarting, KvmStartFailed, KvmRunning
// Suspending:  KvmSuspending, KvmSuspendFailed, KvmSuspended
// Resuming: KvmResuming, KvmResumeFailed, KvmRunning

[
  {
    "id": "demo.bbk.bj",
    "name": "demo",
    "runAs": "bbk",
    "cluster": "bj",
    "desc": "",
    "taskCount": 3,
    "tasksStatus": {       
      "KvmStarting": 1,
      "TASK_RUNNING": 1,
      "pending": 1
    },
    "config": {
      "count": 3,
      "cpus": 1,
      "mems": 500,
      "disks": 10,
      "executor": {
        "url": "http://192.168.0.104:81/swan-executor",
        "command": "./swan-executor kvm"
      },
      "image": {
        "type": "iso",
        "uri": "CentOS-7-x86_64-Minimal-1708.iso"
      },
      "vnc": {
        "enabled": false,
        "password": "password"
      },
      "killPolicy": null,
      "constraints": [
        {
          "attribute": "hostname",
          "operator": "~=",
          "value": "107"
        }
      ]
    },
    "operationStatus": "creating",
    "errmsg": "",
    "createdAt": "2017-09-18T14:41:14.441288453Z",
    "updatedAt": "2017-09-18T14:41:14.441288522Z"
  }
]
```

### list tasks
`GET` `/v1/kvm-apps/demo.bbk.bj`

Response:
```json
[
  {
    "id": "0c43d85a2071.1.demo.bbk.bj",
    "name": "1.demo.bbk.bj",
    "agentId": "c6b39bbe-a21f-4bc2-802b-80d1003abaf1-S1",
    "executorId": "1c4e1494858174cd710fef44270973d9",
    "domainUuid": "",
    "domainName": "",
    "opstatus": "",
    "status": "TASK_RUNNING",
    "errmsg": "",
    "createdAt": "2017-09-18T14:41:14.44901215Z",
    "updatedAt": "2017-09-18T14:41:14.449012257Z",
    "ipAddr": "192.168.0.107",     // ip address
    "vncAddr": "192.168.0.107:1"   // vnc address for vnc graphic viewer
  },
  {
    "id": "5211be224df9.2.demo.bbk.bj",
    "name": "2.demo.bbk.bj",
    "agentId": "c6b39bbe-a21f-4bc2-802b-80d1003abaf1-S1",
    "executorId": "91c4091b79cf2494dddb07d479e701a6",
    "domainUuid": "",
    "domainName": "",
    "opstatus": "",
    "status": "TASK_RUNNING",
    "errmsg": "",
    "createdAt": "2017-09-18T14:41:14.451082166Z",
    "updatedAt": "2017-09-18T14:41:14.451082249Z",
    "ipAddr": "192.168.0.107",
    "vncAddr": "192.168.0.107:0"
  },
  {
    "id": "d3fb5fba23c5.0.demo.bbk.bj",
    "name": "0.demo.bbk.bj",
    "agentId": "c6b39bbe-a21f-4bc2-802b-80d1003abaf1-S1",
    "executorId": "6cbf8263986bf4134c2dc1878fc2d7c9",
    "domainUuid": "",
    "domainName": "",
    "opstatus": "",
    "status": "TASK_RUNNING",
    "errmsg": "",
    "createdAt": "2017-09-18T14:41:14.445218506Z",
    "updatedAt": "2017-09-18T14:41:14.445218601Z",
    "ipAddr": "192.168.0.107",
    "vncAddr": "192.168.0.107:2"
  }
]
```

### get
`GET` `/v1/kvm-apps/demo.bbk.bj`

```json
see above
```

### remove
`DELETE` `/v1/kvm-apps/demo.bbk.bj`

### start
`POST` `/v1/kvm-apps/demo.bbk.bj/start`

### stop
`POST` `/v1/kvm-apps/demo.bbk.bj/stop`

### suspend
`POST` `/v1/kvm-apps/demo.bbk.bj/suspend`

### resume
`POST` `/v1/kvm-apps/demo.bbk.bj/resume`
