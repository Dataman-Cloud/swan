## Trouble Shooting

反馈问题时请带着如下信息：

### 版本和Commit

请求API： `GET /version`

### 调度器运行时
请求API：`GET /v1/debug/dump`

### Golang Runtime Stack
假如调度器容器为 `dataman-swan-master`，执行命令：
```bash
docker exec dataman-swan-master kill -USR1 1
docker cp dataman-swan-master:/tmp/swan-stack-trace.log  .
```
然后将 `swan-stack-trace.log`一并提交

### App／Tasks／Versions
如果是某个App的问题，提供这个App的数据如下：

  - 请求API：`GET /v1/apps/{app_id}`
  - 请求API：`GET /v1/apps/{app_id}/tasks`
  - 请求API：`GET /v1/apps/{app_id}/versions`
