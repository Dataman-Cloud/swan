# IP Address Manager
The IPAM  for Swan is supposed to manage lifecycle of a predefined group
of IPs, all IP addresses should be avaliable within the same layer 2
subnet as well as the hosts, all of them are reserved for containers,
each container could be assigned a unique ip, with underlaying Macvlan
bridge created by docker.

## This is not a docker plugin
The initial thought would be make this as a docker plugin like DHCP
IPAM, so docker daemon could reach the IPAM remotely to where it stay in the Swan
managers. But the truth is that as we have our own scheduler by default so this
IPAM was not intend to run in standlone mode without scheduler, as though the only consumer of
the IPAM would be the schdueler itself, it might be better choice make
the IPAM not a plugin but part of scheduler which can access both from
HTTP API and call directly.

## How to initialize the IP list pool
IP list pool supposed to be entered mannuly through HTTP API, which
each ip should be unique and accessible within the same layer 2 subnet.

## Lifecycle of a ip

  * `avaliable` avaliable to be allocate to a container
  * `reserved` reserved, should not allocated to any container
  * `allocated` currently used by a container
  * `releasing` released but not avaliable soon, will be turn into
    avaliable state after certain time periods

## How to interact with IPAM, the APIs

  * `list` avaliable ips, no matter what state they are
  * `initialize` the ip pool
  * `empty` the ip pool
  * `allocate` the ip from pool
  * `release` a ip back to the pool


