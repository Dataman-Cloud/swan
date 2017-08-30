package manager

import (
	"path/filepath"
	"strings"

	"github.com/Dataman-Cloud/swan/api"
	"github.com/Dataman-Cloud/swan/config"
	"github.com/Dataman-Cloud/swan/mesos"
	"github.com/Dataman-Cloud/swan/mole"
	"github.com/Dataman-Cloud/swan/store"

	log "github.com/Sirupsen/logrus"
	"github.com/samuel/go-zookeeper/zk"
)

type Manager struct {
	sched         *mesos.Scheduler
	apiserver     *api.Server
	clusterMaster *mole.Master
	tcpMux        *tcpMux // dispatch tcp Conn to clusterMaster & apiServer
	ZKClient      *zk.Conn

	cfg                *config.ManagerConfig
	leadershipChangeCh chan Leadership
	errCh              chan error
	electRootPath      string
	leader             string
	myid               string
}

func New(cfg *config.ManagerConfig) (*Manager, error) {
	// connect to zk leader
	conn, err := connect(strings.Split(cfg.ZKURL.Host, ","))
	if err != nil {
		return nil, err
	}

	// db store initilizing
	db, err := store.Setup(cfg.StoreType, cfg.ZKURL, cfg.EtcdAddrs)
	if err != nil {
		log.Fatalln("db store setup", err)
	}

	// tcpMux setup
	tcpMux := newTCPMux(cfg.Listen)
	hl := tcpMux.NewHTTPListener()
	ml := tcpMux.NewMoleListener()

	// mole protocol master
	clusterMaster := mole.NewMaster(ml)

	// scheduler setup
	scfg := mesos.SchedulerConfig{
		ZKHost:                  strings.Split(cfg.MesosURL.Host, ","),
		ZKPath:                  cfg.MesosURL.Path,
		Strategy:                cfg.Strategy,
		ReconciliationInterval:  cfg.ReconciliationInterval,
		ReconciliationStep:      cfg.ReconciliationStep,
		ReconciliationStepDelay: cfg.ReconciliationStepDelay,
		HeartbeatTimeout:        cfg.HeartbeatTimeout,
		MaxTasksPerOffer:        cfg.MaxTasksPerOffer,
		EnableCapabilityKilling: cfg.EnableCapabilityKilling,
		EnableCheckPoint:        cfg.EnableCheckPoint,
	}

	sched, err := mesos.NewScheduler(&scfg, db, clusterMaster)
	if err != nil {
		return nil, err
	}

	// api server
	srvcfg := api.Config{
		Advertise: cfg.Advertise,
		LogLevel:  cfg.LogLevel,
	}
	srv := api.NewServer(&srvcfg, hl, sched, db)

	// final
	return &Manager{
		apiserver:          srv,
		sched:              sched,
		clusterMaster:      clusterMaster,
		tcpMux:             tcpMux,
		ZKClient:           conn,
		cfg:                cfg,
		leadershipChangeCh: make(chan Leadership),
		errCh:              make(chan error, 1),
		electRootPath:      filepath.Join(cfg.ZKURL.Path, LeaderElectionPath),
	}, nil
}

func (m *Manager) Start() error {
	p := m.electRootPath
	exists, _, err := m.ZKClient.Exists(p)
	if err != nil {
		return err
	}
	if !exists {
		_, err = m.ZKClient.Create(p, []byte{}, ZKFlagNone, ZKDefaultACL)
		if err != nil {
			return err
		}
	}

	return m.start()
}

func (m *Manager) start() error {
	go func() {
		p, err := m.electLeader()
		if err != nil {
			log.Info("Electing lead manager failure, ", err)
			m.errCh <- err
			return
		}

		if err := m.watchLeader(p); err != nil {
			log.Info("Electing leader error", err)
			m.errCh <- err
			return
		}
	}()

	go func() {
		if err := m.tcpMux.ListenAndServe(); err != nil {
			log.Errorf("start tcpMux error: %v", err)
			m.errCh <- err
		}
	}()

	go func() {
		if err := m.apiserver.Run(); err != nil {
			log.Errorf("start apiserver error: %v", err)
			m.errCh <- err
		}
	}()

	go func() {
		if err := m.clusterMaster.Serve(); err != nil {
			log.Errorf("start mole master error: %v", err)
			m.errCh <- err
		}
	}()

	for {
		select {
		case c := <-m.leadershipChangeCh:
			switch c {
			case LeadershipLeader:
				if err := m.sched.Subscribe(); err != nil {
					log.Errorf("subscribe to mesos leader error: %v", err)
					m.errCh <- err
				}

				m.apiserver.UpdateLeader(m.leader)

			case LeadershipFollower:
				log.Warnln("became follower, closing all agents ...")
				m.clusterMaster.CloseAllAgents()
				m.apiserver.UpdateLeader(m.leader)
			}

		case err := <-m.errCh:
			return err
		}
	}

}
