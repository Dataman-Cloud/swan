package router

import (
	"net/http"
)

type APIFunc func(w http.ResponseWriter, r *http.Request) error

type Route struct {
	method  string
	path    string
	handler APIFunc
}

func (r *Route) Method() string {
	return r.method
}

func (r *Route) Path() string {
	return r.path
}

func (r *Route) Handler() APIFunc {
	return r.handler
}

func NewRoute(method, path string, handler APIFunc) *Route {
	return &Route{
		method,
		path,
		handler,
	}
}
