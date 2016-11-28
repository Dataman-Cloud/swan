package health

import (
	"github.com/Dataman-Cloud/swan/health/mock"
	"github.com/Dataman-Cloud/swan/types"
	"github.com/stretchr/testify/assert"
	"testing"
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
