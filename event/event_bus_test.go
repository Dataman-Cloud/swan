package event

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

type DemoEventListener struct {
	Pipe chan bool
}

func (d *DemoEventListener) Write(e *Event) error {
	return nil
}

func (d *DemoEventListener) InterestIn(e *Event) bool {
	return true
}

func (d *DemoEventListener) Key() string {
	return "demo"
}

func (d *DemoEventListener) Wait() {
	<-d.Pipe
}

func TestStart(t *testing.T) {
	c, cfun := context.WithCancel(context.Background())
	done := make(chan bool)
	go func() {
		err := Start(c)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "context canceled")
		done <- true
	}()

	time.Sleep(time.Second * 1)
	cfun()
	<-done
}

func TestAddListener(t *testing.T) {
	x := eventBusInstance
	assert.NotNil(t, x)
	listener := &DemoEventListener{}
	AddListener(listener)
}

func TestRemoveListener(t *testing.T) {
	x := eventBusInstance
	assert.NotNil(t, x)
	listener := &DemoEventListener{}
	RemoveListener(listener)
}
