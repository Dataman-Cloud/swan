package api

import (
	. "github.com/Dataman-Cloud/swan/store"
)

type Router struct {
	routes []*Route
	driver Driver
	db     Store
}

func NewRouter(d Driver, s Store) *Router {
	r := &Router{
		driver: d,
		db:     s,
	}

	r.setupRoutes()

	return r
}

func (r *Router) Routes() []*Route {
	return r.routes
}

func (r *Router) setupRoutes() {
	r.routes = []*Route{
		NewRoute("GET", "/v1/apps", r.listApps),
		NewRoute("POST", "/v1/apps", r.createApp),
		NewRoute("GET", "/v1/apps/{app_id}", r.getApp),
		NewRoute("DELETE", "/v1/apps/{app_id}", r.deleteApp),
		NewRoute("PATCH", "/v1/apps/{app_id}/scale", r.scale),
		NewRoute("POST", "/v1/apps/{app_id}/scale/stop", r.stopScale),
		NewRoute("PUT", "/v1/apps/{app_id}/update", r.updateApp),
		NewRoute("POST", "/v1/apps/{app_id}/update/stop", r.stopUpdate),
		NewRoute("PUT", "/v1/apps/{app_id}/rollback", r.rollback),
		NewRoute("PATCH", "/v1/apps/{app_id}/weights", r.updateWeights),
		NewRoute("GET", "/v1/apps/{app_id}/tasks", r.getTasks),
		NewRoute("GET", "/v1/apps/{app_id}/tasks/{task_id}", r.getTask),
		NewRoute("DELETE", "/v1/apps/{app_id}/tasks/{task_id}", r.deleteTask),
		NewRoute("PUT", "/v1/apps/{app_id}/tasks/{task_id}/update", r.updateTask),
		NewRoute("PUT", "/v1/apps/{app_id}/tasks/{task_id}/rollback", r.rollbackTask),
		NewRoute("PATCH", "/v1/apps/{app_id}/tasks/{task_id}/weight", r.updateWeight),
		NewRoute("GET", "/v1/apps/{app_id}/versions", r.getVersions),
		NewRoute("GET", "/v1/apps/{app_id}/versions/{version_id}", r.getVersion),

		NewRoute("POST", "/v1/compose", r.newCompose),
		NewRoute("POST", "/v1/compose/parse", r.parseYAML),
		NewRoute("GET", "/v1/compose", r.listComposes),
		NewRoute("GET", "/v1/compose/{iid}", r.getCompose),
		NewRoute("DELETE", "/v1/compose/{iid}", r.deleteCompose),

		NewRoute("GET", "/ping", r.ping),
		NewRoute("GET", "/v1/events", r.events),
		NewRoute("GET", "/v1/stats", r.stats),
		NewRoute("GET", "/version", r.version),
		NewRoute("GET", "/leader", r.leader),
	}
}
