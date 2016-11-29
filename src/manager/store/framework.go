package store

func (store *ManagerStore) SaveFrameworkID(frameworkId string) error {
	tx, err := store.BoltbDb.Begin(true)
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

func (store *ManagerStore) FetchFrameworkID() (string, error) {
	tx, err := store.BoltbDb.Begin(false)
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

func (store *ManagerStore) HasFrameworkID() (bool, error) {
	tx, err := store.BoltbDb.Begin(false)
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
