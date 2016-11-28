package manager

import (
	"sync"

	"github.com/Dataman-Cloud/swan/src/manager/apiserver"
	"github.com/Dataman-Cloud/swan/src/manager/ipam"
	"github.com/Dataman-Cloud/swan/src/manager/ns"
	"github.com/Dataman-Cloud/swan/src/manager/sched"
	"github.com/Dataman-Cloud/swan/src/manager/swancontext"
	. "github.com/Dataman-Cloud/swan/src/store/local"
	"github.com/Dataman-Cloud/swan/src/util"

	"github.com/Sirupsen/logrus"
)

type Manager struct {
	store     *BoltStore
	apiserver *apiserver.ApiServer
	//proxyserver

	ipamAdapter *ipam.IpamAdapter
	resolver    *ns.Resolver
	sched       *sched.Sched

	swanContext *swancontext.SwanContext
	config      util.SwanConfig
}

func New(config util.SwanConfig) (*Manager, error) {
	manager := &Manager{
		config: config,
	}

	store, err := NewBoltStore(".bolt.db")
	if err != nil {
		logrus.Errorf("Init store engine failed:%s", err)
	}

	manager.swanContext = &swancontext.SwanContext{
		Config: config,
		Store:  store,
		ApiServer: apiserver.NewApiServer(manager.config.HttpListener.TCPAddr,
			manager.config.HttpListener.UnixAddr),
	}

	manager.ipamAdapter, err = ipam.New(manager.swanContext)
	if err != nil {
		return nil, err
	}

	manager.resolver = ns.New(manager.config.DNS)
	manager.sched = sched.New(manager.config.Scheduler, manager.swanContext)

	return manager, nil
}

func (manager *Manager) Stop() error {
	return nil
}

func (manager *Manager) Start() error {
	var wg sync.WaitGroup
	var err error
	wg.Add(3)

	go func() {
		err = manager.resolver.Start()
		wg.Done()
	}()

	go func() {
		err = manager.sched.Start()
		wg.Done()
	}()

	go func() {
		err = manager.swanContext.ApiServer.ListenAndServe()
		wg.Done()
	}()
	wg.Wait()

	if err != nil {
		return err
	}

	return nil
}
