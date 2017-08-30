package types

import (
	"errors"
)

type MesosLabel struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (l *MesosLabel) Validate() error {
	if l.Key == "" {
		return errors.New("label key can not be empty")
	}

	if l.Value == "" {
		return errors.New("label value can not be empty")
	}

	return nil
}
