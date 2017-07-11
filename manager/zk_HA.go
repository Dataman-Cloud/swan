package manager

import (
	"path/filepath"
	"sort"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/samuel/go-zookeeper/zk"
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

func (m *Manager) setLeader(path string) {
	p := filepath.Join(m.electRootPath, path)
	_, err := m.ZKClient.Set(p, []byte(m.cfg.Listen), -1)
	if err != nil {
		log.Infof("Update leader address error %s", err.Error())
	}
}

func (m *Manager) getLeader(path string) (string, error) {
	p := filepath.Join(m.electRootPath, path)
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
