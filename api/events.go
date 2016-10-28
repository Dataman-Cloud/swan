package api

import (
	"net/http"
)

// EventStream is used to retrieve scheduler events as SSE format.
func (r *Router) EventStream(w http.ResponseWriter, req *http.Request) error {
	if err := r.backend.EventStream(w); err != nil {
		return err
	}

	return nil
}
