package main

import (
	"errors"
	"log"
	"net"
	"os"
	"testing"

	check "gopkg.in/check.v1"
)

func init() {
	s, err := newApiSuite()
	if err != nil {
		log.Fatal(err)
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
