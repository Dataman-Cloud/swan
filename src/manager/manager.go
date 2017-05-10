package manager

import (
	"os"
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

const (
	ZK_FLAG_NONE         = 0
	LEADER_ELECTION_PATH = "/leader-election"
)

type Leadership uint8

const (
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

	criticalErrorChan chan error

	conf                 config.ManagerConfig
	leadershipChangeChan chan Leadership
	electPath            string
	myid                 string
}

func New(managerConf config.ManagerConfig) (*Manager, error) {
	conn, _, err := zk.Connect(strings.Split(managerConf.ZkPath.Host, ","), 5*time.Second)
	if err != nil {
		return nil, err
	}

	err = store.InitZkStore(managerConf.ZkPath)
	if err != nil {
		logrus.Fatalln(err)
	}

	sched := scheduler.NewScheduler(managerConf)
	route := apiserver.NewApiServer(managerConf.ListenAddr)
	api.NewAndInstallAppService(route, sched)
	api.NewAndInstallStatsService(route, sched)
	api.NewAndInstallEventsService(route, sched)
	api.NewAndInstallHealthyService(route)
	api.NewAndInstallFrameworkService(route)
	api.NewAndInstallVersionService(route)
	api.NewAndInstallComposeService(route, sched)

	return &Manager{
		apiServer:            route,
		criticalErrorChan:    make(chan error, 1),
		scheduler:            sched,
		ZKClient:             conn,
		conf:                 managerConf,
		leadershipChangeChan: make(chan Leadership),
		electPath:            filepath.Join(managerConf.ZkPath.Path, LEADER_ELECTION_PATH),
	}, nil
}

func (m *Manager) InitAndStart(ctx context.Context) error {
	paths := []string{
		m.conf.ZkPath.Path,
		m.electPath,
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

	return m.start(ctx)
}

func (m *Manager) start(ctx context.Context) error {
	eventbus.Init()

	go func() {
		p, err := m.electLeader()
		if err != nil {
			logrus.Info("Electing lead manager failure, ", err)
			return
		}
		m.watchLeader(p)
	}()

	var stopFunc context.CancelFunc
	for {
		select {
		case change := <-m.leadershipChangeChan:
			// do nothing when leadership not change
			switch change {
			case LeadershipLeader:
				var stopCtx context.Context
				stopCtx, stopFunc = context.WithCancel(ctx)
				go func() {
					m.criticalErrorChan <- eventbus.Start(stopCtx)
				}()

				go func() {
					m.criticalErrorChan <- m.scheduler.Start(stopCtx)
				}()

				go func() {
					m.criticalErrorChan <- m.apiServer.Start(stopCtx)
				}()

				go func() {
					m.criticalErrorChan <- store.DB().Start(stopCtx)
				}()

			case LeadershipFollower:
				if stopFunc != nil {
					stopFunc()
				}

			case LeadershipUnknown:
				// do nothing
			}

		case err := <-m.criticalErrorChan:
			// for any error contains `context canceled` should be
			// those caused by leadership changed to LeadershipFollower
			logrus.Error(err)
			if !strings.Contains(err.Error(), "context canceled") {
				return err
			}

		case <-ctx.Done():
			if stopFunc != nil {
				stopFunc()
			}
			os.Exit(0)
		}
	}

}

func (m *Manager) setLeader(path string) {
	p := filepath.Join(m.electPath, path)
	_, err := m.ZKClient.Set(p, []byte(m.conf.ListenAddr), -1)
	if err != nil {
		logrus.Infof("Update leader address error %s", err.Error())
	}
}

func (m *Manager) getLeader(path string) string {
	p := filepath.Join(m.electPath, path)
	b, _, err := m.ZKClient.Get(p)
	if err != nil {
		logrus.Infof("Get leader address error %s", err.Error())
		return ""
	}

	return string(b)
}

func (m *Manager) isLeader(path string) (bool, error, string) {
	children, _, err := m.ZKClient.Children(m.electPath)
	if err != nil {
		return false, err, ""
	}

	sort.Strings(children)

	p := children[0]

	return path == p, nil, p
}

func (m *Manager) elect(path string) (string, error) {
	leader, err, p := m.isLeader(path)
	if err != nil {
		return "", err
	}
	if leader {
		logrus.Info("Electing leader success.")
		m.setLeader(p)
		m.leadershipChangeChan <- LeadershipLeader

		return p, nil
	}

	logrus.Infof("Leader manager has been elected.")
	logrus.Infof("Detect new leader at %s", m.getLeader(p))
	m.leadershipChangeChan <- LeadershipFollower

	return p, nil

}

func (m *Manager) electLeader() (string, error) {
	path, err := m.ZKClient.Create(filepath.Join(m.electPath, "0"), nil, zk.FlagEphemeral|zk.FlagSequence, ZK_DEFAULT_ACL)
	if err != nil {
		return "", err
	}

	p := filepath.Base(path)
	m.myid = p

	return m.elect(p)
}

func (m *Manager) watchLeader(path string) {
	pathW := filepath.Join(m.electPath, path)
	_, _, childCh, err := m.ZKClient.ChildrenW(pathW)
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
			p, err := m.elect(m.myid)
			if err != nil {
				logrus.Infof("Electing new leader error %s", err.Error())
				return
			}
			m.watchLeader(p)
		}
	}
}
