package main

import (
	"github.com/Dataman-Cloud/swan/types"
)

func newVersion(name string, instances int) *types.Version {
	ver := new(types.Version)
	ver.Name = name
	ver.Instances = int32(instances)
	ver.Command = ""
	ver.CPUs = 0.01
	ver.Mem = 5
	ver.Disk = 0
	ver.RunAs = "integration"
	ver.Priority = 0
	ver.Constraints = nil
	ver.Container = &types.Container{
		Type: "docker",
		Docker: &types.Docker{
			Image:          "nginx",
			Network:        "bridge",
			Parameters:     nil,
			ForcePullImage: false,
			Privileged:     true,
			PortMappings: []*types.PortMapping{
				{
					Name:          "web",
					Protocol:      "tcp",
					ContainerPort: 80,
					HostPort:      80,
				},
			},
		},
	}

	ver.DeployPolicy = &types.DeployPolicy{
		Step:      2,
		OnFailure: "continue",
	}

	return ver
}
