package swancontext

import (
	"github.com/Dataman-Cloud/swan/src/config"
	"github.com/Dataman-Cloud/swan/src/manager/event"
)

type SwanContext struct {
	Config   config.SwanConfig
	EventBus *event.EventBus
}
