package event

import (
	"testing"

	"github.com/Dataman-Cloud/swan/src/types"
	"github.com/stretchr/testify/assert"
)

func TestNewEvent(t *testing.T) {
	e := NewEvent(EventTypeTaskHealthy, "foo")
	assert.NotNil(t, e)
}

func TestBuildResolverEventErr(t *testing.T) {
	e := NewEvent(EventTypeTaskHealthy, "foo")
	recordGeneratorEvent, err := BuildResolverEvent(e)
	assert.Nil(t, recordGeneratorEvent)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "payload")
}

func TestBuildResolverEventRight(t *testing.T) {
	e := NewEvent(EventTypeTaskHealthy, &types.TaskInfoEvent{})
	recordGeneratorEvent, err := BuildResolverEvent(e)
	assert.NotNil(t, recordGeneratorEvent)
	assert.Nil(t, err)
}

func TestBuildJanitorEventErr(t *testing.T) {
	e := NewEvent(EventTypeTaskHealthy, "foo")
	janitorE, err := BuildJanitorEvent(e)
	assert.Nil(t, janitorE)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "payload")
}

func TestBuildJanitorEventRight(t *testing.T) {
	e := NewEvent(EventTypeTaskHealthy, &types.TaskInfoEvent{})
	janitorE, err := BuildJanitorEvent(e)
	assert.NotNil(t, janitorE)
	assert.Nil(t, err)
}
