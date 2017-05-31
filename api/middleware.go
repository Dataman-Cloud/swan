package api

import ()

type Middleware interface {
	Name() string

	WrapHandler(HandlerFunc) HandlerFunc
}
