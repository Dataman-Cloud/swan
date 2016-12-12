package swancontext

import (
	"github.com/Dataman-Cloud/swan/src/manager/event"
	"github.com/Dataman-Cloud/swan/src/manager/store"
	"github.com/Dataman-Cloud/swan/src/util"
)

type SwanContext struct {
	Store    store.Store
	Config   util.SwanConfig
	EventBus *event.EventBus
}
