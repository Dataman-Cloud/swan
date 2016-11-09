package utils

import (
	"fmt"
	"net/http"
)

// CheckForJSON makes sure that the request's Content-Type is application/json.
func CheckForJSON(r *http.Request) error {
	if r.Header.Get("Content-Type") != "application/json" {
		return fmt.Errorf("Content-Type must be 'application/json'")
	}

	return nil
}
