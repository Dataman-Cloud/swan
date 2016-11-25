package apiserver

import (
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestListenAndServe(t *testing.T) {
	srv := httptest.NewServer(nil)
	srv.Close()
	s := NewServer(srv.URL, "test.sock")
	defer os.Remove("test.sock")
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
