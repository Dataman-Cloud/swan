## DNS

replicates mode application

```
dig @localhost -p $DNS_PORT 0.app.user.cluster.swan.com SRV
```

fixed mode application

```
dig @localhost -p $DNS_PORT 0.app.user.cluster.swan.com A
```
