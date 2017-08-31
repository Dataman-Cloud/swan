package types

import (
	"time"
)

type VCluster struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	Nodes   []*Node   `json:"nodes"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

type Node struct {
	ID    string            `json:"id"`
	IP    string            `json:"ip"`
	Attrs map[string]string `json:"attrs"`
}

type CreateVClusterBody struct {
	Name string `json:"name"`
}
