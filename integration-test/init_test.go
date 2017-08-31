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

	check.Suite(s) // register the test suit
}

// make fit with go test mechanism
func Test(t *testing.T) {
	// run all test cases
	// check.TestingT(t)

	// run regexp matched test cases
	result := check.RunAll(&check.RunConf{
		Filter: os.Getenv("TESTON"),
	})
	if !result.Passed() {
		log.Fatal(result.String())
	}
	log.Println(result.String())
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
