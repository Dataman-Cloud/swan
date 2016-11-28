package mock

import (
	"github.com/Dataman-Cloud/swan/src/types"
)

type Store struct {
}

func (s *Store) ListChecks() ([]*types.Check, error) {
	return nil, nil
}
