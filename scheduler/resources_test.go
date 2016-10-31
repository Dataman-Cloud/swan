package scheduler

import (
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
