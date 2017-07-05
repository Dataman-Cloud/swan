package manager

import (
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Dataman-Cloud/swan/api"
	//"github.com/Dataman-Cloud/swan/api/middleware"
	"github.com/Dataman-Cloud/swan/config"
	"github.com/Dataman-Cloud/swan/mesos"
	"github.com/Dataman-Cloud/swan/mesos/filter"
	"github.com/Dataman-Cloud/swan/mesos/strategy"
	"github.com/Dataman-Cloud/swan/mole"
	zkstore "github.com/Dataman-Cloud/swan/store/zk"

	log "github.com/Sirupsen/logrus"
	"github.com/samuel/go-zookeeper/zk"
	//"golang.org/x/net/context"
)

type Leadership uint8

const (
	ZKFlagNone         = 0
	LeaderElectionPath = "/leader-election"

	LeadershipUnknown  Leadership = 1
	LeadershipLeader   Leadership = 2
	LeadershipFollower Leadership = 3
)

var (
	ZKDefaultACL = zk.WorldACL(zk.PermAll)
)

type Manager struct {
	sched         *mesos.Scheduler
	apiserver     *api.Server
	clusterMaster *mole.Master
	ZKClient      *zk.Conn

	cfg                *config.ManagerConfig
	leadershipChangeCh chan Leadership
	errCh              chan error
	electRootPath      string
	leader             string
	myid               string
}

func New(cfg *config.ManagerConfig) (*Manager, error) {
	conn, err := connect(strings.Split(cfg.ZKURL.Host, ","))
	if err != nil {
		return nil, err
	}

	db, err := zkstore.NewZKStore(cfg.ZKURL)
	if err != nil {
		log.Fatalln(err)
	}

	// mole master
	clusterMaster := mole.NewMaster(&mole.Config{
		Role:   mole.RoleMaster,
		Listen: "0.0.0.0:10000", // TODO
	})

	scfg := mesos.SchedulerConfig{
		ZKHost:                  strings.Split(cfg.MesosURL.Host, ","),
		ZKPath:                  cfg.MesosURL.Path,
		ReconciliationInterval:  cfg.ReconciliationInterval,
		ReconciliationStep:      cfg.ReconciliationStep,
		ReconciliationStepDelay: cfg.ReconciliationStepDelay,
	}

	var s mesos.Strategy
	switch cfg.Strategy {
	case "random":
		s = strategy.NewRandomStrategy()
	case "binpack", "binpacking":
		s = strategy.NewBinPackStrategy()
	case "spread":
		s = strategy.NewSpreadStrategy()
	default:
		s = strategy.NewBinPackStrategy()
	}

	eventMgr := mesos.NewEventManager()

	sched, err := mesos.NewScheduler(&scfg, db, s, eventMgr)
	if err != nil {
		return nil, err
	}

	filters := []mesos.Filter{
		filter.NewResourceFilter(),
		filter.NewConstraintsFilter(),
	}
	sched.InitFilters(filters)

	srvcfg := api.Config{
		Listen:   cfg.Listen,
		LogLevel: cfg.LogLevel,
	}

	srv := api.NewServer(&srvcfg)

	router := api.NewRouter(sched, db, clusterMaster)

	srv.InstallRouter(router)

	return &Manager{
		apiserver:          srv,
		sched:              sched,
		clusterMaster:      clusterMaster,
		ZKClient:           conn,
		cfg:                cfg,
		leadershipChangeCh: make(chan Leadership),
		errCh:              make(chan error, 1),
		electRootPath:      filepath.Join(cfg.ZKURL.Path, LeaderElectionPath),
	}, nil
}

