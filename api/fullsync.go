package api

import "net/http"

func (r *Server) fullEventsAndRecords(w http.ResponseWriter, req *http.Request) {
	ret := r.driver.FullTaskEventsAndRecords()
	writeJSON(w, http.StatusOK, ret)
}
