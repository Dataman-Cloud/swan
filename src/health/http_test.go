package health

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"os"
	"testing"
)

var (
	baseUrl string
	server  *httptest.Server
)

func TestMain(m *testing.M) {
	server = startHttpServer()
	baseUrl = server.URL
	ret := m.Run()
	server.Close()
	os.Exit(ret)
}

func startHttpServer() *httptest.Server {
	return httptest.NewServer(nil)
}

func TestNewHTTPChecker(t *testing.T) {
	checker := NewHTTPChecker("xxxxx", "x.x.x.x:yyyy", 3, 5, 5, nil, "xxx", "yyy")
	assert.Equal(t, checker.ID, "xxxxx")
}

func TestHTTPCheckerStart(t *testing.T) {
	handler := func(appId, taskId string) error {
		fmt.Println(appId, taskId)
		return nil
	}
	checker := NewHTTPChecker("xxxxx", baseUrl, 1, 2, 2, handler, "xxx", "yyy")
	checker.Start()
}
