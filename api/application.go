package api

import (
	"encoding/json"
	"net/http"

	"github.com/Dataman-Cloud/swan/types"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

func (r *Router) applicationCreate(w http.ResponseWriter, req *http.Request) error {
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

func (r *Router) applicationList(w http.ResponseWriter, req *http.Request) error {
	apps, err := r.sched.ListApplications()
	if err != nil {
		logrus.Info(err)
	}

	return json.NewEncoder(w).Encode(apps)
}

func (r *Router) applicationFetch(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)

	app, err := r.sched.FetchApplication(vars["id"])
	if err != nil {
		logrus.Errorf("Fetch application %s failed: %s", vars["id"], err.Error())
	}

	return json.NewEncoder(w).Encode(app)
}

func (r *Router) applicationDelete(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)

	if err := r.sched.DeleteApplication(vars["id"]); err != nil {
		return err
	}

	return nil
}
