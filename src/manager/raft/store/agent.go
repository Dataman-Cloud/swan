package store

import (
	"github.com/Dataman-Cloud/swan/src/manager/raft/types"

	"github.com/boltdb/bolt"
)

func withCreateAgentsBucketIfNotExists(tx *bolt.Tx, ID string, fn func(bkt *bolt.Bucket) error) error {
	bkt, err := createBucketIfNotExists(tx, bucketKeyStorageVersion, bucketKeyAgents, []byte(ID))
	if err != nil {
		return err
	}

	return fn(bkt)
}

func getAgentsBucket(tx *bolt.Tx) *bolt.Bucket {
	return getBucket(tx, bucketKeyStorageVersion, bucketKeyAgents)
}

func getAgentBucket(tx *bolt.Tx, ID string) *bolt.Bucket {
	return getBucket(tx, bucketKeyStorageVersion, bucketKeyAgents, []byte(ID))
}

func withAgentBucket(tx *bolt.Tx, ID string, fn func(bkt *bolt.Bucket) error) error {
	bkt := getAgentBucket(tx, ID)
	if bkt == nil {
		return ErrAgentUnknown
	}

	return fn(bkt)
}

func createAgent(tx *bolt.Tx, agent *types.Agent) error {
	return withCreateAgentsBucketIfNotExists(tx, agent.ID, func(bkt *bolt.Bucket) error {
		p, err := agent.Marshal()
		if err != nil {
			return err
		}

		return bkt.Put(BucketKeyData, p)
	})
}

func updateAgent(tx *bolt.Tx, agent *types.Agent) error {
	return withAgentBucket(tx, agent.ID, func(bkt *bolt.Bucket) error {
		p, err := agent.Marshal()
		if err != nil {
			return err
		}

		return bkt.Put(BucketKeyData, p)
	})
}

func removeAgent(tx *bolt.Tx, ID string) error {
	agentsBkt := getAgentsBucket(tx)
	if agentsBkt == nil {
		return nil
	}

	return agentsBkt.DeleteBucket([]byte(ID))
}
