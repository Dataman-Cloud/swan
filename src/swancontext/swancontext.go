package swancontext

import (
	"github.com/Dataman-Cloud/swan/src/apiserver"
	"github.com/Dataman-Cloud/swan/src/config"
	"github.com/Dataman-Cloud/swan/src/event"

	"github.com/Sirupsen/logrus"
)

var instance *SwanContext

type SwanContext struct {
	Config    config.SwanConfig
	EventBus  *event.EventBus
	ApiServer *apiserver.ApiServer
}

func NewSwanContext(c config.SwanConfig, eventBus *event.EventBus) *SwanContext {
	instance = &SwanContext{
		Config:   c,
		EventBus: eventBus,
	}

	instance.ApiServer = apiserver.NewApiServer(c.ListenAddr)

	return instance
}

func Instance() *SwanContext {
	if instance == nil {
		logrus.Errorf("swancontext doesn't exists, NewSwanContext to instanciate it")
		return nil
	} else {
		return instance
	}
}

func IsManager() bool {
	return instance.Config.Mode == config.Manager || instance.Config.Mode == config.Mixed
}

func IsAgent() bool {
	return instance.Config.Mode == config.Agent || instance.Config.Mode == config.Mixed
}

// TODO(upccup) use joinAddrs to differentiate start new cluster or join to an existed cluster is not
// very exact. may we need a new started flags to mark it.
func IsNewCluster() bool {
	if !IsManager() {
		return false
	}

	if len(instance.Config.JoinAddrs) == 0 {
		return true
	}

	for _, joinAddr := range instance.Config.JoinAddrs {
		if instance.Config.AdvertiseAddr == joinAddr {
			return true
		}
	}

	return false
}
