package api

import (
	"net/http"
)

type HandlerFunc func(w http.ResponseWriter, r *http.Request)

type Route struct {
	method  string
	path    string
	handler HandlerFunc
}

func (r *Route) Method() string {
	return r.method
}

func (r *Route) Path() string {
	return r.path
}

func (r *Route) Handler() HandlerFunc {
	return r.handler
}

func NewRoute(method, path string, handler HandlerFunc) *Route {
	return &Route{
		method,
		path,
		handler,
	}
}
