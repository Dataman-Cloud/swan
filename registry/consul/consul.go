package consul

import (
	consul "github.com/hashicorp/consul/api"
)

type Consul struct {
	client *consul.Client
}

func NewConsul(addr string) (*Consul, error) {
	cfg := consul.Config{
		Address: addr,
		Scheme:  "http",
	}

	client, err := consul.NewClient(&cfg)
	if err != nil {
		return nil, err
	}

	return &Consul{
		client: client,
	}, nil
}
