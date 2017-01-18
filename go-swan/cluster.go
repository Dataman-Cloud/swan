package swan

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

// swanCluster is a collection of swan nodes
type swanCluster struct {
	sync.RWMutex
	// a collection of managers
	managers []*manager
	// the http client
	client       *http.Client
	managerIndex int
}

// manager represents an individual endpoint
type manager struct {
	// the name / ip address of the host
	endpoint string
}

// newSwanCluster returns a new swanCluster
func newSwanCluster(client *http.Client, swanURL string) (*swanCluster, error) {
	// step: extract and basic validate the endpoints
	var managers []*manager
	var defaultProto string

	for _, endpoint := range strings.Split(swanURL, ",") {
		// step: check for nothing
		if endpoint == "" {
			return nil, errors.New("endpoint is blank")
		}
		// step: parse the url
		u, err := url.Parse(endpoint)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("endpoint: %s is invalid reason: %s", endpoint, err))
		}
		// step: set the default protocol schema
		if defaultProto == "" {
			if u.Scheme != "http" && u.Scheme != "https" {
				return nil, errors.New(fmt.Sprintf("endpoint: %s protocol must be (http|https)", endpoint))
			}
			defaultProto = u.Scheme
		}
		// step: does the url have a protocol schema? if not, use the default
		if u.Scheme == "" || u.Opaque != "" {
			urlWithScheme := fmt.Sprintf("%s://%s", defaultProto, u.String())
			if u, err = url.Parse(urlWithScheme); err != nil {
				panic(fmt.Sprintf("unexpected parsing error for URL '%s' with added default scheme: %s", urlWithScheme, err))
			}
		}

		// step: check for empty hosts
		if u.Host == "" {
			return nil, errors.New(fmt.Sprintf("endpoint: %s must have a host", endpoint))
		}

		// step: create a new node for this endpoint
		managers = append(managers, &manager{endpoint: u.String()})
	}

	return &swanCluster{
		client:       client,
		managers:     managers,
		managerIndex: 0,
	}, nil
}

// retrieve the next manager
func (c *swanCluster) getNextManager() (*manager, error) {
	c.RLock()
	defer c.RUnlock()
	if c.managerIndex >= c.size() {
		return nil, ErrSwanDown
	} else {
		manager := c.managers[c.managerIndex]
		c.managerIndex = c.managerIndex + 1
		return manager, nil
	}
}

func (c *swanCluster) resetManagerIndex() {
	c.RLock()
	defer c.RUnlock()
	c.managerIndex = 0
	return
}

// managerList returns a list of managers
func (c *swanCluster) managersList() []string {
	c.RLock()
	defer c.RUnlock()
	var list []string
	for _, m := range c.managers {
		list = append(list, m.endpoint)
	}
	return list
}

// size returns the size of the swanCluster
func (c *swanCluster) size() int {
	return len(c.managers)
}
