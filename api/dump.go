package api

import "net/http"

func (r *Server) dump(w http.ResponseWriter, req *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"scheduler": r.driver.Dump(),
	})
}
