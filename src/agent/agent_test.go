package agent

import (
	"testing"
	"time"

	"github.com/Dataman-Cloud/swan/src/config"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestNew(t *testing.T) {
	a, err := New("foobar", config.AgentConfig{})
	assert.NotNil(t, a)
	assert.Nil(t, err)
}

func TestJoinAndStart(t *testing.T) {
	a, _ := New("foobar", config.AgentConfig{
		ListenAddr: "0.0.0.0:8765",
		DNS: config.DNS{
			Domain: "foobar.com",
		},
		Janitor: config.Janitor{
			ListenAddr: "0.0.0.0:8764",
		},
	})
	assert.NotNil(t, a)

	done := make(chan bool)

	ctx, cfun := context.WithCancel(context.Background())
	go func() {
		e := a.JoinAndStart(ctx)
		assert.NotNil(t, e)
		done <- true
	}()

	time.Sleep(time.Second)
	cfun()

	<-done
}

func TestJoinToCluster(t *testing.T) {
	a, _ := New("foobar", config.AgentConfig{
		ListenAddr: "0.0.0.0:8765",
		DNS: config.DNS{
			Domain: "foobar.com",
		},
		Janitor: config.Janitor{
			ListenAddr: "0.0.0.0:8764",
		},
	})
	assert.NotNil(t, a)
	done := make(chan bool)

	ctx, cfun := context.WithCancel(context.Background())
	go func() {
		e := a.JoinToCluster(ctx)
		assert.NotNil(t, e)
		done <- true
	}()

	time.Sleep(time.Second)
	cfun()

	<-done
}
