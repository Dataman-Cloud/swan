package store

import (
	"errors"
	"sync"

	"github.com/Dataman-Cloud/swan/src/manager/raft"

	"github.com/boltdb/bolt"
)

var (
	errAppNotFound  = errors.New("Update app failed: target app was not found")
	errTaskNotFound = errors.New("Update task failed: target task was not found")
)

var (
	fs   *FrameworkStore
	once sync.Once
)

type FrameworkStore struct {
	BoltbDb  *bolt.DB
	RaftNode *raft.Node
}

func InitStore(db *bolt.DB, raftNode *raft.Node) {
	if db == nil || raftNode == nil {
		panic("db or raftnode not initialized yet")
	}
	once.Do(func() {
		fs = &FrameworkStore{
			BoltbDb:  db,
			RaftNode: raftNode,
		}
	})
}

func DB() *FrameworkStore {
	return fs
}
