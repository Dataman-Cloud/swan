package types

// NetworkSettings exposes the network settings in the api
type NetworkSettings struct {
	Networks map[string]*EndpointSettings
}

// EndpointSettings stores the network endpoint details
type EndpointSettings struct {
	// Configurations
	IPAMConfig *EndpointIPAMConfig
	Links      []string
	Aliases    []string
	// Operational data
	NetworkID           string
	EndpointID          string
	Gateway             string
	IPAddress           string
	IPPrefixLen         int
	IPv6Gateway         string
	GlobalIPv6Address   string
	GlobalIPv6PrefixLen int
	MacAddress          string
	DriverOpts          map[string]string
}

// EndpointIPAMConfig represents IPAM configurations for the endpoint
type EndpointIPAMConfig struct {
	IPv4Address  string   `json:",omitempty"`
	IPv6Address  string   `json:",omitempty"`
	LinkLocalIPs []string `json:",omitempty"`
}

// following definations are borrowed from:
// github.com/docker/engine-api/types/types.go
//

// NetworkResource is the body of the "get network" http response message
type NetworkResource struct {
	Name       string                      // Name is the requested name of the network
	ID         string                      `json:"Id"` // ID uniquely identifies a network on a single machine
	Scope      string                      // Scope describes the level at which the network exists (e.g. `global` for cluster-wide or `local` for machine level)
	Driver     string                      // Driver is the Driver name used to create the network (e.g. `bridge`, `overlay`)
	EnableIPv6 bool                        // EnableIPv6 represents whether to enable IPv6
	IPAM       IPAM                        // IPAM is the network's IP Address Management
	Internal   bool                        // Internal represents if the network is used internal only
	Attachable bool                        // Attachable represents if the global scope is manually attachable by regular containers from workers in swarm mode.
	Containers map[string]EndpointResource // Containers contains endpoints belonging to the network
	Options    map[string]string           // Options holds the network specific options to use for when creating the network
	Labels     map[string]string           // Labels holds metadata specific to the network being created
}

// EndpointResource contains network resources allocated and used for a container in a network
type EndpointResource struct {
	Name        string
	EndpointID  string
	MacAddress  string
	IPv4Address string
	IPv6Address string
}

// IPAM represents IP Address Management
type IPAM struct {
	Driver  string
	Options map[string]string //Per network IPAM driver options
	Config  []IPAMConfig
}

// IPAMConfig represents IPAM configurations
type IPAMConfig struct {
	Subnet     string            `json:",omitempty"`
	IPRange    string            `json:",omitempty"`
	Gateway    string            `json:",omitempty"`
	AuxAddress map[string]string `json:"AuxiliaryAddresses,omitempty"`
}
