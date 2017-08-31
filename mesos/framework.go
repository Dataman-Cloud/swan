package mesos

import (
	"os"

	"github.com/golang/protobuf/proto"

	"github.com/Dataman-Cloud/swan/mesosproto"
)

const DefaultFrameworkFailoverTimeout = 7 * 24 * 60 * 60

var (
	defaultFrameworkUser            = "root"
	defaultFrameworkName            = "swan"
	defaultFrameworkPrincipal       = "swan"
	defaultFrameworkFailoverTimeout = float64(DefaultFrameworkFailoverTimeout)
)

func (s *Scheduler) buildFramework() *mesosproto.FrameworkInfo {
	hostName, err := os.Hostname()
	if err != nil {
		hostName = "UNKNOWN"
	}

	fw := &mesosproto.FrameworkInfo{
		// ID:              proto.String(""), // reset later
		User:            proto.String(defaultFrameworkUser),
		Name:            proto.String(defaultFrameworkName),
		Principal:       proto.String(defaultFrameworkPrincipal),
		FailoverTimeout: proto.Float64(defaultFrameworkFailoverTimeout),
		Checkpoint:      proto.Bool(s.cfg.EnableCheckPoint),
		Hostname:        proto.String(hostName),
		Capabilities: []*mesosproto.FrameworkInfo_Capability{
			{Type: mesosproto.FrameworkInfo_Capability_PARTITION_AWARE.Enum()},
			{Type: mesosproto.FrameworkInfo_Capability_GPU_RESOURCES.Enum()},
		},
	}

	if s.cfg.EnableCapabilityKilling {
		fw.Capabilities = append(fw.Capabilities, &mesosproto.FrameworkInfo_Capability{
			Type: mesosproto.FrameworkInfo_Capability_TASK_KILLING_STATE.Enum(),
		})
	}

	return fw
}
