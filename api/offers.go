package api

import "net/http"

func (s *Server) offers(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.driver.Offers())
}
