package swancontext

import (
	"github.com/Dataman-Cloud/swan/src/config"
	"github.com/Dataman-Cloud/swan/src/manager/event"

	"github.com/Sirupsen/logrus"
)

var instance *SwanContext

type SwanContext struct {
	Config   config.SwanConfig
	EventBus *event.EventBus
}

func NewSwanContext(c config.SwanConfig, eventBus *event.EventBus) *SwanContext {
	instance = &SwanContext{
		Config:   c,
		EventBus: eventBus,
	}

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
