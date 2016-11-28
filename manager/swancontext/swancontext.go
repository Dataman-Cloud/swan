package swancontext

import (
	"github.com/Dataman-Cloud/swan/manager/apiserver"
	"github.com/Dataman-Cloud/swan/store"
	"github.com/Dataman-Cloud/swan/util"
)

type SwanContext struct {
	Store     store.Store
	ApiServer *apiserver.ApiServer
	Config    util.SwanConfig
}
