package store

import (
	"github.com/Dataman-Cloud/swan/src/manager/raft/types"

	"golang.org/x/net/context"
)

type Store interface {
	CreateApp(ctx context.Context, app *types.Application, cb func()) error
	GetApp(appId string) (*types.Application, error)
	ListApps() ([]*types.Application, error)
	DeleteApp(ctx context.Context, appId string, cb func()) error
	UpdateVersion(ctx context.Context, appId string, version *types.Version, cb func()) error
	ListVersions(appId string) ([]*types.Version, error)
}
