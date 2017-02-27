package utils

import (
	"errors"
)

type SwanErrorSeverity int

const (
	SeverityLow SwanErrorSeverity = iota
	SeverityMiddle
	SeverityHigh
)

type SwanError struct {
	Severity SwanErrorSeverity `json:"Severity"`
	Err      error             `json:"Err"`
}

func NewError(severity SwanErrorSeverity, err error) error {
	return &SwanError{
		Severity: severity,
		Err:      err,
	}
}

func NewErrorFromString(severity SwanErrorSeverity, message string) error {
	return &SwanError{
		Severity: severity,
		Err:      errors.New(message),
	}
}

func (e *SwanError) Error() string {
	return e.Err.Error()
}
