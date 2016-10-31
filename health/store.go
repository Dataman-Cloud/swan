package health

import (
	"github.com/Dataman-Cloud/swan/types"
)

type Store interface {
	ListChecks() ([]*types.Check, error)
}
