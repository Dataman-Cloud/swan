package health

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Dataman-Cloud/swan/src/health/mock"
	"github.com/Dataman-Cloud/swan/src/types"
)

func TestHealthCheckManager(t *testing.T) {
	m := NewHealthCheckManager(&mock.Store{}, make(chan types.ReschedulerMsg))
	m.Init()
	go func() {
		time.Sleep(1)
		m.Stop()
	}()
	m.Start()
}

func TestHealthCHeckManagerAdd(t *testing.T) {
	m := NewHealthCheckManager(&mock.Store{}, make(chan types.ReschedulerMsg))

	httpCheck := types.Check{
		ID:       "xxxxxxxx",
		Address:  "y.y.y.y",
		Port:     8080,
		TaskID:   "mmmm",
		AppID:    "zzzzzz",
		Protocol: "http",
		Interval: 5,
		Timeout:  5,
	}

	m.Add(&httpCheck)

	tcpCheck := types.Check{
		ID:       "xxxxxxxx",
		Address:  "y.y.y.y",
		Port:     8080,
		TaskID:   "mmmm",
		AppID:    "zzzzzz",
		Protocol: "tcp",
		Interval: 5,
		Timeout:  5,
	}

	m.Add(&tcpCheck)
}

func TestHealthCheckNext(t *testing.T) {
	m := NewHealthCheckManager(&mock.Store{}, make(chan types.ReschedulerMsg))
	assert.Nil(t, m.Next())
}

func TestHealthCheckStopCheck(t *testing.T) {
	m := NewHealthCheckManager(&mock.Store{}, make(chan types.ReschedulerMsg))
	check := types.Check{
		ID:       "xxxxx",
		Address:  "y.y.y.y",
		Port:     8080,
		TaskID:   "mmmm",
		AppID:    "zzzzzz",
		Protocol: "http",
		Interval: 5,
		Timeout:  5,
	}

	m.Add(&check)
	m.StopCheck("mmmm")
}

func TestHasCheck(t *testing.T) {
	m := NewHealthCheckManager(&mock.Store{}, make(chan types.ReschedulerMsg))
	assert.Equal(t, m.HasCheck("xxxxxxx"), false)
}
