package api

import (
	"encoding/json"
	"fmt"
	"github.com/Dataman-Cloud/swan/types"
	"github.com/Sirupsen/logrus"
	"net/http"
	"time"
)

func (r *Router) tasksAdd(w http.ResponseWriter, req *http.Request) error {
	defer req.Body.Close()

	var task types.Task

	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&task); err != nil {
		return err
	}

	resources := r.sched.BuildResources(task.Cpus, task.Mem, task.Disk)
	offer, err := r.sched.RequestOffer(resources)
	if err != nil {
		logrus.Errorf("Request offers failed: %s", err.Error())
		return err
	}
	if offer != nil {
		task.ID = fmt.Sprintf("%d", time.Now().UnixNano())
		task.Name = "testapp"
		resp, err := r.sched.LaunchTask(offer, resources, &task)
		if err != nil {
			return err
		}

		if resp != nil && resp.StatusCode != http.StatusAccepted {
			return fmt.Errorf("status code %d received", resp.StatusCode)
		}
		return nil
	}

	return fmt.Errorf("No offers available")
}
