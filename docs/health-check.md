## health-check


###  protocol
  which kind of health check let `Mesos` impose on tasks, currently they
are HTTP, TCP and CMD

 * TCP

 For replicates app with `BRIDGE` network mode, `Mesos` try connecting to
randomly assigned port which paired with `containerPort`. For any replicates app
with `HOST` network mode, `port` specified within `portMapping` will be tested. Default IP address tested is `127.0.0.1` which is the network namespace isolated by mesos executor. If any port is open, healthy status for the task will be true, otherwise false.

 * HTTP

  The port been chosen have the same rule as TCP protocol above. Intead
of testing port open, mesos will try `curl` to test if status code of any HTTP response match one of 200,201,301,302. If any healthy status should be true, otherwise false.

 * CMD

  Mesos will try execute an command within any mesos agent that running
the task, if return value is 0, healthy status is true, otherwise false.

### PATH

  if protocol is HTTP, `PATH` is an mandatory field, it means path of
`curl` command. e.g if path is `/status`, mesos agent will try check 
`127.0.0.1:$RANDOM_PORT_OR_PORT_FOR_HOST_MODE/status`.

### VALUE

  if protocol is CMD, `VALUE` is an mandatory field, it means the
command that test healthy status. e.g `ping baidu.com -c 2` or `[ [ -f
/var/run/foobar.pid ] ] && exit 0`

### delaySeconds

  amount of time to wait util starting health checks

### intervalSeconds

  interval between health checks

### timeoutSeconds

  amount of time before health check stop

### consecutiveFailures
  
  umber of consecutive failures until signaling kill task.
  

### gracePeriodSeconds

  Amount of time to allow failed health checks since launch.



## NOTEs

the only protocol that fixed app supported is `CMD` because mesos could
not reach to the IP specifed by fixed app.
