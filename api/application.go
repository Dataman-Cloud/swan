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

	var version types.Version

	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&version); err != nil {
		return err
	}

	app, err := r.backend.FetchApplication(version.ID)
	if err != nil {
		return err
	}
	if app != nil {
		return errors.New("Applicaiton Id Duplicated")
	}

	user := req.Form.Get("user")
	if user == "" {
		user = "default"
	}

	application := types.Application{
		ID:                version.ID,
		Name:              version.ID,
		Instances:         0,
		UpdatedInstances:  0,
		RunningInstances:  0,
		RollbackInstances: 0,
		UserId:            user,
		ClusterId:         r.backend.ClusterId(),
		Status:            "STAGING",
		Created:           time.Now().Unix(),
		Updated:           time.Now().Unix(),
	}

	if err := r.backend.RegisterApplication(&application); err != nil {
		return err
	}

	if err := r.backend.RegisterApplicationVersion(version.ID, &version); err != nil {
		return err
	}

	if err := r.backend.LaunchApplication(&version); err != nil {
		logrus.Infof("Launch application %s failed with error: %s", version.ID, err.Error())
		return err
	}

	return nil
}

// ListApplication is used to list all applications.
func (r *Router) ListApplications(w http.ResponseWriter, req *http.Request) error {
	apps, err := r.backend.ListApplications()
	if err != nil {
		logrus.Info(err)
	}

	return json.NewEncoder(w).Encode(apps)
}

// FetchApplication is used to fetch a application via applicaiton id.
func (r *Router) FetchApplication(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)

	app, err := r.backend.FetchApplication(vars["appId"])
	if err != nil {
		logrus.Errorf("Fetch application %s failed: %s", vars["appId"], err.Error())
	}

	return json.NewEncoder(w).Encode(app)
}

// DeleteApplication is used to delete a application from mesos and consul via application id.
func (r *Router) DeleteApplication(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)

	if err := r.backend.DeleteApplication(vars["appId"]); err != nil {
		return err
	}

	return nil
}

// ListApplications is used to list all tasks belong to application via application id.
func (r *Router) ListApplicationTasks(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)

	tasks, err := r.backend.ListApplicationTasks(vars["appId"])
	if err != nil {
		return err
	}

	return json.NewEncoder(w).Encode(tasks)
}

// DeleteApplicationTasks is used to delete all tasks belong to application via applicaiton id.
func (r *Router) DeleteApplicationTasks(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)

	if err := r.backend.DeleteApplicationTasks(vars["appId"]); err != nil {
		return err
	}

	return nil
}

// DeleteApplicationTask is used to delete specified task belong to application via application id and task id.
func (r *Router) DeleteApplicationTask(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)

	if err := r.backend.DeleteApplicationTask(vars["appId"], vars["taskId"]); err != nil {
		return err
	}

	return nil
}

// ListApplicationVersions is used to list all versions for a application specified by applicationId.
func (r *Router) ListApplicationVersions(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)

	appVersions, err := r.backend.ListApplicationVersions(vars["appId"])
	if err != nil {
		return err
	}

	return json.NewEncoder(w).Encode(appVersions)
}

// FetchApplicationVersion is used to fetch specified version from consul by version id and application id.
func (r *Router) FetchApplicationVersion(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)

	version, err := r.backend.FetchApplicationVersion(vars["appId"], vars["versionId"])
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

	instances, err := strconv.ParseInt(req.Form.Get("instances"), 10, 64)
	if err != nil {
		return errors.New("instances must be specified in url and can't be null")
	}

	var version types.Version

	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&version); err != nil {
		return err
	}

	vars := mux.Vars(req)

	if err := r.backend.RegisterApplicationVersion(vars["appId"], &version); err != nil {
		return err
	}

	if err := r.backend.UpdateApplication(vars["appId"], instances, &version); err != nil {
		return err
	}

	return nil
}

// ScaleApplication is used to scale application instances.
func (r *Router) ScaleApplication(w http.ResponseWriter, req *http.Request) error {
	if err := req.ParseForm(); err != nil {
		return err
	}

	instances, err := strconv.ParseInt(req.Form.Get("instances"), 10, 64)
	if err != nil {
		return errors.New("instances must be specified in url and can't be null")
	}

	vars := mux.Vars(req)

	if err := r.backend.ScaleApplication(vars["appId"], instances); err != nil {
		return err
	}

	return nil
}

// RollbackApplication rollback application to previous version.
func (r *Router) RollbackApplication(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)

	if err := r.backend.RollbackApplication(vars["appId"]); err != nil {
		return err
	}

	return nil
}
