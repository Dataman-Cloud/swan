package scheduler

import (
	"github.com/Dataman-Cloud/swan/mesosproto/mesos"
	"github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"
)

func createScalarResource(name string, value float64) *mesos.Resource {
	return &mesos.Resource{
		Name:   &name,
		Type:   mesos.Value_SCALAR.Enum(),
		Scalar: &mesos.Value_Scalar{Value: &value},
	}
}

func createRangeResource(name string, begin, end uint64) *mesos.Resource {
	return &mesos.Resource{
		Name: &name,
		Type: mesos.Value_RANGES.Enum(),
		Ranges: &mesos.Value_Ranges{
			Range: []*mesos.Value_Range{
				{
					Begin: proto.Uint64(begin),
					End:   proto.Uint64(end),
				},
			},
		},
	}
}

func (s *Scheduler) BuildResources(cpus, mem, disk float64) []*mesos.Resource {
	logrus.WithFields(logrus.Fields{"cpus": cpus, "mem": mem, "disk": disk}).Info("Building resources...")
	var resources = []*mesos.Resource{}

	if cpus > 0 {
		resources = append(resources, createScalarResource("cpus", cpus))
	}

	if mem > 0 {
		resources = append(resources, createScalarResource("mem", mem))
	}

	if disk > 0 {
		resources = append(resources, createScalarResource("disk", disk))
	}

	return resources
}
