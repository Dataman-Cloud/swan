package api

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

func (r *Router) listAgents(w http.ResponseWriter, req *http.Request) {
	ret := r.master.Agents()
	writeJSON(w, http.StatusOK, ret)
}

func (r *Router) getAgent(w http.ResponseWriter, req *http.Request) {
	var (
		id    = mux.Vars(req)["agent_id"]
		agent = r.master.Agent(id)
	)

	if agent == nil {
		http.Error(w, "no such agent: "+id, http.StatusNotFound)
		return
	}

	agentReq, _ := http.NewRequest("GET", fmt.Sprintf("http://%s/sysinfo", id), nil)
	r.proxyAgentHandle(id, agentReq, w)
}

func (r *Router) proxyAgentHandle(id string, req *http.Request, w http.ResponseWriter) {
	resp, err := r.proxyAgent(id, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (r *Router) proxyAgent(id string, req *http.Request) (*http.Response, error) {
	agent := r.master.Agent(id)
	if agent == nil {
		return nil, errors.New("no such agent: " + id)
	}

	// rewrite request
	req.Close = true
	req.Header.Set("Connection", "close")
	req.Host = id

	log.Printf("proxying agent request: %s", req.URL.String())

	client := agent.Client()
	return client.Do(req)
}
