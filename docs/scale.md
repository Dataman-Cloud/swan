#### Scale Policy

Spec 
```
{
    "instances": 100,
    "ips": ['192.168.1.100', '192.168.1.101', '192.168.1.102'],
    "step": 10,
    "onfailure": "continue"
}

```

Json Parameters:
+ *instances*(int): The goal to scale up/down.
+ *ips*(array): IP list for static ip(brige or host or scale down ignore).
+ *step*(int): The number of tasks to run at one time for scale up.
+ *onfailure*(string): The action for failure. Possible values include:
```
stop
continue
```
