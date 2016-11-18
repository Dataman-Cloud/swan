package ipam

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/boltdb/bolt"
)

const (
	BUCKET_IPAM = "ipam"
)

const (
	BOLTDB_OPEN_TIMEOUT = time.Second * 5
)

type BoltStore struct {
	conn *bolt.DB
	path string
}

func NewBoltStore(path string) (*BoltStore, error) {
	handle, err := bolt.Open(path, 0600, &bolt.Options{Timeout: BOLTDB_OPEN_TIMEOUT})
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

	buckets_need_create := []string{
		"swan",
		"applications",
		"tasks",
		"versions",
		"checks",
		BUCKET_IPAM,
	}

	for _, bucket := range buckets_need_create {
		if _, err := tx.CreateBucketIfNotExists([]byte(bucket)); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (b *BoltStore) Close() error {
	return b.conn.Close()
}

func (b *BoltStore) SaveIP(ip IP) error {
	tx, err := b.conn.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bucket := tx.Bucket([]byte(BUCKET_IPAM))

	data, err := json.Marshal(ip)
	if err != nil {
		return err
	}

	if err := bucket.Put([]byte(ip.Key()), data); err != nil {
		return err
	}

	return tx.Commit()
}

func (b *BoltStore) RetriveIP(key string) (IP, error) {
	tx, err := b.conn.Begin(true)
	if err != nil {
		return IP{}, err
	}
	defer tx.Rollback()

	bucket := tx.Bucket([]byte(BUCKET_IPAM))

	data := bucket.Get([]byte(key))
	if data == nil {
		return IP{}, errors.New("Not Found")
	}

	var ip IP
	if err := json.Unmarshal(data, &ip); err != nil {
		return IP{}, err
	}

	return ip, nil
}

func (b *BoltStore) ListAllIPs() (IPList, error) {
	tx, err := b.conn.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	bucket := tx.Bucket([]byte(BUCKET_IPAM))

	ips := []IP{}
	bucket.ForEach(func(k, v []byte) error {
		var ip IP
		if err := json.Unmarshal(v, &ip); err != nil {
			return err
		}
		ips = append(ips, ip)
		return nil
	})

	return ips, nil
}

func (b *BoltStore) UpdateIP(ip IP) error {
	_, err := b.RetriveIP(ip.Key())
	if err != nil {
		return err
	}

	if err := b.SaveIP(ip); err != nil {
		return err
	}

	return nil
}

func (b *BoltStore) EmptyPool() error {
	tx, err := b.conn.Begin(true)
	if err != nil {
		return err
	}
	bucket := tx.Bucket([]byte(BUCKET_IPAM))

	bucket.ForEach(func(k, v []byte) error {
		bucket.Delete(k)
		return nil
	})

	return tx.Commit()
}
