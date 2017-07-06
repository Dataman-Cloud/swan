package api

import (
	"net/http"
)

type HandlerFunc func(w http.ResponseWriter, r *http.Request)

type Route struct {
	method  string
	path    string
	handler HandlerFunc
	prefix  bool // flag on prefix route matcher or not
}

func (r *Route) Methods() []string {
	if r.method == "ANY" {
		return []string{"GET", "PUT", "POST", "PATCH", "DELETE"}
	}
	return []string{r.method}
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
		false,
	}
}

func NewPrefixRoute(method, path string, handler HandlerFunc) *Route {
	return &Route{
		method,
		path,
		handler,
		true,
	}
}
