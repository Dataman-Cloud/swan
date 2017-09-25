package janitor

import (
	"encoding/json"
	"net/http"

	"github.com/Dataman-Cloud/swan/agent/janitor/stats"
	"github.com/Dataman-Cloud/swan/agent/janitor/upstream"

	"github.com/gorilla/mux"
)

func (s *JanitorServer) ListUpstreams(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(upstream.AllUpstreams())
}

func (s *JanitorServer) GetUpstream(w http.ResponseWriter, r *http.Request) {
	var (
		uid = mux.Vars(r)["uid"]
		m   = upstream.AllUpstreams()
		ret = new(upstream.Upstream)
	)

	for idx, u := range m {
		if u.Name == uid {
			ret = m[idx]
			break
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ret)
}

func (s *JanitorServer) UpsertUpstream(w http.ResponseWriter, r *http.Request) {
	var cmb *upstream.BackendCombined
	if err := json.NewDecoder(r.Body).Decode(&cmb); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	if err := cmb.Valid(); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	if err := s.UpsertBackend(cmb); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *JanitorServer) DelUpstream(w http.ResponseWriter, r *http.Request) {
	var cmb *upstream.BackendCombined
	if err := json.NewDecoder(r.Body).Decode(&cmb); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	s.removeBackend(cmb)
	w.WriteHeader(http.StatusNoContent)
}

func (s *JanitorServer) ListSessions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(upstream.AllSessions())
}

func (s *JanitorServer) ShowConfigs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.config)
}

func (s *JanitorServer) ShowStats(w http.ResponseWriter, r *http.Request) {
	wrapper := map[string]interface{}{
		"httpd":    s.config.ListenAddr,
		"httpdTLS": s.config.TLSListenAddr,
		"counter":  stats.Get(),
		"tcpd":     s.tcpd,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wrapper)
}

func (s *JanitorServer) ShowUpstreamStats(w http.ResponseWriter, r *http.Request) {
	uid := mux.Vars(r)["uid"]
	if m, ok := stats.UpstreamStats()[uid]; ok {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(m)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(make(map[string]interface{}))
}

func (s *JanitorServer) ShowBackendStats(w http.ResponseWriter, r *http.Request) {
	var (
		vars = mux.Vars(r)
		uid  = vars["uid"]
		bid  = vars["bid"]
	)

	if ups, ok := stats.UpstreamStats()[uid]; ok {
		if backend, ok := ups[bid]; ok {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(backend)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(make(map[string]interface{}))
}
