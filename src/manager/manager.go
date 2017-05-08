package manager

import (
	"errors"
	"fmt"
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
	"github.com/Dataman-Cloud/swan/src/utils"

	"github.com/Sirupsen/logrus"
	"github.com/samuel/go-zookeeper/zk"
	"golang.org/x/net/context"
)

const ZK_FLAG_NONE = 0
const SWAN_LEADER_ELECTION_NODE_PATH = "%s/leader-election"
const SWAN_ATOMIC_STORE_NODE_PATH = "%s/atomic-store"

type Leadership uint8

const (
	LeadershipUnknown  Leadership = 1
	LeadershipLeader   Leadership = 2
	LeadershipFollower Leadership = 3
)

func (l Leadership) String() string {
	switch l {
	case LeadershipUnknown:
		return "LeadershipUnknown"
	case LeadershipFollower:
		return "LeadershipFollower"
	case LeadershipLeader:
		return "LeadershipLeader"
	}
	return ""
}

var (
	ZK_DEFAULT_ACL = zk.WorldACL(zk.PermAll)
)

type Manager struct {
	scheduler *scheduler.Scheduler
	apiServer *apiserver.ApiServer
	zkConn    *zk.Conn

	previousLeadership Leadership
	cfg                config.ManagerConfig
}

func New(cfg config.ManagerConfig) (*Manager, error) {
	conn, _, err := zk.Connect(strings.Split(cfg.ZkPath.Host, ","), 5*time.Second)
	if err != nil {
		return nil, err
	}

	err = store.InitZkStore(cfg.ZkPath)
	if err != nil {
		logrus.Fatalln(err)
	}

	sched := scheduler.NewScheduler(cfg)
	route := apiserver.NewApiServer(cfg.ListenAddr)
	api.NewAndInstallAppService(route, sched)
	api.NewAndInstallStatsService(route, sched)
	api.NewAndInstallEventsService(route, sched)
	api.NewAndInstallHealthyService(route)
	api.NewAndInstallFrameworkService(route)
	api.NewAndInstallVersionService(route)

	return &Manager{
		apiServer:          route,
		previousLeadership: LeadershipUnknown,
		scheduler:          sched,
		zkConn:             conn,
		cfg:                cfg,
	}, nil
}

func (manager *Manager) Start(ctx context.Context) error {
	zkNodesPath := []string{
		manager.cfg.ZkPath.Path,
		fmt.Sprintf(SWAN_LEADER_ELECTION_NODE_PATH, manager.cfg.ZkPath.Path),
		fmt.Sprintf(SWAN_ATOMIC_STORE_NODE_PATH, manager.cfg.ZkPath.Path),
	}

	nodeExists, _, err := manager.zkConn.Exists(manager.cfg.ZkPath.Path)
	if err != nil && !isNodeDoesNotExists(err) {
		return err
	}

	if !nodeExists {
		for _, path := range zkNodesPath {
			_, err := manager.zkConn.Create(path, []byte(""), ZK_FLAG_NONE, ZK_DEFAULT_ACL)
			if err != nil {
				return err
			}
		}
	}

	eventbus.Init()

	return manager.start(ctx)
}

func (manager *Manager) start(ctx context.Context) error {
	leadershipChangeChan := make(chan Leadership)

	go func() {
		for {
			err := manager.watchLeaderChange(ctx, leadershipChangeChan)
			if err != nil {
				logrus.Errorf("watchLeaderChange go error: %+v", err)
				return
			}
			logrus.Info("watchLeaderChange exit normally")
		}
	}()

	stopCtx, stopFunc := context.WithCancel(ctx)
	errC := make(chan error)
	for {
		select {
		case change := <-leadershipChangeChan:
			// do nothing when leadership not change
			if change == manager.previousLeadership {
				continue
			}
			manager.previousLeadership = change

			switch change {
			case LeadershipLeader:
				go func() {
					errC <- eventbus.Start(stopCtx)
				}()

				go func() {
					errC <- manager.scheduler.Start(stopCtx)
				}()

				go func() {
					errC <- manager.apiServer.Start(stopCtx)
				}()

				go func() {
					errC <- store.DB().Start(stopCtx)
				}()

			case LeadershipFollower:
				stopFunc()

			case LeadershipUnknown:
				// do nothing
			}

		case err := <-errC:
			// for any error contains `context canceled` should be
			// those caused by leadership changed to LeadershipFollower
			logrus.Error(err)
			if !strings.Contains(err.Error(), "context canceled") {
				return err
			}

		case <-ctx.Done():
			manager.zkConn.Close()
			stopFunc()
		}
	}

}

func (manager *Manager) watchLeaderChange(ctx context.Context, leadershipChangeChan chan Leadership) error {
	leaderPath := filepath.Join(fmt.Sprintf(SWAN_LEADER_ELECTION_NODE_PATH, manager.cfg.ZkPath.Path), "node")
	myNode, err := manager.zkConn.CreateProtectedEphemeralSequential(leaderPath, []byte(""), ZK_DEFAULT_ACL)
	if err != nil {
		return err
	}
	_, myNodePath := filepath.Split(myNode)

reevaluateLeader:
	leaderNode, err := manager.minimalValueChild(fmt.Sprintf(SWAN_LEADER_ELECTION_NODE_PATH, manager.cfg.ZkPath.Path))
	if myNodePath == leaderNode {
		leadershipChangeChan <- LeadershipLeader
	} else {
		leadershipChangeChan <- LeadershipFollower
	}
	_, _, existsWChan, err := manager.zkConn.ExistsW(filepath.Join(fmt.Sprintf(SWAN_LEADER_ELECTION_NODE_PATH,
		manager.cfg.ZkPath.Path), leaderNode))
	if err != nil {
		return err
	}

	if _, ok := <-existsWChan; ok {
		goto reevaluateLeader
	}

	return nil
}

func (manager *Manager) minimalValueChild(path string) (string, error) {
	children, _, err := manager.zkConn.Children(path)
	if err != nil {
		return "", err
	}

	if len(children) == 0 {
		return "", errors.New("empty children in minimalValueChild")
	}

	sortablePathes := utils.SortableNodePath(children)
	sort.Sort(sortablePathes)

	return sortablePathes[0], nil
}

func isNodeDoesNotExists(err error) bool {
	return strings.Contains(err.Error(), "node does not exist")
}
