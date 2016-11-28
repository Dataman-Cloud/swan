package manager

import (
	"github.com/Dataman-Cloud/swan/manager/apiserver"
	"github.com/Dataman-Cloud/swan/manager/ipam"
	"github.com/Dataman-Cloud/swan/manager/ns"
	"github.com/Dataman-Cloud/swan/manager/sched"
	"github.com/Dataman-Cloud/swan/manager/swancontext"
	. "github.com/Dataman-Cloud/swan/store/local"
	"github.com/Dataman-Cloud/swan/util"

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

func (manager *Manager) Start() error {
	return nil
}

func (manager *Manager) Stop() error {
	return nil
}

func (manager *Manager) Run() error {
	dnsServer := ns.New(manager.config.DNS)
	_, errCh := dnsServer.Run()
	go func() {
		err := <-errCh
		logrus.Errorf("dns running go error %s", err)
	}()

	manager.swanContext.ApiServer.ListenAndServe()
	err := manager.sched.Run()
	if err != nil {
		return err
	}

	return nil
}
