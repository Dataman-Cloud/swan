package store

import (
	"github.com/boltdb/bolt"
	"github.com/coreos/etcd/raft/raftpb"
	"github.com/gogo/protobuf/proto"
)

func withCreateRaftStateBucketIfNotExists(tx *bolt.Tx, fn func(bkt *bolt.Bucket) error) error {
	bkt, err := createBucketIfNotExists(tx, bucketKeyStorageVersion, BucketKeyRaftState)
	if err != nil {
		return err
	}

	return fn(bkt)
}

func getRaftStateBucket(tx *bolt.Tx) *bolt.Bucket {
	return getBucket(tx, bucketKeyStorageVersion, BucketKeyRaftState)
}

func putRaftState(tx *bolt.Tx, state raftpb.HardState) error {
	return withCreateRaftStateBucketIfNotExists(tx, func(bkt *bolt.Bucket) error {
		p, err := proto.Marshal(&state)
		if err != nil {
			return err
		}

		return bkt.Put(BucketKeyData, p)
	})
}

func getRaftState(tx *bolt.Tx) (raftpb.HardState, error) {
	bkt := getRaftStateBucket(tx)
	if bkt == nil {
		return raftpb.HardState{}, nil
	}

	var hardState raftpb.HardState
	p := bkt.Get(BucketKeyData)

	if err := proto.Unmarshal(p, &hardState); err != nil {
		return hardState, err
	}

	return hardState, nil
}

func removeRaftState(tx *bolt.Tx) error {
	storageVersionBkt := getBucket(tx, bucketKeyStorageVersion)
	if storageVersionBkt == nil {
		return nil
	}

	return storageVersionBkt.DeleteBucket(BucketKeyRaftState)
}
