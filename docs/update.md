#### Update Policy

Spec
```
"update": {
    "delay": 5,
    "onfailure": "continue"
}
```

Json Parameters:
+ *delay*(int): The delay between two updates.
+ *onfailure*(string): The action on failure. Possible values are:
```
stop

continue

rollback (not supported yet) 
```
