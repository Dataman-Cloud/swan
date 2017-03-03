package swancontext

import (
	"github.com/Dataman-Cloud/swan/src/event"

	"github.com/Sirupsen/logrus"
)

var instance *SwanContext

type SwanContext struct {
	EventBus *event.EventBus
}

func NewSwanContext(eventBus *event.EventBus) *SwanContext {
	instance = &SwanContext{
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
