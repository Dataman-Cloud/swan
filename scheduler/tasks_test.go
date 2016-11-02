package scheduler

import (
	"github.com/Dataman-Cloud/swan/mesosproto/mesos"
	"github.com/Dataman-Cloud/swan/scheduler/mock"
	"github.com/Dataman-Cloud/swan/types"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBuildTask(t *testing.T) {
	offer := &mesos.Offer{
		Id: &mesos.OfferID{
			Value: proto.String("abcdefghigklmn"),
		},
		FrameworkId: nil,
		AgentId: &mesos.AgentID{
			Value: proto.String("xxxxxx"),
		},
		Hostname: proto.String("x.x.x.x"),
	}

	version := &types.ApplicationVersion{
		ID:        "test",
		Command:   nil,
		Cpus:      0.1,
		Mem:       16,
		Disk:      0,
		Instances: 5,
		Container: &types.Container{
			Type: "DOCKER",
			Docker: &types.Docker{
				ForcePullImage: proto.Bool(false),
				Image:          proto.String("nginx:1.10"),
				Network:        "BRIDGE",
				Parameters: &[]types.Parameter{
					{
						Key:   "xxxx",
						Value: "yyyy",
					},
				},
				PortMappings: &[]types.PortMapping{
					{
						ContainerPort: 8080,
						Name:          "web",
						Protocol:      "http",
					},
				},
				Privileged: proto.Bool(true),
			},
			Volumes: nil,
		},
		Labels: &map[string]string{
			"USER_ID": "xxxxx"},
		HealthChecks: []*types.HealthCheck{
			{
				ID:        "xxxxx",
				Address:   "x.x.x.x",
				TaskID:    "aaaa",
				AppID:     "bbbb",
				Port:      nil,
				PortIndex: nil,
				Command:   nil,
				Path:      nil,
			},
		},
		Env: nil,
		KillPolicy: &types.KillPolicy{
			Duration: int64(5),
		},
		UpdatePolicy: nil,
	}

	sched := NewScheduler("x.x.x.x:yyyy", nil, &mock.Store{}, "xxxx", nil, nil)
	task, _ := sched.BuildTask(offer, version, "a.b.c.d")
	assert.Equal(t, task.Name, "a.b.c.d")
}

func TestBuildTaskInfo(t *testing.T) {
	offer := &mesos.Offer{
		Id: &mesos.OfferID{
			Value: proto.String("abcdefghigklmn"),
		},
		FrameworkId: nil,
		AgentId: &mesos.AgentID{
			Value: proto.String("xxxxxx"),
		},
		Hostname: proto.String("x.x.x.x"),
	}

	offer.Resources = append(offer.Resources, &mesos.Resource{
		Name: proto.String("ports"),
		Type: mesos.Value_RANGES.Enum(),
		Ranges: &mesos.Value_Ranges{
			Range: []*mesos.Value_Range{
				{
					Begin: proto.Uint64(uint64(1000)),
					End:   proto.Uint64(uint64(1001)),
				},
			},
		},
	})

	resources := []*mesos.Resource{
		&mesos.Resource{
			Name:   proto.String("cpus"),
			Type:   mesos.Value_SCALAR.Enum(),
			Scalar: &mesos.Value_Scalar{Value: proto.Float64(0.1)},
		},
		&mesos.Resource{
			Name:   proto.String("mem"),
			Type:   mesos.Value_SCALAR.Enum(),
			Scalar: &mesos.Value_Scalar{Value: proto.Float64(16)},
		},
		&mesos.Resource{
			Name:   proto.String("disk"),
			Type:   mesos.Value_SCALAR.Enum(),
			Scalar: &mesos.Value_Scalar{Value: proto.Float64(10)},
		},
		&mesos.Resource{
			Name: proto.String("ports"),
			Type: mesos.Value_RANGES.Enum(),
			Ranges: &mesos.Value_Ranges{
				Range: []*mesos.Value_Range{
					{
						Begin: proto.Uint64(uint64(1000)),
						End:   proto.Uint64(uint64(1001)),
					},
				},
			},
		},
	}

	task := &types.Task{
		ID:      "xxxxx",
		Name:    "a.b.c.d",
		Command: nil,
		Cpus:    float64(0.1),
		Mem:     float64(16),
		Disk:    float64(10),
		Image:   proto.String("nginx:1.10"),
		Network: "BRIDGE",
		PortMappings: []*types.PortMappings{
			{
				Port:     8080,
				Protocol: "http",
			},
		},
		Privileged: proto.Bool(false),
		Parameters: []*types.Parameter{
			{
				Key:   "xxxx",
				Value: "yyyy",
			},
		},
		ForcePullImage: proto.Bool(false),
		Volumes: []*types.Volume{
			{
				ContainerPath: "/tmp/xxxx",
				HostPath:      "/tmp/xxxx",
				Mode:          "rw",
			},
		},
		Env:    map[string]string{"DB": "xxxxxxx"},
		Labels: &map[string]string{"USER": "yyyy"},
		HealthChecks: []*types.HealthCheck{
			{
				ID:        "xxxxx",
				Address:   "x.x.x.x",
				TaskID:    "aaaa",
				AppID:     "bbbb",
				Port:      nil,
				PortIndex: nil,
				Command:   nil,
				Path:      nil,
			},
		},
		KillPolicy: &types.KillPolicy{
			Duration: int64(5),
		},
		OfferId:       proto.String("xxxyyyzzz"),
		AgentId:       proto.String("mmmnnnzzz"),
		AgentHostname: proto.String("x.x.x.x"),
		Status:        "RUNNING",
		AppId:         "testapp",
	}

	s := NewScheduler("x.x.x.x:yyyy", nil, &mock.Store{}, "xxxx", nil, nil)
	taskInfo := s.BuildTaskInfo(offer, resources, task)

	assert.Equal(t, *taskInfo.Container.Docker.Image, "nginx:1.10")
}

func TestLaunchTasks(t *testing.T) {
	offer := &mesos.Offer{
		Id: &mesos.OfferID{
			Value: proto.String("abcdefghigklmn"),
		},
		FrameworkId: nil,
		AgentId: &mesos.AgentID{
			Value: proto.String("xxxxxx"),
		},
		Hostname: proto.String("x.x.x.x"),
	}

	offer.Resources = append(offer.Resources, &mesos.Resource{
		Name: proto.String("ports"),
		Type: mesos.Value_RANGES.Enum(),
		Ranges: &mesos.Value_Ranges{
			Range: []*mesos.Value_Range{
				{
					Begin: proto.Uint64(uint64(1000)),
					End:   proto.Uint64(uint64(1001)),
				},
			},
		},
	})

	resources := []*mesos.Resource{
		&mesos.Resource{
			Name:   proto.String("cpus"),
			Type:   mesos.Value_SCALAR.Enum(),
			Scalar: &mesos.Value_Scalar{Value: proto.Float64(0.1)},
		},
		&mesos.Resource{
			Name:   proto.String("mem"),
			Type:   mesos.Value_SCALAR.Enum(),
			Scalar: &mesos.Value_Scalar{Value: proto.Float64(16)},
		},
		&mesos.Resource{
			Name:   proto.String("disk"),
			Type:   mesos.Value_SCALAR.Enum(),
			Scalar: &mesos.Value_Scalar{Value: proto.Float64(10)},
		},
		&mesos.Resource{
			Name: proto.String("ports"),
			Type: mesos.Value_RANGES.Enum(),
			Ranges: &mesos.Value_Ranges{
				Range: []*mesos.Value_Range{
					{
						Begin: proto.Uint64(uint64(1000)),
						End:   proto.Uint64(uint64(1001)),
					},
				},
			},
		},
	}

	task := &types.Task{
		ID:      "xxxxx",
		Name:    "a.b.c.d",
		Command: nil,
		Cpus:    float64(0.1),
		Mem:     float64(16),
		Disk:    float64(10),
		Image:   proto.String("nginx:1.10"),
		Network: "BRIDGE",
		PortMappings: []*types.PortMappings{
			{
				Port:     8080,
				Protocol: "http",
			},
		},
		Privileged: proto.Bool(false),
		Parameters: []*types.Parameter{
			{
				Key:   "xxxx",
				Value: "yyyy",
			},
		},
		ForcePullImage: proto.Bool(false),
		Volumes: []*types.Volume{
			{
				ContainerPath: "/tmp/xxxx",
				HostPath:      "/tmp/xxxx",
				Mode:          "rw",
			},
		},
		Env:    map[string]string{"DB": "xxxxxxx"},
		Labels: &map[string]string{"USER": "yyyy"},
		HealthChecks: []*types.HealthCheck{
			{
				ID:        "xxxxx",
				Address:   "x.x.x.x",
				TaskID:    "aaaa",
				AppID:     "bbbb",
				Port:      nil,
				PortIndex: nil,
				Command:   nil,
				Path:      nil,
			},
		},
		KillPolicy: &types.KillPolicy{
			Duration: int64(5),
		},
		OfferId:       proto.String("xxxyyyzzz"),
		AgentId:       proto.String("mmmnnnzzz"),
		AgentHostname: proto.String("x.x.x.x"),
		Status:        "RUNNING",
		AppId:         "testapp",
	}

	s := NewScheduler("x.x.x.x:yyyy", nil, &mock.Store{}, "xxxx", nil, nil)
	taskInfo := s.BuildTaskInfo(offer, resources, task)

	var tasks []*mesos.TaskInfo
	_, err := s.LaunchTasks(offer, append(tasks, taskInfo))
	assert.NotNil(t, err)
}

func TestKillTask(t *testing.T) {
	task := &types.Task{
		ID:      "xxxxx",
		Name:    "a.b.c.d",
		Command: nil,
		Cpus:    float64(0.1),
		Mem:     float64(16),
		Disk:    float64(10),
		Image:   proto.String("nginx:1.10"),
		Network: "BRIDGE",
		PortMappings: []*types.PortMappings{
			{
				Port:     8080,
				Protocol: "http",
			},
		},
		Privileged: proto.Bool(false),
		Parameters: []*types.Parameter{
			{
				Key:   "xxxx",
				Value: "yyyy",
			},
		},
		ForcePullImage: proto.Bool(false),
		Volumes: []*types.Volume{
			{
				ContainerPath: "/tmp/xxxx",
				HostPath:      "/tmp/xxxx",
				Mode:          "rw",
			},
		},
		Env:    map[string]string{"DB": "xxxxxxx"},
		Labels: &map[string]string{"USER": "yyyy"},
		HealthChecks: []*types.HealthCheck{
			{
				ID:        "xxxxx",
				Address:   "x.x.x.x",
				TaskID:    "aaaa",
				AppID:     "bbbb",
				Port:      nil,
				PortIndex: nil,
				Command:   nil,
				Path:      nil,
			},
		},
		KillPolicy: &types.KillPolicy{
			Duration: int64(5),
		},
		OfferId:       proto.String("xxxyyyzzz"),
		AgentId:       proto.String("mmmnnnzzz"),
		AgentHostname: proto.String("x.x.x.x"),
		Status:        "RUNNING",
		AppId:         "testapp",
	}
	s := NewScheduler("x.x.x.x:yyyy", nil, &mock.Store{}, "xxxx", nil, nil)

	_, err := s.KillTask(task)
	assert.NotNil(t, err)
}
