package mock

import (
	"github.com/Dataman-Cloud/swan/types"
)

type Store struct {
}

func (s *Store) ListChecks() ([]*types.Check, error) {
	return nil, nil
}