func connect(srvs []string) (*zk.Conn, error) {
	conn, connChan, err := zk.Connect(srvs, 5*time.Second)
	if err != nil {
		return nil, err
	}

	for {
		select {
		case connEvent := <-connChan:
			if connEvent.State == zk.StateConnected {
				log.Info("connect to zookeeper server success!")
				return conn, nil
			}
			// TODO(nmg) should be re-connect.
			if connEvent.State == zk.StateDisconnected {
				log.Info("lost connection from zookeeper")
				return nil, nil
			}
			// TOOD(nmg) currently not work.
		case _ = <-time.After(time.Second * 5):
			conn.Close()
			return nil, nil
		}
	}
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

				m.apiserver.Update(m.leader)
			case LeadershipFollower:
				m.apiserver.Update(m.leader)
			}

		case err := <-m.errCh:
			return err
		}
	}

}

func (m *Manager) setLeader(path string) {
	p := filepath.Join(m.electRootPath, path)
	_, err := m.ZKClient.Set(p, []byte(m.cfg.Listen), -1)
	if err != nil {
		log.Infof("Update leader address error %s", err.Error())
	}
}

func (m *Manager) getLeader(path string) (string, error) {
	p := filepath.Join(m.electRootPath, path)
	// NOTE(nmg):Example to use node-data-changed event to get leader.
	// b, _, evCh, err := m.ZKClient.GetW(p)
	// if err != nil {
	// 	log.Infof("Get leader address error %s", err.Error())
	// 	return "", err
	// }
	// if len(b) > 0 {
	// 	return string(b), nil
	// }

	// for {
	// 	ev := <-evCh
	// 	if ev.Type == zk.EventNodeDataChanged {
	// 		b, _, err := m.ZKClient.Get(p)
	// 		if err != nil {
	// 			log.Infof("Get leader address error %s", err.Error())
	// 			return "", err
	// 		}
	// 		return string(b), nil
	// 	}
	// }
	for {
		b, _, err := m.ZKClient.Get(p)
		if err != nil {
			log.Infof("Get leader address error %s", err.Error())
			return "", err
		}

		if len(b) > 0 {
			return string(b), nil
		}

		time.Sleep(1 * time.Second)
	}
}

func (m *Manager) isLeader(path string) (bool, error, string) {
	children, _, err := m.ZKClient.Children(m.electRootPath)
	if err != nil {
		return false, err, ""
	}

	sort.Strings(children)

	p := children[0]

	return path == p, nil, p
}

func (m *Manager) elect() (string, error) {
	leader, err, p := m.isLeader(m.myid)
	if err != nil {
		return "", err
	}
	if leader {
		log.Info("Electing leader success.")
		m.leader = m.cfg.Listen
		m.setLeader(p)
		m.leadershipChangeCh <- LeadershipLeader

		return p, nil
	}

	log.Infof("Leader manager has been elected.")

	l, err := m.getLeader(p)
	if err != nil {
		if err == zk.ErrNoNode {
			log.Errorf("Leader lost again. start new electing...")
			return m.elect()
		}
		log.Errorf("Detect new leader error %s", err.Error())
		return "", err
	}
	log.Infof("Detect new leader at %s", l)

	m.leader = l

	m.leadershipChangeCh <- LeadershipFollower

	return p, nil

}

func (m *Manager) electLeader() (string, error) {
	p := filepath.Join(m.electRootPath, "0")
	path, err := m.ZKClient.Create(p, nil, zk.FlagEphemeral|zk.FlagSequence, ZKDefaultACL)
	if err != nil {
		return "", err
	}

	m.myid = filepath.Base(path)

	return m.elect()
}

func (m *Manager) watchLeader(path string) error {
	p := filepath.Join(m.electRootPath, path)
	_, _, childCh, err := m.ZKClient.ChildrenW(p)
	if err != nil {
		log.Infof("Watch children error %s", err)
		return err
	}

	for {
		childEvent := <-childCh
		if childEvent.Type == zk.EventNodeDeleted {
			// re-election
			log.Info("Lost leading manager. Start electing new leader...")
			// If it is better to run following steps in a seprated goroutine?
			// (memory leak maybe)
			p, err := m.elect()
			if err != nil {
				log.Infof("Electing new leader error %s", err.Error())
				return err
			}
			m.watchLeader(p)
		}
	}
}
