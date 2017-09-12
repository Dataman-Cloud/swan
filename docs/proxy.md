## Proxy

Spec
```
"proxy": {
      "enabled": false,
      "alias": "www.example.com",
      "listen": 9999,
      "sticky": false
}
```

Json Parameters:
+ *enabled*(optional): whether to enable proxy access.
+ *alias*(optional): the domain name for app access from outside.
+ *listen*(optional): the port listening on swan proxy. through the port you can access application from outside.
+ *sticky*(optional): whether to enable session sticky.
