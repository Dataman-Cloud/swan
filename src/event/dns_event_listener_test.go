package event

import (
	"testing"

	"github.com/Dataman-Cloud/swan/src/types"
	"github.com/stretchr/testify/assert"
)

func TestNewDNSListener(t *testing.T) {
	l := NewDNSListener()
	assert.NotNil(t, l)
}

func TestAddDNSAcceptor(t *testing.T) {
	l := NewDNSListener()
	ja := types.ResolverAcceptor{ID: "foobar"}

	l.AddAcceptor(ja)
}

func TestRemoveDNSAcceptor(t *testing.T) {
	l := NewDNSListener()
	ja := types.ResolverAcceptor{ID: "foobar"}

	l.AddAcceptor(ja)
	l.RemoveAcceptor("foobar")
}

func TestDNSIniterestIn(t *testing.T) {
	e := &Event{ID: "foobar", AppMode: "return"}
	l := NewDNSListener()

	assert.False(t, l.InterestIn(e))
}

func TestDNSIniterestInHealthTask(t *testing.T) {
	e := &Event{ID: "foobar", AppMode: "replicates", Type: EventTypeTaskHealthy}
	l := NewDNSListener()

	assert.True(t, l.InterestIn(e))
}

func TestDNSIniterestInUnHealthTask(t *testing.T) {
	e := &Event{ID: "foobar", AppMode: "replicates", Type: EventTypeTaskUnhealthy}
	l := NewDNSListener()

	assert.True(t, l.InterestIn(e))
}

func TestDNSIniterestInNonHealthEvent(t *testing.T) {
	e := &Event{ID: "foobar", AppMode: "replicates", Type: EventTypeAppStateScaleDown}
	l := NewDNSListener()

	assert.False(t, l.InterestIn(e))
}
