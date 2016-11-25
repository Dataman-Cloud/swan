package scheduler

import (
	"github.com/Dataman-Cloud/swan/scheduler/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateScaleResource(t *testing.T) {
	resource := createScalarResource("cpu", 0.1)
	assert.Equal(t, *resource.Name, "cpu")
	assert.Equal(t, *resource.Scalar.Value, 0.1)
}

func TestCreateRangeResource(t *testing.T) {
	resource := createRangeResource("ports", 1000, 10001)
	assert.Equal(t, *resource.Name, "ports")
}

func TestBuildResource(t *testing.T) {
	sched := NewScheduler("x.x.x.x:yyyy", nil, &mock.Store{}, "xxxx", nil, nil, nil)
	resources := sched.BuildResources(0.1, 16, 10)
	assert.Equal(t, *resources[0].Name, "cpus")
	assert.Equal(t, *resources[1].Name, "mem")
	assert.Equal(t, *resources[2].Name, "disk")
}
