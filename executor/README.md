
## Refers
  - http://mesos.apache.org/documentation/latest/executor-http-api/
  - http://mesos.apache.org/documentation/latest/app-framework-development-guide/

## Mesos Slave Injected Envs to Executor
```liquid
LIBPROCESS_IP=192.168.1.130
MESOS_AGENT_ENDPOINT=192.168.1.130:5051
MESOS_DIRECTORY=/data/mesos/slaves/75a6f244-5f12-4ae7-b474-94c500a3d98f-S1/frameworks/132b4f53-fbf3-4a7f-93c9-a87d49268ff5-0000/executors/12345/runs/2dcfd257-f67c-4a91-a897-793fd068a0f7
MESOS_EXECUTOR_ID=12345
MESOS_EXECUTOR_SHUTDOWN_GRACE_PERIOD=5secs
MESOS_NATIVE_JAVA_LIBRARY=/usr/lib/libmesos-1.2.0.so
MESOS_NATIVE_LIBRARY=/usr/lib/libmesos-1.2.0.so
MESOS_HTTP_COMMAND_EXECUTOR=0
MESOS_SLAVE_PID=slave(1)@192.168.1.130:5051
MESOS_FRAMEWORK_ID=132b4f53-fbf3-4a7f-93c9-a87d49268ff5-0000
MESOS_CHECKPOINT=0
SHLVL=2
LIBPROCESS_PORT=0
MESOS_SLAVE_ID=75a6f244-5f12-4ae7-b474-94c500a3d98f-S1
MESOS_SANDBOX=/data/mesos/slaves/75a6f244-5f12-4ae7-b474-94c500a3d98f-S1/frameworks/132b4f53-fbf3-4a7f-93c9-a87d49268ff5-0000/executors/12345/runs/2dcfd257-f67c-4a91-a897-793fd068a0f7
```
