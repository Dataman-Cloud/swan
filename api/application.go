package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Dataman-Cloud/swan/types"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

// BuildApplication is used to build a new application.
func (r *Router) BuildApplication(w http.ResponseWriter, req *http.Request) error {
	if err := CheckForJSON(req); err != nil {
		return err
	}

	if err := req.ParseForm(); err != nil {
		return err
	}

	var applicationVersion types.ApplicationVersion

	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&applicationVersion); err != nil {
		return err
	}

	user := req.Form.Get("user")
	if user == "" {
		user = "default"
	}

	application := types.Application{
		ID:              applicationVersion.ID,
		Name:            applicationVersion.ID,
		Instances:       0,
		InstanceUpdated: 0,
		UserId:          user,
		ClusterId:       r.sched.ClusterId,
		Status:          "STAGING",
		Created:         time.Now().Unix(),
		Updated:         time.Now().Unix(),
	}

	if err := r.sched.RegisterApplication(&application); err != nil {
		return err
	}

	if err := r.sched.RegisterApplicationVersion(&applicationVersion); err != nil {
		return err
	}

	if err := r.sched.LaunchApplication(&applicationVersion); err != nil {
		logrus.Infof("Launch application %s failed with error: %s", applicationVersion.ID, err.Error())
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

// ListApplicationVersions is used to list all versions for a application specified by applicationId.
func (r *Router) ListApplicationVersions(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)

	appVersions, err := r.sched.ListApplicationVersions(vars["appId"])
	if err != nil {
		return err
	}

	return json.NewEncoder(w).Encode(appVersions)
}

// FetchApplicationVersion is used to fetch specified version from consul by version id and application id.
func (r *Router) FetchApplicationVersion(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)

	version, err := r.sched.FetchApplicationVersion(vars["appId"], vars["versionId"])
	if err != nil {
		return err
	}

	return json.NewEncoder(w).Encode(version)
}

// UpdateApplication is used to update application version.
func (r *Router) UpdateApplication(w http.ResponseWriter, req *http.Request) error {
	if err := CheckForJSON(req); err != nil {
		return err
	}

	if err := req.ParseForm(); err != nil {
		return err
	}
	inst, err := strconv.Atoi(req.Form.Get("instances"))
	if err != nil {
		return errors.New("instances must be specified in url and can't be null")
	}

	var applicationVersion types.ApplicationVersion

	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&applicationVersion); err != nil {
		return err
	}

	if err := r.sched.RegisterApplicationVersion(&applicationVersion); err != nil {
		return err
	}

	vars := mux.Vars(req)

	if err := r.sched.UpdateApplication(vars["appId"], inst, &applicationVersion); err != nil {
		return err
	}

	return nil
}

// ScaleApplication is used to scale application instances.
func (r *Router) ScaleApplication(w http.ResponseWriter, req *http.Request) error {
	if err := req.ParseForm(); err != nil {
		return err
	}

	instances := req.Form.Get("instances")
	// if instances == "" {
	// 	return errors.New("instances must be specified in url")
	// }
	inst, err := strconv.Atoi(instances)
	if err != nil {
		return errors.New("instances must be specified in url and can't be null")
	}

	vars := mux.Vars(req)

	if err := r.sched.ScaleApplication(vars["appId"], inst); err != nil {
		return err
	}

	return nil
}

// RollbackApplication rollback application to previous version.
func (r *Router) RollbackApplication(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)

	if err := r.sched.RollbackApplication(vars["appId"]); err != nil {
		return err
	}

	return nil
}
