package api

import (
	"encoding/json"
	"net/http"

	"github.com/Dataman-Cloud/swan/types"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

// BuildApplication is used to build a new application.
func (r *Router) BuildApplication(w http.ResponseWriter, req *http.Request) error {
	if err := CheckForJSON(req); err != nil {
		return err
	}

	var application types.Application

	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&application); err != nil {
		return err
	}

	if err := r.sched.LaunchApplication(&application); err != nil {
		logrus.Infof("Launch application %s failed with error: %s", application.ID, err.Error())
		return err
	}

	return nil
}

// ListApplication is used to list all applications.
func (r *Router) ListApplication(w http.ResponseWriter, req *http.Request) error {
	apps, err := r.sched.ListApplications()
	if err != nil {
		logrus.Info(err)
	}

	return json.NewEncoder(w).Encode(apps)
}

// FetchApplication is used to fetch a application via applicaiton id.
func (r *Router) FetchApplication(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)

	app, err := r.sched.FetchApplication(vars["appId"])
	if err != nil {
		logrus.Errorf("Fetch application %s failed: %s", vars["appId"], err.Error())
	}

	return json.NewEncoder(w).Encode(app)
}

// DeleteApplication is used to delete a application from mesos and consul via application id.
func (r *Router) DeleteApplication(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)

	if err := r.sched.DeleteApplication(vars["appId"]); err != nil {
		return err
	}

	return nil
}

// ListApplications is used to list all tasks belong to application via application id.
func (r *Router) ListApplicationTasks(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)

	tasks, err := r.sched.ListApplicationTasks(vars["appId"])
	if err != nil {
		return err
	}

	return json.NewEncoder(w).Encode(tasks)
}

// DeleteApplicationTasks is used to delete all tasks belong to application via applicaiton id.
func (r *Router) DeleteApplicationTasks(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)

	if err := r.sched.DeleteApplicationTasks(vars["appId"]); err != nil {
		return err
	}

	return nil
}

// DeleteApplicationTask is used to delete specified task belong to application via application id and task id.
func (r *Router) DeleteApplicationTask(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)

	if err := r.sched.DeleteApplicationTask(vars["appId"], vars["taskId"]); err != nil {
		return err
	}

	return nil
}
