package swancontext

import (
	"github.com/Dataman-Cloud/swan/manager/apiserver"
	"github.com/Dataman-Cloud/swan/store/local"
)

type SwanContext struct {
	Store     *boltdb.BoltStore
	ApiServer *apiserver.ApiServer
}
