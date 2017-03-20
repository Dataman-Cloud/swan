package event

import (
	"testing"

	"github.com/Dataman-Cloud/swan/src/types"
	"github.com/stretchr/testify/assert"
)

func TestNewJanitorListener(t *testing.T) {
	l := NewJanitorListener()
	assert.NotNil(t, l)
}

func TestAddJanitorAcceptor(t *testing.T) {
	l := NewJanitorListener()
	ja := types.JanitorAcceptor{ID: "foobar"}

	l.AddAcceptor(ja)
}

func TestRemoveJanitorAcceptor(t *testing.T) {
	l := NewJanitorListener()
	ja := types.JanitorAcceptor{ID: "foobar"}

	l.AddAcceptor(ja)
	l.RemoveAcceptor("foobar")
}

func TestIniterestIn(t *testing.T) {
	e := &Event{ID: "foobar", AppMode: "return"}
	l := NewJanitorListener()

	assert.False(t, l.InterestIn(e))
}

func TestIniterestInHealthTask(t *testing.T) {
	e := &Event{ID: "foobar", AppMode: "replicates", Type: EventTypeTaskHealthy}
	l := NewJanitorListener()

	assert.True(t, l.InterestIn(e))
}

func TestIniterestInUnHealthTask(t *testing.T) {
	e := &Event{ID: "foobar", AppMode: "replicates", Type: EventTypeTaskUnhealthy}
	l := NewJanitorListener()

	assert.True(t, l.InterestIn(e))
}

func TestIniterestInNonHealthEvent(t *testing.T) {
	e := &Event{ID: "foobar", AppMode: "replicates", Type: EventTypeAppStateScaleDown}
	l := NewJanitorListener()

	assert.False(t, l.InterestIn(e))
}
