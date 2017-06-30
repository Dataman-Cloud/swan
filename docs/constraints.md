#### Constraints

##### Spec
```
{
    attribute : "vcluster"
    operator  : "="
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
constraint: [
    {
      attribute   : "vcluster"
      operator    : "=="
      value       : "dataman"
    }
]
```
+ schedule all tasks on agent with attribute "disk:ssd".
```
constraint: [
    {
      attribute   : "disk"
      operator    : "=="
      value       : "ssd"
    }
]
```
+ scheduler all tasks on agent with attribute "vcluster:dataman" and with attribute "kernel.os: centos".
```
constraint: [
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
constraint: [
    {
      attribute : "kernel.name"
      operator  : "=="
      value     : "linux"
    }
]
```
In the future, `operator` will be optional in some cases. eg.:
```
constraint: [
    {
      attribute   : "vcluster"
      value       : "dev"
    }
]
```
```
constraint: [
    {
      attribute   : "kernel.name"
      value       : "linux"
    }
]
```
