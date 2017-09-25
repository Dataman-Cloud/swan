package resolver

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Resolver) ListRecords(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.allRecords())
}

func (s *Resolver) GetRecord(w http.ResponseWriter, r *http.Request) {
	var (
		vars = mux.Vars(r)
		id   = vars["id"]
		m    = s.allRecords()
		ret  = make([]*Record, 0)
	)
	if val, ok := m[id]; ok {
		ret = val
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ret)
}

func (s *Resolver) UpsertRecord(w http.ResponseWriter, r *http.Request) {
	var record *Record
	if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	if err := s.Upsert(record); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(http.StatusCreated)

}

func (s *Resolver) DelRecord(w http.ResponseWriter, r *http.Request) {
	var record *Record
	if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	if s.remove(record) {
		s.stats.Del(record.Parent)
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Resolver) ShowConfigs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.config)
}

func (s *Resolver) ShowStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.stats.Get())
}

func (s *Resolver) ShowParentStats(w http.ResponseWriter, r *http.Request) {
	pid := mux.Vars(r)["id"]
	m := s.stats.Get()
	if m, ok := m.Parents[pid]; ok {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(m)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(make(map[string]interface{}))
}
