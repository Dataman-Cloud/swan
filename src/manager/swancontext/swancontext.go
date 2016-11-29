package swancontext

import (
	"github.com/Dataman-Cloud/swan/src/manager/apiserver"
	"github.com/Dataman-Cloud/swan/src/manager/store"
	"github.com/Dataman-Cloud/swan/src/util"
)

type SwanContext struct {
	Store     store.Store
	ApiServer *apiserver.ApiServer
	Config    util.SwanConfig
}
