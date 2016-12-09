package store

import (
	"github.com/Dataman-Cloud/swan/src/manager/raft"

	"github.com/boltdb/bolt"
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
