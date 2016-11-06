package boltdb

import ()

func (b *BoltStore) SaveFrameworkID(frameworkId string) error {
	tx, err := b.conn.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bucket := tx.Bucket([]byte("swan"))
	if err := bucket.Put([]byte("frameworkId"), []byte(frameworkId)); err != nil {
		return err
	}

	return tx.Commit()
}

func (b *BoltStore) FetchFrameworkID() (string, error) {
	tx, err := b.conn.Begin(false)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	bucket := tx.Bucket([]byte("swan"))
	val := bucket.Get([]byte("frameworkId"))

	if val == nil {
		return "", nil
	}
	return string(val[:]), nil
}

func (b *BoltStore) HasFrameworkID() (bool, error) {
	tx, err := b.conn.Begin(false)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	bucket := tx.Bucket([]byte("swan"))
	val := bucket.Get([]byte("frameworkId"))

	if val != nil {
		return true, nil
	}

	return false, nil
}
