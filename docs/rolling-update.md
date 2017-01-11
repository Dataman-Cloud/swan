## rolling update

rolling update定位于用户需要有新版本的更新时发布新版本，
可修改的内容包括容器镜像，容器环境变量设置，label设置以及其他容器配置；抑或任务资源配置，比如CPU,
Mem, Disk等信息。不推荐的配置修改包括constraint，restart policy, update
policy等内容。


## 更新一个新版本

```
  http put localhost:9999/v_beta/apps/$APPID @new_version.json
```
此时第0个Slot被更新至ProposedVersion，其余不变

## Proceed Rolling Update过程
```
  http patch localhost:9999/v_beta/apps/$APPID/proceed instances:=$NUM
```
NUM应小于未更新数量，当NUM=未更新数量时更新结束,
CurrentVersion加入到History当中，ProposedVersion变成CurrentVersion;
此时从最小未更新slot之后NUM个slot被更新


## Cancel Rolling Update过程

```
  http patch localhost:9999/v_beta/apps/$APPID/cancel instances:=$NUM
```

已更新到ProposedVersion恢复至CurrentVersion

## 内部实现
一个App包括关于滚动更新数据结构有

* ProposedVersion 目标更新版本
* CurrentVersion 当前版本
* VersionHistory 版本历史

