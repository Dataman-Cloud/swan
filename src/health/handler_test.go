package health

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Dataman-Cloud/swan/src/health/mock"
	"github.com/Dataman-Cloud/swan/src/types"
)

func TestHealthCheckFailedHandler(t *testing.T) {
	msgQueue := make(chan types.ReschedulerMsg, 1)
	go func() {
		msg := <-msgQueue
		msg.Err <- nil
	}()

	m := NewHealthCheckManager(&mock.Store{}, msgQueue)
	err := m.HealthCheckFailedHandler("xxxxx", "yyyyy")
	assert.Nil(t, err)
}
