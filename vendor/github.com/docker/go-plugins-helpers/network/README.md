# Docker network extension API

Go handler to create external network extensions for Docker.

## Usage

This library is designed to be integrated in your program.

1. Implement the `network.Driver` interface.
2. Initialize a `network.Handler` with your implementation.
3. Call either `ServeTCP` or `ServeUnix` from the `network.Handler`.

### Example using TCP sockets:

```go
  import "github.com/docker/go-plugins-helpers/network"

  d := MyNetworkDriver{}
  h := network.NewHandler(d)
  h.ServeTCP("test_network", ":8080")
```

### Example using Unix sockets:

```go
  import "github.com/docker/go-plugins-helpers/network"

  d := MyNetworkDriver{}
  h := network.NewHandler(d)
  h.ServeUnix("test_network", 0)
```

## Full example plugins

- [docker-ovs-plugin](https://github.com/gopher-net/docker-ovs-plugin) - An Open vSwitch Networking Plugin
