package router

// Router defines an interface to specify a group of routes to add to the api server.
type Router interface {
	Routes() []*Route
}
