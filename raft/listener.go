package raft

import (
	"errors"
	"net"
	"time"
)

// stoppabelListener sets TCP keep-alive timeout on accepted
// connections and waits on stopC message
type stoppableListener struct {
	*net.TCPListener
	stopC <-chan struct{}
}

func newStoppableListener(addr string, stopC <-chan struct{}) (*stoppableListener, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &stoppableListener{ln.(*net.TCPListener), stopC}, nil
}

func (ln stoppableListener) Accept() (net.Conn, error) {
	connC := make(chan *net.TCPConn, 1)
	errC := make(chan error, 1)

	go func() {
		tc, err := ln.AcceptTCP()
		if err != nil {
			errC <- err
			return
		}

		connC <- tc
	}()

	select {
	case <-ln.stopC:
		return nil, errors.New("server stopped")
	case err := <-errC:
		return nil, err
	case tc := <-connC:
		tc.SetKeepAlive(true)
		tc.SetKeepAlivePeriod(3 * time.Second)
		return tc, nil
	}
}
