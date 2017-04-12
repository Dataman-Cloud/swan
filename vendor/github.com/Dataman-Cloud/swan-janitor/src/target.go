package janitor

import (
	"fmt"
	"net/url"

	log "github.com/Sirupsen/logrus"
)

type Target struct {
	AppID    string `json:"appID"`
	TaskID   string `json:"taskID"`
	TaskIP   string `json:"taskIP"`
	TaskPort uint32 `json:"taskPort"`
	PortName string `json:"portName"`

	Weight float64 `json:"weihgt"`
}

func (t *Target) Equal(t1 *Target) bool {
	return t.AppID == t1.AppID &&
		t.TaskID == t1.TaskID &&
		t.TaskIP == t1.TaskIP &&
		t.TaskPort == t1.TaskPort &&
		t.PortName == t1.PortName
}

func (t *Target) ToString() string {
	return fmt.Sprintf("%s %s %s %d %s with weight is %f\n", t.AppID, t.TaskID, t.TaskIP, t.TaskPort, t.PortName, t.Weight)
}

func (t Target) Entry() *url.URL {
	taskEntry := fmt.Sprintf("http://%s:%d", t.TaskIP, t.TaskPort)
	url, err := url.Parse(taskEntry)
	if err != nil {
		log.Error("parse target entry %s to url got err %s", taskEntry, err)
	}

	return url
}
