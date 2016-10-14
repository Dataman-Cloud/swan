package api

import (
	"github.com/Dataman-Cloud/swan/scheduler"
)

type Router struct {
	routes []Route
	sched  *scheduler.Scheduler
}

func NewRouter(sched *scheduler.Scheduler) *Router {
	r := &Router{
		sched: sched,
	}
	r.initRoutes()
	return r
}

func (r *Router) initRoutes() {
	r.routes = []Route{
		// task
		//NewRoute("POST", "/tasks", r.tasksAdd),

		// app
		NewRoute("POST", "/v1/apps", r.BuildApplication),
		NewRoute("GET", "/v1/apps", r.ListApplication),
		NewRoute("GET", "/v1/apps/{appId}", r.FetchApplication),
		NewRoute("DELETE", "/v1/apps/{appId}", r.DeleteApplication),
		NewRoute("GET", "/v1/apps/{appId}/tasks", r.ListApplicationTasks),
		NewRoute("DELETE", "/v1/apps/{appId}/tasks", r.DeleteApplicationTasks),
		NewRoute("DELETE", "/v1/apps/{appId}/tasks/{taskId}", r.DeleteApplicationTask),

		// events
		NewRoute("GET", "/v1/events", r.EventStream),
	}
}
