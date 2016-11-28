package health

import (
	"github.com/Dataman-Cloud/swan/src/types"
)

type Store interface {
	ListChecks() ([]*types.Check, error)
}
