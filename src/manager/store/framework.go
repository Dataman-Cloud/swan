package store

func (store *ManagerStore) SaveFrameworkID(frameworkId string) error {
	return nil
}

func (store *ManagerStore) FetchFrameworkID() (string, error) {
	return "", nil
}

func (store *ManagerStore) HasFrameworkID() (bool, error) {
	frameworkId, err := store.FetchFrameworkID()
	if err != nil {
		return false, err
	}

	if frameworkId == "" {
		return false, nil
	}

	return true, nil
}
