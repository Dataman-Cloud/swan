package manager

import (
	"errors"
	"fmt"
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
	"github.com/Dataman-Cloud/swan/src/utils"

	"github.com/Sirupsen/logrus"
	"github.com/samuel/go-zookeeper/zk"
	"golang.org/x/net/context"
)

const ZK_FLAG_NONE = 0
const SWAN_LEADER_ELECTION_NODE_PATH = "%s/leader-election"
const SWAN_ATOMIC_STORE_NODE_PATH = "%s/atomic-store"

var (
	ErrNormalExit = errors.New("normal exit 0")
)

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

	criticalErrorChan chan error

	previousLeadership Leadership
	conf               config.ManagerConfig
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

	return &Manager{
		apiServer:          route,
		criticalErrorChan:  make(chan error, 1),
		previousLeadership: LeadershipUnknown,
		scheduler:          sched,
		zkConn:             conn,
		conf:               managerConf,
	}, nil
}

func (manager *Manager) InitAndStart(ctx context.Context) error {
	zkNodesPath := []string{
		manager.conf.ZkPath.Path,
		fmt.Sprintf(SWAN_LEADER_ELECTION_NODE_PATH, manager.conf.ZkPath.Path),
		fmt.Sprintf(SWAN_ATOMIC_STORE_NODE_PATH, manager.conf.ZkPath.Path),
	}

	nodeExists, _, err := manager.zkConn.Exists(manager.conf.ZkPath.Path)
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

	return manager.start(ctx)
}

func (manager *Manager) start(ctx context.Context) error {
	leadershipChangeChan := make(chan Leadership)
	eventbus.Init()

	go func() {
		for {
			err := manager.watchLeaderChange(ctx, leadershipChangeChan)
			if err == ErrNormalExit {
				logrus.Info("watchLeaderChange exit normally")
				return
			} else {
				logrus.Errorf("watchLeaderChange go error: %+v", err)
			}
		}
	}()

	var stopFunc context.CancelFunc
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
				var stopCtx context.Context
				stopCtx, stopFunc = context.WithCancel(ctx)
				go func() {
					manager.criticalErrorChan <- eventbus.Start(stopCtx)
				}()

				go func() {
					manager.criticalErrorChan <- manager.scheduler.Start(stopCtx)
				}()

				go func() {
					manager.criticalErrorChan <- manager.apiServer.Start(stopCtx)
				}()

				go func() {
					manager.criticalErrorChan <- store.DB().Start(stopCtx)
				}()

			case LeadershipFollower:
				if stopFunc != nil {
					stopFunc()
				}

			case LeadershipUnknown:
				// do nothing
			}

		case err := <-manager.criticalErrorChan:
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

func (manager *Manager) watchLeaderChange(ctx context.Context, leadershipChangeChan chan Leadership) error {
	leaderPath := filepath.Join(fmt.Sprintf(SWAN_LEADER_ELECTION_NODE_PATH, manager.conf.ZkPath.Path), "node")
	myNode, err := manager.zkConn.CreateProtectedEphemeralSequential(leaderPath, []byte(""), ZK_DEFAULT_ACL)
	if err != nil {
		return err
	}
	_, myNodePath := filepath.Split(myNode)

reevaluateLeader:
	leaderNode, err := manager.minimalValueChild(fmt.Sprintf(SWAN_LEADER_ELECTION_NODE_PATH, manager.conf.ZkPath.Path))
	if myNodePath == leaderNode {
		leadershipChangeChan <- LeadershipLeader
	} else {
		leadershipChangeChan <- LeadershipFollower
	}
	_, _, existsWChan, err := manager.zkConn.ExistsW(filepath.Join(fmt.Sprintf(SWAN_LEADER_ELECTION_NODE_PATH,
		manager.conf.ZkPath.Path), leaderNode))
	if err != nil {
		return err
	}

	select {
	case <-existsWChan:
		goto reevaluateLeader
	case <-ctx.Done():
		return ErrNormalExit
	}
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
