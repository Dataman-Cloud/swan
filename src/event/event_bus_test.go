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

func TestStart(t *testing.T) {
	c, cfun := context.WithCancel(context.Background())
	done := make(chan bool)
	go func() {
		Init()
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
	x := Instance()
	assert.NotNil(t, x)
	listener := &DemoEventListener{}
	AddListener(listener)
}

func TestRemoveListener(t *testing.T) {
	x := Instance()
	assert.NotNil(t, x)
	listener := &DemoEventListener{}
	RemoveListener(listener)
}

func TestStop(t *testing.T) {
	c, _ := context.WithCancel(context.Background())
	done := make(chan bool)
	go func() {
		err := Start(c)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "bye")
		done <- true
	}()

	time.Sleep(time.Second * 1)
	Stop()
	<-done
}

func TestDumpEvent(t *testing.T) {
	c, _ := context.WithCancel(context.Background())
	done := make(chan bool)
	go func() {
		err := Start(c)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "bye")
		done <- true
	}()

	time.Sleep(time.Second * 1)
	listener := &DemoEventListener{}
	AddListener(listener)

	go func() {
		res := <-listener.Pipe
		assert.True(t, res)
	}()
	e := Event{}
	WriteEvent(&e)
	Stop()
	<-done
}
