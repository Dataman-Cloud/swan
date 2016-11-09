package api

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestListenAndServe(t *testing.T) {
	srv := httptest.NewServer(nil)
	srv.Close()
	s := NewServer(srv.URL)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-quit:
				return
			default:
				s.ListenAndServe()
			}
		}
	}()

	go func() {
		time.Sleep(2)
		close(quit)
	}()
}
