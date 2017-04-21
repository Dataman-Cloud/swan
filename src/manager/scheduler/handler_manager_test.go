package scheduler

import (
	"testing"

	"github.com/Dataman-Cloud/swan/src/manager/event"

	"github.com/stretchr/testify/assert"
)

func TestNewHandlerManager(t *testing.T) {
	s := &Scheduler{}

	hm := NewHandlerManager(s)
	assert.NotNil(t, hm)
}

func TestRegister(t *testing.T) {
	s := &Scheduler{}

	hm := NewHandlerManager(s)
	assert.NotNil(t, hm)

	f := func(s *Scheduler, e event.Event) error {
		return nil
	}
	hm.Register("foobar", f)

}
