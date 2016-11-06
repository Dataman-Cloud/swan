package api

import (
	"github.com/Dataman-Cloud/swan/api/mock"
	"testing"
	"time"
)

func TestListenAndServe(t *testing.T) {
	s := NewServer(&mock.Backend{})
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-quit:
				return
			default:
				s.ListenAndServe("localhost:35008")
			}
		}
	}()

	go func() {
		time.Sleep(2)
		close(quit)
	}()
}
