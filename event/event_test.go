package event

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEvent(t *testing.T) {
	e := NewEvent(EventTypeTaskHealthy, "foo")
	assert.NotNil(t, e)
}
