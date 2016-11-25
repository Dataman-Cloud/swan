package api

import (
	"github.com/Dataman-Cloud/swan/manager/apiserver/router"
)

type Router struct {
	routes  []*router.Route
	backend Backend
}

// NewRouter initializes a new application router.
func NewRouter(b Backend) *Router {
	r := &Router{
		backend: b,
	}

	r.initRoutes()
	return r
}

func (r *Router) Routes() []*router.Route {
	return r.routes
}

func (r *Router) initRoutes() {
	r.routes = []*router.Route{
		router.NewRoute("POST", "/v1/apps", r.BuildApplication),
		router.NewRoute("GET", "/v1/apps", r.ListApplications),
		router.NewRoute("GET", "/v1/apps/{appId}", r.FetchApplication),
		router.NewRoute("DELETE", "/v1/apps/{appId}", r.DeleteApplication),
		router.NewRoute("POST", "/v1/apps/{appId}/update", r.UpdateApplication),
		router.NewRoute("POST", "/v1/apps/{appId}/scale", r.ScaleApplication),
		router.NewRoute("POST", "/v1/apps/{appId}/rollback", r.RollbackApplication),

		router.NewRoute("GET", "/v1/apps/{appId}/tasks", r.ListApplicationTasks),
		router.NewRoute("DELETE", "/v1/apps/{appId}/tasks", r.DeleteApplicationTasks),
		router.NewRoute("DELETE", "/v1/apps/{appId}/tasks/{taskId}", r.DeleteApplicationTask),

		router.NewRoute("GET", "/v1/apps/{appId}/versions", r.ListApplicationVersions),
		router.NewRoute("GET", "/v1/apps/{appId}/versions/{versionId}", r.FetchApplicationVersion),
	}
}
