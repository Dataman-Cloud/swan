package mesos

import (
	"os"

	"github.com/golang/protobuf/proto"

	"github.com/Dataman-Cloud/swan/proto/mesos"
)

const DefaultFrameworkFailoverTimeout = 7 * 24 * 60 * 60

var (
	defaultFrameworkUser            = "root"
	defaultFrameworkName            = "swan"
	defaultFrameworkPrincipal       = "swan"
	defaultFrameworkFailoverTimeout = float64(DefaultFrameworkFailoverTimeout)
	defaultFrameworkCheckpoint      = false
)

func defaultFramework() *mesos.FrameworkInfo {
	hostName, err := os.Hostname()
	if err != nil {
		hostName = "UNKNOWN"
	}

	return &mesos.FrameworkInfo{
		// ID:              proto.String(""), // reset later
		User:            proto.String(defaultFrameworkUser),
		Name:            proto.String(defaultFrameworkName),
		Principal:       proto.String(defaultFrameworkPrincipal),
		FailoverTimeout: proto.Float64(defaultFrameworkFailoverTimeout),
		Checkpoint:      proto.Bool(defaultFrameworkCheckpoint),
		Hostname:        proto.String(hostName),
		Capabilities: []*mesos.FrameworkInfo_Capability{
			{Type: mesos.FrameworkInfo_Capability_PARTITION_AWARE.Enum()},
			{Type: mesos.FrameworkInfo_Capability_TASK_KILLING_STATE.Enum()},
		},
	}
}
