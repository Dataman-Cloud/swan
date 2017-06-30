#### Port Mapping

Spec:
```
    {
        "name": "web",
        "protocol": "tcp",
        "containerPort": 80,
        "hostPort": 80
    }
```

Example:

*bridge* - network is bridge, service port is 8080. 
```
{
    "name": web,
    "protocol": "tcp",
    "containerPort": 8080,
    "hostPort": 0, // will be ignored
}
```

*host* - network is host, service port is 8080.
```
{
    "name": web,
    "protocol": "tcp",
    "containerPort": 0, // will be ignored
    "hostPort": 8080, 
}
```
