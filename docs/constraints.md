## API Demo

## Constraints

`Constraints` is a main feature we want support for now. Without
`Constraints` it is not possible to dispatch a set of tasks to
desired Mesos agents.

## Example Constraints supported

  ```
      UNIQUE hostname
  ```

  ```
      UNIQUE agentid
  ```

  ```
      LIKE hostname "ssd-machine*"
  ```

  ```
      LIKE ip "192.168*"
  ```

  ```
      NOT( LIKE ip "192.168*" )
  ```

  ```
      AND ( LIKE ip "192.168*" ) (UNIQUE hostname)
  ```

  ```
      OR ( LIKE ip "192.168*" ) (UNIQUE hostname)
  ```

  ```
      OR ( AND ( LIKE ip "192.168*" ) ( UNIQUE agentid) ) (UNIQUE hostname)
  ```



## Mesos Agent attributes

Also `Swan` allow to use `Mesos` attributes to filter desired agent,
before doing that Mesos agent should started with attributes added, for
example:

```
  ./mesos-agent.sh --attribues="label:ssd;mem-intensive:true"
```

we can dispatch to these host with

```
  LIKE label ssd
```



