package janitor

import (
	"fmt"
	"net/url"

	log "github.com/Sirupsen/logrus"
)

type Target struct {
	AppID    string
	TaskID   string
	TaskIP   string
	TaskPort uint32
	PortName string
}

func (t *Target) Equal(t1 *Target) bool {
	return t.AppID == t1.AppID &&
		t.TaskID == t1.TaskID &&
		t.TaskIP == t1.TaskIP &&
		t.TaskPort == t1.TaskPort &&
		t.PortName == t1.PortName
}

func (t *Target) ToString() string {
	return fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s", t.AppID, t.TaskID, t.TaskIP, t.TaskPort, t.PortName)
}

func (t Target) Entry() *url.URL {
	taskEntry := fmt.Sprintf("http://%s:%d", t.TaskIP, t.TaskPort)
	url, err := url.Parse(taskEntry)
	if err != nil {
		log.Error("parse target entry %s to url got err %s", taskEntry, err)
	}

	return url
}
