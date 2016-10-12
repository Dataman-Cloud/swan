package api

import (
	"encoding/json"
	"github.com/Dataman-Cloud/swan/types"
	"github.com/Sirupsen/logrus"
	"net/http"
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
