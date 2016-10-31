package api

import ()

type Router struct {
	routes  []Route
	backend Backend
}

func NewRouter(backend Backend) *Router {
	r := &Router{
		backend: backend,
	}
	r.initRoutes()
	return r
}

func (r *Router) initRoutes() {
	r.routes = []Route{
		// app
		NewRoute("POST", "/v1/apps", r.BuildApplication),
		NewRoute("GET", "/v1/apps", r.ListApplications),
		NewRoute("GET", "/v1/apps/{appId}", r.FetchApplication),
		NewRoute("DELETE", "/v1/apps/{appId}", r.DeleteApplication),
		NewRoute("POST", "/v1/apps/{appId}/update", r.UpdateApplication),
		NewRoute("POST", "/v1/apps/{appId}/scale", r.ScaleApplication),
		NewRoute("POST", "/v1/apps/{appId}/rollback", r.RollbackApplication),

		NewRoute("GET", "/v1/apps/{appId}/tasks", r.ListApplicationTasks),
		NewRoute("DELETE", "/v1/apps/{appId}/tasks", r.DeleteApplicationTasks),
		NewRoute("DELETE", "/v1/apps/{appId}/tasks/{taskId}", r.DeleteApplicationTask),

		NewRoute("GET", "/v1/apps/{appId}/versions", r.ListApplicationVersions),
		NewRoute("GET", "/v1/apps/{appId}/versions/{versionId}", r.FetchApplicationVersion),
	}
}
