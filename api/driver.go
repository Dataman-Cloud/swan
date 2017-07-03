package api

import (
	"net/http"

	"github.com/Dataman-Cloud/swan/mesos"
	"github.com/Dataman-Cloud/swan/types"
)

type Driver interface {
	LaunchTask(*mesos.Task) error
	KillTask(string, string) error

	ClusterName() string

	SubscribeEvent(http.ResponseWriter, string) error
	TaskEvents() []*types.TaskEvent

	// for debug convenience
	Dump() interface{}
}
