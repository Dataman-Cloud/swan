#### DeployPolicy

Spec 

```
"deploy": {
    "step": 10,
    "onfailure": "stop"
}
```

Parameters:
+ *step*(int): The number of tasks to run at one time.
+ *onfailure*(string): The action for failure. Possible values include: 
```
stop
continue
```

