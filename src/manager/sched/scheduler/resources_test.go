package scheduler

import (
	"testing"

	"github.com/Dataman-Cloud/swan/src/manager/sched/mock"

	"github.com/stretchr/testify/assert"
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
	s := NewScheduler(FakeConfig(), &mock.Store{})
	resources := s.BuildResources(0.1, 16, 10)
	assert.Equal(t, *resources[0].Name, "cpus")
	assert.Equal(t, *resources[1].Name, "mem")
	assert.Equal(t, *resources[2].Name, "disk")
}
