#### DeployPolicy

Spec 

```
"deploy": {
    "step": 10,
    "onfailure": "stop"
}
```

Parameters:
+ *step*(int): The number of tasks to run one at a time.
+ *onfailure*(string): The action for failure. Possible values include: 
```
stop
continue
rollback
```

