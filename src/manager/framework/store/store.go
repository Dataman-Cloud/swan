package store

import (
	"errors"

	"github.com/Dataman-Cloud/swan/src/manager/raft"

	"github.com/boltdb/bolt"
)

var (
	ErrAppNotFound  = errors.New("Update app failed: target app was not found")
	ErrTaskNotFound = errors.New("Update task failed: target task was not found")
)

type FrameworkStore struct {
	BoltbDb  *bolt.DB
	RaftNode *raft.Node
}

func NewStore(db *bolt.DB, raftNode *raft.Node) *FrameworkStore {
	return &FrameworkStore{
		BoltbDb:  db,
		RaftNode: raftNode,
	}
}
