package boltdb

import (
	"github.com/boltdb/bolt"
)

type BoltStore struct {
	conn *bolt.DB
}

func NewBoltStore(db *bolt.DB) (*BoltStore, error) {
	store := &BoltStore{
		conn: db,
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
