package types

import "time"

type Meta struct {
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Annotations struct {
	Name   string
	Labels map[string]string
}
