package utils

import "fmt"

type SwanErrorSeverity int

const (
	SeverityLow SwanErrorSeverity = iota
	SeverityMiddle
	SeverityHigh
)

type SwanError struct {
	Severity SwanErrorSeverity `json:"Severity"`
	Err      string            `json:"Err"`
}

func NewError(severity SwanErrorSeverity, err interface{}) error {
	var finalError string
	switch v := err.(type) {
	case string:
		finalError = v
	case error:
		finalError = v.Error()
	default:
		finalError = fmt.Sprintf("%v", v)
	}
	return &SwanError{severity, finalError}
}

func (e *SwanError) Error() string {
	return e.Err
}
