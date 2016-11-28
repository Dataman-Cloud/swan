package boltdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Dataman-Cloud/swan/src/types"
)

func (b *BoltStore) SaveVersion(version *types.Version) error {
	tx, err := b.conn.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bucket := tx.Bucket([]byte("versions"))

	data, err := json.Marshal(version)
	if err != nil {
		return err
	}

	if err := bucket.Put([]byte(fmt.Sprintf("%d", time.Now().UnixNano())), data); err != nil {
		return err
	}

	return tx.Commit()
}

func (b *BoltStore) ListVersions(appId string) ([]string, error) {
	tx, err := b.conn.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	bucket := tx.Bucket([]byte("versions"))

	var versionList []string

	bucket.ForEach(func(k, v []byte) error {
		var version types.Version
		if err := json.Unmarshal(v, &version); err != nil {
			return err
		}
		if version.ID == appId {
			versionList = append(versionList, string(k[:]))
		}

		return nil
	})

	return versionList, nil
}

func (b *BoltStore) FetchVersion(versionId string) (*types.Version, error) {
	tx, err := b.conn.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	bucket := tx.Bucket([]byte("versions"))

	data := bucket.Get([]byte(versionId))
	if data == nil {
		return nil, errors.New("Not Found")
	}

	var version types.Version
	if err := json.Unmarshal(data, &version); err != nil {
		return nil, err
	}

	return &version, nil
}

func (b *BoltStore) DeleteVersion(versionId string) error {
	tx, err := b.conn.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bucket := tx.Bucket([]byte("versions"))

	if err := bucket.Delete([]byte(versionId)); err != nil {
		return err
	}

	return tx.Commit()
}
