package scheduler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHandlerManager(t *testing.T) {
	s := &Scheduler{}
	installHandler := func(hm *HandlerManager) {}

	hm := NewHandlerManager(s, installHandler)
	assert.NotNil(t, hm)
}

func TestRegister(t *testing.T) {
	s := &Scheduler{}
	installHandler := func(hm *HandlerManager) {}

	hm := NewHandlerManager(s, installHandler)
	assert.NotNil(t, hm)

	f := func(s *Handler) (*Handler, error) {
		return s, nil
	}
	hm.Register("foobar", f)

}
