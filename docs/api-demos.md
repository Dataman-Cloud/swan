## API Demo

更详细可参考 [examples](https://github.com/Dataman-Cloud/swan/tree/master/api-test)


+ applicaiton deloyment
```
curl -X POST -H "Content-Type: application/json" -d@example/template-replicates.json http://localhost:9999/v_beta/apps
```

+ applications list
```
curl http://localhost:9999/v_beta/apps
```

+ application show
```
curl http://localhost:9999/v_beta/apps/nginx0003-xcm-unnamed
```

+ application delete
```
curl -X DELETE http://localhost:9999/v_beta/apps/nginx0003-xcm-unnamed
```

+ application scale up
```
curl -X PATCH -H "Content-Type: application/json" http://localhost:9999/v_beta/apps/nginx0003-xcm-unnamed/scale-up -d '{"instances": 1}'
```

+ application scale down
```
curl -X PATCH -H "Content-Type: application/json" http://localhost:9999/v_beta/apps/nginx0003-xcm-unamed/scale-down -d '{"instances": 1}'
```

+ application rolling update
```
curl -X PUT -H "Content-Type: application/json"  http://localhost:9999/v_beta/apps/nginx0003-xcm-unnamed -d @example/template-replicates.json
```

+ application rolling update - proceed
```
curl -X PATCH -H "Content-Type: application/json"  http://localhost:9999/v_beta/apps/nginx0003-xcm-unnamed/proceed-update -d '{"instances":1, "weights": {"1": 100,"2":200 }}'
```

+ update weight number for single slot
```
curl -X PATCH -H "Content-Type: application/json"  http://localhost:9999/v_beta/apps/nginx0003-xcm-unnamed/tasks/1/weight -d '{ "weight": 1}'
```

`instances` -1 means updating all instances left, other value means updating the specified instances at one time.

+ application rolling update - cancel
```
curl -X PATCH -H "Content-Type: application/json"  http://localhost:9999/v_beta/apps/nginx0003-xcm-unnamed/cancel-update
```

+ list application versions
```
curl http://localhost:9999/v_beta/apps/nginx0003-xcm-unnamed/versions
```

+ get application version
```
curl
http://localhost:9999/v_beta/apps/nginx0003-xcm-unnamed/versions/14012934223
```
