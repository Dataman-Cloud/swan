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
	"github.com/Dataman-Cloud/swan/src/manager/apiserver"
	fstore "github.com/Dataman-Cloud/swan/src/manager/store"

	"github.com/Sirupsen/logrus"
	"github.com/samuel/go-zookeeper/zk"
	"golang.org/x/net/context"
)

const ZK_FLAG_NONE = 0

var (
	ErrNormalExit = errors.New("normal exit 0")
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

// example node path  => /swan/leader-election/_c_c7b2927d40ec05db4d199a804437995c-node0000000023
// sortable value is node0000000023
type SortableNodePath []string

func (a SortableNodePath) Len() int      { return len(a) }
func (a SortableNodePath) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a SortableNodePath) Less(i, j int) bool {
	return strings.SplitN(a[i], "-", -1)[1] < strings.SplitN(a[j], "-", -1)[1]
}

type ZkManager struct {
	CancelFunc context.CancelFunc

	framework         *Framework
	apiServer         *apiserver.ApiServer
	criticalErrorChan chan error
	zkConn            *zk.Conn
	zkPath            string

	previousLeadership Leadership
}

func NewZK(managerConf config.ManagerConfig) (*ZkManager, error) {
	managerServer := apiserver.NewApiServer(managerConf.ListenAddr, managerConf.AdvertiseAddr)

	frameworkStore := fstore.NewZkStore()
	framework, err := New(frameworkStore, managerServer)
	if err != nil {
		logrus.Errorf("init framework failed. Error: %s", err.Error())
		return nil, err
	}

	manager := &ZkManager{
		framework:          framework,
		apiServer:          managerServer,
		criticalErrorChan:  make(chan error, 1),
		previousLeadership: LeadershipUnknown,
	}

	conn, _, err := zk.Connect(strings.Split(managerConf.ZkPath.Host, ","), 5*time.Second)
	if err != nil {
		return nil, err
	}

	manager.zkConn = conn
	manager.zkPath = managerConf.ZkPath.Path

	return manager, nil
}

func (manager *ZkManager) Stop() {
	manager.CancelFunc()

	return
}

func (manager *ZkManager) InitAndStart(ctx context.Context) error {
	zkNodesPath := []string{
		manager.zkPath,
		filepath.Join(manager.zkPath, "leader-election"),
		filepath.Join(manager.zkPath, "store-op"),
	}

	nodeExists, _, err := manager.zkConn.Exists(manager.zkPath)
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

func (manager *ZkManager) start(ctx context.Context) error {
	var eventBusStarted, frameworkStarted bool
	eventbus.Init()
	leadershipChangeChan := make(chan Leadership)

	go func() {
		for {
			err := manager.subcribeLeaderChange(ctx, leadershipChangeChan)
			if err == ErrNormalExit {
				logrus.Info("subcribeLeaderChange exit normally")
				return
			} else {
				logrus.Errorf("subcribeLeaderChange go error: %+v", err)
			}
		}
	}()

	for {
		select {
		case change := <-leadershipChangeChan:
			if change == manager.previousLeadership {
				continue
			}

			// toggle state
			manager.previousLeadership = change

			switch change {
			case LeadershipLeader:
				go func() {
					eventBusStarted = true
					eventbus.Start(ctx)
				}()

				go func() {
					frameworkStarted = true
					manager.criticalErrorChan <- manager.framework.Start(ctx)
				}()

				go func() {
					manager.criticalErrorChan <- manager.apiServer.Start()
				}()

			case LeadershipUnknown:
			case LeadershipFollower:
				if eventBusStarted {
					eventbus.Stop()
					eventBusStarted = false
				}

				if frameworkStarted {
					manager.framework.Stop()
					frameworkStarted = false
				}
			}
		case err := <-manager.criticalErrorChan:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (manager *ZkManager) subcribeLeaderChange(ctx context.Context, leadershipChangeChan chan Leadership) error {
	myNode, err := manager.zkConn.CreateProtectedEphemeralSequential(manager.zkPath+"/"+"leader-election/node", []byte(""), ZK_DEFAULT_ACL)
	if err != nil {
		return err
	}
	_, myNodePath := filepath.Split(myNode)

reevaluateLeader:
	leaderNode, err := manager.minimalValueChild(manager.zkPath + "/" + "leader-election")
	if myNodePath == leaderNode {
		leadershipChangeChan <- LeadershipLeader
	} else {
		leadershipChangeChan <- LeadershipFollower
	}
	_, _, existsWChan, err := manager.zkConn.ExistsW(filepath.Join(manager.zkPath, "leader-election", leaderNode))
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

func (manager *ZkManager) minimalValueChild(path string) (string, error) {
	children, _, err := manager.zkConn.Children(path)
	if err != nil {
		return "", err
	}

	sortablePathes := SortableNodePath(children)
	sort.Sort(sortablePathes)

	return sortablePathes[0], nil
}

func isNodeDoesNotExists(err error) bool {
	fmt.Println(strings.Contains(err.Error(), "node does not exist"))
	return strings.Contains(err.Error(), "node does not exist")
}
