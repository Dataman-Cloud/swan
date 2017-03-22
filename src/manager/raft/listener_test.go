package raft

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewStoppableListener(t *testing.T) {
	stopC := make(chan struct{})
	ln, err := newStoppableListener(":4232", stopC)
	defer func() {
		ln.Close()
	}()
	assert.Nil(t, err)
	time.AfterFunc(1*time.Second, func() {
		connClient, err := net.Dial("tcp", ":4232")
		assert.Nil(t, err)
		assert.NotNil(t, connClient)
	})
	connServ, err := ln.Accept()
	assert.NotNil(t, connServ)
}

func TestNewStoppableListenerStopC(t *testing.T) {
	//stopC := make(chan struct{})
	//ln, _ := newStoppableListener(":4232", stopC)
	//time.AfterFunc(1*time.Second, func() {
	//stopC <- struct{}{}
	//connClient, err := net.Dial("tcp", ":4232")
	//assert.NotNil(t, err)
	//assert.Nil(t, connClient)
	//})
	//connServ, _ := ln.Accept()
	//assert.Nil(t, connServ)
}
