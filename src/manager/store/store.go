package store

import (
	"github.com/Dataman-Cloud/swan/src/manager/raft"

	"github.com/boltdb/bolt"
)

type ManagerStore struct {
	BoltbDb  *bolt.DB
	RaftNode *raft.Node
}

func NewManagerStore(db *bolt.DB, raftNode *raft.Node) *ManagerStore {
	return &ManagerStore{
		BoltbDb:  db,
		RaftNode: raftNode,
	}
}
