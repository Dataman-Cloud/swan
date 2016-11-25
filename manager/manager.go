package manager

import (
	"github.com/Dataman-Cloud/swan/manager/apiserver"
	//"github.com/Dataman-Cloud/swan/manager/ipam"
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

	//IPAM     *ipam.IPAM
	resolver *ns.Resolver
	sched    *sched.Sched

	config      util.SwanConfig
	swanContext *swancontext.SwanContext
}

func New(config util.SwanConfig) *Manager {
	manager := &Manager{
		config: config,
	}

	store, err := NewBoltStore(".bolt.db")
	if err != nil {
		logrus.Errorf("Init store engine failed:%s", err)
	}

	manager.swanContext = &swancontext.SwanContext{
		Store: store,
		ApiServer: apiserver.NewApiServer(manager.config.HttpListener.TCPAddr,
			manager.config.HttpListener.UnixAddr),
	}

	//manager.IPAM = ipam.NewIPAM(manager.managerContext)
	manager.resolver = ns.New(manager.config.DNS)
	manager.sched = sched.New(manager.config.Scheduler, manager.swanContext)

	return manager
}

func (manager *Manager) Start() {}

func (manager *Manager) Run() error {
	dnsServer := ns.New(manager.config.DNS)
	_, errCh := dnsServer.Run()
	go func() {
		err := <-errCh
		logrus.Errorf("dns runing go error %s", err)
	}()

	manager.swanContext.ApiServer.ListenAndServe()
	err := manager.sched.Run()
	if err != nil {
		return err
	}

	return nil
}
