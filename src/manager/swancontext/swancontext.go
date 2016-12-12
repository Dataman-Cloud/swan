package swancontext

import (
	"github.com/Dataman-Cloud/swan/src/config"
	"github.com/Dataman-Cloud/swan/src/manager/event"
	"github.com/Dataman-Cloud/swan/src/manager/store"
)

type SwanContext struct {
	Store    store.Store
	Config   config.SwanConfig
	EventBus *event.EventBus
}
