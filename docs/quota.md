## Quota

`Swan` 支持基于应用优先级的多层级的资源 Quota 管理，目前包括 `HardLimit`，`OverCapacity` 和 `FreeOfCharge` 三层。

### 用例

 * CRUD Quota & Quota Status
 * 提供 API plugin 控制检查资源申请，如果超过 Quota 限制，则返回 `403`
 * 提供 controller 定期同步 Quota Status （未实现）

### 数据模型

```go
// v2
// map<string,int64> resourceList
type ResourceQuotaSpec struct {
    Limit map[string]int64 `protobuf:"bytes,1,rep,name=limit" json:"limit,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
}

type ResourceQuotaStatus struct {
    //    map<string,int64> limit = 1;
    Offered map[string]int64 `protobuf:"bytes,2,rep,name=offered" json:"offered,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
}

type TierResourceQuota struct {
    QuotaSpec   *ResourceQuotaSpec   `protobuf:"bytes,1,opt,name=quotaSpec" json:"quotaSpec,omitempty"`
    QuotaStatus *ResourceQuotaStatus `protobuf:"bytes,2,opt,name=quotaStatus" json:"quotaStatus,omitempty"`
}

type ResourceQuota struct {
    QuotaGroup string                        `protobuf:"bytes,1,opt,name=quotaGroup,proto3" json:"quotaGroup,omitempty"`
    Quotas     map[string]*TierResourceQuota `protobuf:"bytes,2,rep,name=quotas" json:"quotas,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value"`
}
```

样例：

```
{
    "quotaGroup": "default",
    "quotas": {
        "hardLimit": {
            "quotaSpec": {
                "limit": {
                    "cpu": 2000,
                    "mem": 100
                }
            }
        },
        "overCapacity": {
            "quotaSpec": {
                "limit": {
                    "cpu": 1000,
                    "mem": 50
                }
            }
        }
    }
}
```

### 资源定义

目前支持 CPU，Mem，比较方便扩展其他资源定义，比如 硬盘资源，甚至包括 应用数，实例数。
  
注：其中 CPU 从 Mesos 的 float(0.01) 通过乘 1000 转化为 int(10)，方便和其他资源定义统一数据格式; Mem 默认单位为 M.

### 设计

目前的 Quota 分为两部分内容，一部分是 Quota 定义（`ResourceQuotaSpec`），指定当前资源组下各个资源层的各个资源的限制；
一部分是 Quota 状态（`ResourceQuotaStatus`），用来统计当前资源组在各个资源层的所有资源的申请量。

资源层，包括 `HardLimit`，`OverCapacity` 和 `FreeOfCharge` 三层，应用配置中有优先级，根据优先级数值对应到不同的资源层上。

具体的规则为：

* 200+  `HardLimit`，该层资源原则上对应实际的物资资源，属于硬限制；
* 0-199 `OverCapacity`，该层资源根据经验规划，属于"超卖"部分；
* -1    `FreeOfCharge`，免费资源，但是优先级最低；

#### REST API

对外提供 Quota 的 REST API，支持创建／更新／删除／获取 Quota 列表／获取指定资源组的 Quota

#### API plugin

实现 API 级别的 plugin，类似 middleware，拦截检查需要的 API 进行 Quota 分析，判断是否超过 Quota 限制，如果超过，返回 `403` 错误；
如果未超过，更新对应资源组的 Quota Status。

#### Quota Controller（未实现）

计划实现一个 `Controller` 来定期监控实际资源和 Quota Status 是否一致，不一致需要更新 Quota Status，
最主要的原因就是，像 `删除应用` 这样的操作，属于异步操作，对资源的回收只能通过 `Controller` 来完成。

### 更多信息

参见 swagger 中的 API 文档，或直接阅读代码。
