package manager

import (
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Dataman-Cloud/swan/src/config"
	eventbus "github.com/Dataman-Cloud/swan/src/event"
	"github.com/Dataman-Cloud/swan/src/manager/api"
	"github.com/Dataman-Cloud/swan/src/manager/apiserver"
	"github.com/Dataman-Cloud/swan/src/manager/scheduler"
	"github.com/Dataman-Cloud/swan/src/manager/store"

	"github.com/Sirupsen/logrus"
	"github.com/samuel/go-zookeeper/zk"
	"golang.org/x/net/context"
)

type Leadership uint8

const (
	ZK_FLAG_NONE = 0

	LEADER_ELECTION_PATH = "/leader-election"

	LeadershipUnknown  Leadership = 1
	LeadershipLeader   Leadership = 2
	LeadershipFollower Leadership = 3
)

var (
	ZK_DEFAULT_ACL = zk.WorldACL(zk.PermAll)
)

type Manager struct {
	scheduler *scheduler.Scheduler
	apiServer *apiserver.ApiServer
	ZKClient  *zk.Conn

	cfg                config.ManagerConfig
	leadershipChangeCh chan Leadership
	errCh              chan error
	electRootPath      string
	myid               string
}

func New(cfg config.ManagerConfig) (*Manager, error) {
	conn, err := connect(strings.Split(cfg.ZKURL.Host, ","))
	if err != nil {
		return nil, err
	}

	err = store.InitZkStore(cfg.ZKURL)
	if err != nil {
		logrus.Fatalln(err)
	}

	sched := scheduler.NewScheduler(cfg)
	route := apiserver.NewApiServer(cfg.ListenAddr)

	setupRoutes(route, sched)

	return &Manager{
		apiServer:          route,
		scheduler:          sched,
		ZKClient:           conn,
		cfg:                cfg,
		leadershipChangeCh: make(chan Leadership),
		errCh:              make(chan error, 1),
		electRootPath:      filepath.Join(cfg.ZKURL.Path, LEADER_ELECTION_PATH),
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
				logrus.Info("connect to zookeeper server success!")
				return conn, nil
			}
		}
	}
}

func setupRoutes(r *apiserver.ApiServer, s *scheduler.Scheduler) {
	api.NewAndInstallAppService(r, s)
	api.NewAndInstallStatsService(r, s)
	api.NewAndInstallEventsService(r, s)
	api.NewAndInstallHealthyService(r)
	api.NewAndInstallFrameworkService(r)
	api.NewAndInstallVersionService(r)
	api.NewAndInstallComposeService(r, s)
}

func (m *Manager) Start() error {
	paths := []string{
		m.cfg.ZKURL.Path,
		m.electRootPath,
	}

	var (
		err    error
		exists bool
	)
	for _, p := range paths {
		exists, _, err = m.ZKClient.Exists(p)
		if err != nil {
			return err
		}
		if exists {
			continue
		}
		_, err = m.ZKClient.Create(p, []byte{}, ZK_FLAG_NONE, ZK_DEFAULT_ACL)
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
			logrus.Info("Electing lead manager failure, ", err)
			return
		}

		m.watchLeader(p)
	}()

	var (
		stopFunc context.CancelFunc
		stopCtx  context.Context
	)
	for {
		select {
		case c := <-m.leadershipChangeCh:
			// do nothing when leadership not change
			switch c {
			case LeadershipLeader:
				stopCtx, stopFunc = context.WithCancel(context.TODO())
				m.startServices(stopCtx, m.errCh)

			case LeadershipFollower:
				m.stopServices(stopFunc)

				// NOTE(nmg): this case should be removed. testing.
			case LeadershipUnknown:
				// do nothing
			}

		case err := <-m.errCh:
			// for any error contains `context canceled` should be
			// those caused by leadership changed to LeadershipFollower
			if !strings.Contains(err.Error(), "context canceled") {
				m.stopServices(stopFunc)
				return err
			}
		}
	}

}

func (m *Manager) startServices(ctx context.Context, err chan error) {
	// NOTE(nmg): m.errCh never be closed.
	go func() {
		err <- eventbus.Start(ctx)
	}()

	go func() {
		err <- m.scheduler.Start(ctx)
	}()

	go func() {
		err <- m.apiServer.Start(ctx)
	}()

	go func() {
		err <- store.DB().Start(ctx)
	}()
}

func (m *Manager) stopServices(cancel context.CancelFunc) {
	if cancel != nil {
		cancel()
	}
}

func (m *Manager) setLeader(path string) {
	p := filepath.Join(m.electRootPath, path)
	_, err := m.ZKClient.Set(p, []byte(m.cfg.ListenAddr), -1)
	if err != nil {
		logrus.Infof("Update leader address error %s", err.Error())
	}
}

func (m *Manager) getLeader(path string) string {
	p := filepath.Join(m.electRootPath, path)
	b, _, err := m.ZKClient.Get(p)
	if err != nil {
		logrus.Infof("Get leader address error %s", err.Error())
		return ""
	}

	return string(b)
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
		logrus.Info("Electing leader success.")
		m.setLeader(p)
		m.leadershipChangeCh <- LeadershipLeader

		return p, nil
	}

	logrus.Infof("Leader manager has been elected.")
	logrus.Infof("Detect new leader at %s", m.getLeader(p))
	m.leadershipChangeCh <- LeadershipFollower

	return p, nil

}

func (m *Manager) electLeader() (string, error) {
	p := filepath.Join(m.electRootPath, "0")
	path, err := m.ZKClient.Create(p, nil, zk.FlagEphemeral|zk.FlagSequence, ZK_DEFAULT_ACL)
	if err != nil {
		return "", err
	}

	m.myid = filepath.Base(path)

	return m.elect()
}

func (m *Manager) watchLeader(path string) {
	p := filepath.Join(m.electRootPath, path)
	_, _, childCh, err := m.ZKClient.ChildrenW(p)
	if err != nil {
		logrus.Infof("Watch children error %s", err)
		return
	}

	for {
		childEvent := <-childCh
		if childEvent.Type == zk.EventNodeDeleted {
			// re-election
			logrus.Info("Lost leading manager. Start electing new leader...")
			// If it is better to run following steps in a seprated goroutine?
			// (memory leak maybe)
			p, err := m.elect()
			if err != nil {
				logrus.Infof("Electing new leader error %s", err.Error())
				return
			}
			m.watchLeader(p)
		}
	}
}
