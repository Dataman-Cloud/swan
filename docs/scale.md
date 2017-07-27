#### Scale Spec 

Spec 
```
{
    "instances": 100,
    "ips": ['192.168.1.100', '192.168.1.101', '192.168.1.102'],
}

```

Json Parameters:
+ *instances*(int): The goal to scale up/down.
+ *ips*(array): IP list for static ip(brige or host or scale down ignore).
