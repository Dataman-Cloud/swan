package etcd

import (
	"errors"

	"github.com/Dataman-Cloud/swan/types"
)

var (
	errNotImplemented = errors.New("not implemented yet")
)

// Kvm App
//
//
func (s *EtcdStore) CreateKvmApp(app *types.KvmApp) error {
	return errNotImplemented
}

func (s *EtcdStore) UpdateKvmApp(app *types.KvmApp) error {
	return errNotImplemented
}

func (s *EtcdStore) GetKvmApp(id string) (*types.KvmApp, error) {
	return nil, errNotImplemented
}

func (s *EtcdStore) ListKvmApps() ([]*types.KvmApp, error) {
	return nil, errNotImplemented
}

func (s *EtcdStore) DeleteKvmApp(id string) error {
	return errNotImplemented
}

// Kvm Task
//
//
func (s *EtcdStore) CreateKvmTask(aid string, task *types.KvmTask) error {
	return errNotImplemented
}

func (s *EtcdStore) UpdateKvmTask(aid string, task *types.KvmTask) error {
	return errNotImplemented
}

func (s *EtcdStore) ListKvmTasks(id string) ([]*types.KvmTask, error) {
	return nil, errNotImplemented
}

func (s *EtcdStore) DeleteKvmTask(id string) error {
	return errNotImplemented
}

func (s *EtcdStore) GetKvmTask(aid, tid string) (*types.KvmTask, error) {
	return nil, errNotImplemented
}
