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
		NewRoute("PATCH", "/v1/apps/{app_id}", r.scaleApp),
		NewRoute("PUT", "/v1/apps/{app_id}", r.updateApp),
		NewRoute("POST", "/v1/apps/{app_id}", r.rollback),
		NewRoute("PATCH", "/v1/apps/{app_id}/weights", r.updateWeights),

		NewRoute("GET", "/v1/apps/{app_id}/tasks", r.getTasks),
		NewRoute("GET", "/v1/apps/{app_id}/tasks/{task_id}", r.getTask),
		NewRoute("DELETE", "/v1/apps/{app_id}/tasks/{task_id}", r.deleteTask),
		NewRoute("DELETE", "/v1/apps/{app_id}/tasks", r.deleteTasks),
		NewRoute("PUT", "/v1/apps/{app_id}/tasks/{task_id}", r.updateTask),
		NewRoute("POST", "/v1/apps/{app_id}/tasks/{task_id}", r.rollbackTask),
		NewRoute("PATCH", "/v1/apps/{app_id}/tasks/{task_id}/weight", r.updateWeight),

		NewRoute("GET", "/v1/apps/{app_id}/versions", r.getVersions),
		NewRoute("GET", "/v1/apps/{app_id}/versions/{version_id}", r.getVersion),
		NewRoute("POST", "/v1/apps/{app_id}/versions", r.createVersion),

		NewRoute("POST", "/v1/apps/{app_id}/gray", r.grayPublish),

		NewRoute("POST", "/v1/compose", r.newCompose),
		NewRoute("POST", "/v1/compose/parse", r.parseYAML),
		NewRoute("GET", "/v1/compose", r.listComposes),
		NewRoute("GET", "/v1/compose/{compose_id}", r.getCompose),
		NewRoute("DELETE", "/v1/compose/{compose_id}", r.deleteCompose),

		NewRoute("GET", "/ping", r.ping),
		NewRoute("GET", "/v1/events", r.events),
		NewRoute("GET", "/v1/stats", r.stats),
		NewRoute("GET", "/version", r.version),
		NewRoute("GET", "/v1/leader", r.getLeader),
		NewRoute("DELETE", "/v1/purge", r.purge),
	}
}
