package boltdb

import (
	"github.com/boltdb/bolt"
)

type BoltStore struct {
	conn *bolt.DB
	path string
}

func NewBoltStore(path string) (*BoltStore, error) {
	handle, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	store := &BoltStore{
		conn: handle,
		path: path,
	}

	if err := store.initialize(); err != nil {
		store.Close()
		return nil, err
	}

	return store, nil
}

func (b *BoltStore) initialize() error {
	tx, err := b.conn.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Create all the buckets
	if _, err := tx.CreateBucketIfNotExists([]byte("swan")); err != nil {
		return err
	}

	if _, err := tx.CreateBucketIfNotExists([]byte("applications")); err != nil {
		return err
	}

	if _, err := tx.CreateBucketIfNotExists([]byte("tasks")); err != nil {
		return err
	}

	if _, err := tx.CreateBucketIfNotExists([]byte("versions")); err != nil {
		return err
	}

	if _, err := tx.CreateBucketIfNotExists([]byte("checks")); err != nil {
		return err
	}

	return tx.Commit()
}

func (b *BoltStore) Close() error {
	return b.conn.Close()
}
