## health-check

`Swan`支持为同一端口设置多个HealthCheck，
通过portName指定要check的端口。

### protocol

 * TCP 当前端口为tcp端口，
   mesos内部会尝试connect到此端口，以判断服务是否正常
 * HTTP
   需要Check的端口是HTTP服务，Mesos会发起curl请求到此端口，当返回结果为200,201,301,302中之一时认为结果正常

### path

 当协议是HTTP时，表示请求的path地址

### delaySeconds

  health Check开始的最初delay

### inntervalSeconds

  多次health check之间的interval

### timeoutSeconds

  health check的timeout时长

### consecutiveFailures
  
  失败多久以后删除task

### gracePeriodSeconds

  任务启动之后允许的失败时常




