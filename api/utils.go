package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// WriteJSON write response as json format.
func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.Encode(v)
}

// CheckForJSON makes sure that the request's Content-Type is application/json.
func checkForJSON(req *http.Request) error {
	if req.Header.Get("Content-Type") != "application/json" {
		return fmt.Errorf("Content-Type must be 'application/json'")
	}

	return nil
}

func decode(b io.ReadCloser, v interface{}) error {
	dec := json.NewDecoder(b)

	return dec.Decode(&v)
}
