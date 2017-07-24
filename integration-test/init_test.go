package main

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"testing"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/check.v1"
)

func init() {
	s, err := newApiSuite()
	if err != nil {
		log.Fatalln(err)
	}

	check.Suite(s)
}

func Test(t *testing.T) {
	check.TestingT(t)
}

type ApiSuite struct {
	SwanHost string
}

func newApiSuite() (*ApiSuite, error) {
	swanHost := os.Getenv("SWAN_HOST")

	if swanHost == "" {
		return nil, errors.New("env SWAN_HOST required")
	}

	_, _, err := net.SplitHostPort(swanHost)
	if err != nil {
		return nil, err
	}

	return &ApiSuite{
		SwanHost: swanHost,
	}, nil
}

func (s *ApiSuite) sendRequest(method, uri string, data io.Reader) (code int, body []byte, err error) {
	req, err := http.NewRequest(method, "http://"+s.SwanHost+uri, data)
	if err != nil {
		return -1, nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return -1, nil, err
	}
	defer resp.Body.Close()

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return -1, nil, err
	}

	return resp.StatusCode, bs, nil
}

func (s *ApiSuite) bind(data []byte, val interface{}) error {
	return json.Unmarshal(data, &val)
}
