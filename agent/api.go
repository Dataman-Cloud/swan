package agent

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func (agent *Agent) NewHTTPMux() http.Handler {
	m := mux.NewRouter()

	m.Path("/sysinfo").Methods("GET").HandlerFunc(agent.sysinfo)
	m.Path("/configs").Methods("GET").HandlerFunc(agent.showConfigs)

	// /proxy/**
	if agent.config.Janitor.Enabled {
		agent.setupProxyHandlers(m)
	}

	// /dns/**
	if agent.config.DNS.Enabled {
		agent.setupDNSHandlers(m)
	}

	// /ipam/**
	if agent.config.IPAM.Enabled {
		agent.setupIPAMHandlers(m)
	}

	m.NotFoundHandler = http.HandlerFunc(agent.serveProxy)

	return m
}

func (agent *Agent) setupProxyHandlers(mux *mux.Router) {
	var (
		janitor = agent.janitor
	)

	r := mux.PathPrefix("/proxy").Subrouter()
	r.Path("").Methods("GET").HandlerFunc(janitor.ListUpstreams)
	r.Path("/upstreams").Methods("GET").HandlerFunc(janitor.ListUpstreams)
	r.Path("/upstreams/{uid}").Methods("GET").HandlerFunc(janitor.GetUpstream)
	r.Path("/upstreams").Methods("PUT").HandlerFunc(janitor.UpsertUpstream)
	r.Path("/upstreams").Methods("DELETE").HandlerFunc(janitor.DelUpstream)
	r.Path("/sessions").Methods("GET").HandlerFunc(janitor.ListSessions)
	r.Path("/configs").Methods("GET").HandlerFunc(janitor.ShowConfigs)
	r.Path("/stats").Methods("GET").HandlerFunc(janitor.ShowStats)
	r.Path("/stats/{uid}").Methods("GET").HandlerFunc(janitor.ShowUpstreamStats)
	r.Path("/stats/{uid}/{bid}").Methods("GET").HandlerFunc(janitor.ShowBackendStats)
}

func (agent *Agent) setupDNSHandlers(mux *mux.Router) {
	var (
		resolver = agent.resolver
	)

	r := mux.PathPrefix("/dns").Subrouter()
	r.Path("").Methods("GET").HandlerFunc(resolver.ListRecords)
	r.Path("/records").Methods("GET").HandlerFunc(resolver.ListRecords)
	r.Path("/records/{id}").Methods("GET").HandlerFunc(resolver.GetRecord)
	r.Path("/records").Methods("PUT").HandlerFunc(resolver.UpsertRecord)
	r.Path("/records").Methods("DELETE").HandlerFunc(resolver.DelRecord)
	r.Path("/configs").Methods("GET").HandlerFunc(resolver.ShowConfigs)
	r.Path("/stats").Methods("GET").HandlerFunc(resolver.ShowStats)
	r.Path("/stats/{id}").Methods("GET").HandlerFunc(resolver.ShowParentStats)
}

func (agent *Agent) setupIPAMHandlers(mux *mux.Router) {
	var (
		ipam = agent.ipam
	)

	r := mux.PathPrefix("/ipam").Subrouter()
	r.Path("").Methods("GET").HandlerFunc(ipam.ListSubNets)
	r.Path("/subnets").Methods("GET").HandlerFunc(ipam.ListSubNets)
	r.Path("/subnets").Methods("PUT").HandlerFunc(ipam.SetSubNetPool)
}

func (agent *Agent) sysinfo(w http.ResponseWriter, r *http.Request) {
	info, err := Gather()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

func (agent *Agent) showConfigs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agent.config)
}
