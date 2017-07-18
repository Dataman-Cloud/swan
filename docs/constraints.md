#### Constraints

##### Spec
```
{
    attribute : "vcluster"
    operator  : "=="
    value     : "dataman"
}
```
+ *attribute*(string) - Specifies the name of attribute setting on mesos agent. the attribute must be set on mesos agent.

+ *operator*(string) - Specifies the comparison operator. Possible values include:
```
==
!=
~=
```
+ *value*(string) - Specifies the value to compare the attribute against using the specified operation.

##### Examples
+ schedule all tasks on agent with attribute "vcluster:dataman".
```
constraints: [
    {
      attribute   : "vcluster"
      operator    : "=="
      value       : "dataman"
    }
]
```
+ schedule all tasks on agent with attribute "disk:ssd".
```
constraints: [
    {
      attribute   : "disk"
      operator    : "=="
      value       : "ssd"
    }
]
```
+ scheduler all tasks on agent with attribute "vcluster:dataman" and with attribute "kernel.os: centos".
```
constraints: [
    {
      attribute   : "vcluster"
      operator    : "=="
      value       : "dataman"
    },
    {
      attribute   : "kernel.os"
      operator    : "=="
      value       : "centos"
    },
]
```
+ scheduler all tasks on linux box.
```
constraints: [
    {
      attribute : "kernel.name"
      operator  : "=="
      value     : "linux"
    }
]
```
In the future, `operator` will be optional in some cases. eg.:
```
constraints: [
    {
      attribute   : "vcluster"
      value       : "dev"
    }
]
```
```
constraints: [
    {
      attribute   : "kernel.name"
      value       : "linux"
    }
]
```
