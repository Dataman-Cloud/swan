package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/Dataman-Cloud/swan/types"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

func (r *Server) listAgents(w http.ResponseWriter, req *http.Request) {
	var ret = map[string]interface{}{}
	for id := range r.driver.ClusterAgents() {
		info, err := r.getAgentInfo(id)
		if err != nil {
			ret[id] = err.Error()
		} else {
			ret[id] = info
		}
	}
	writeJSON(w, http.StatusOK, ret)
}

func (r *Server) getAgent(w http.ResponseWriter, req *http.Request) {
	var (
		id = mux.Vars(req)["agent_id"]
	)

	info, err := r.getAgentInfo(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, info)
}

func (r *Server) closeAgent(w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["agent_id"]
	r.driver.CloseClusterAgent(id)
}

func (r *Server) getAgentConfigs(w http.ResponseWriter, req *http.Request) {
	var (
		id    = mux.Vars(req)["agent_id"]
		agent = r.driver.ClusterAgent(id)
	)

	if agent == nil {
		http.Error(w, "no such agent: "+id, http.StatusNotFound)
		return
	}

	agentReq, _ := http.NewRequest("GET", fmt.Sprintf("http://%s/configs", id), nil)
	r.proxyAgentHandle(id, agentReq, w)
}

func (r *Server) fullEventsAndRecords(w http.ResponseWriter, req *http.Request) {
	ret := r.driver.FullTaskEventsAndRecords()
	writeJSON(w, http.StatusOK, ret)
}

func (r *Server) redirectAgentDocker(w http.ResponseWriter, req *http.Request) {
	n := len(`/v1/agents/docker/`) + 16
	r.redirectAgent(n, w, req)
}

func (r *Server) redirectAgentProxy(w http.ResponseWriter, req *http.Request) {
	n := len(`/v1/agents/`) + 16
	r.redirectAgent(n, w, req)
}

func (r *Server) redirectAgentDNS(w http.ResponseWriter, req *http.Request) {
	n := len(`/v1/agents/`) + 16
	r.redirectAgent(n, w, req)
}

func (r *Server) redirectAgentIPAM(w http.ResponseWriter, req *http.Request) {
	n := len(`/v1/agents/`) + 16
	r.redirectAgent(n, w, req)
}

func (r *Server) redirectAgent(stripN int, w http.ResponseWriter, req *http.Request) {
	var (
		id    = mux.Vars(req)["agent_id"]
		agent = r.driver.ClusterAgent(id)
	)

	if agent == nil {
		http.Error(w, "no such agent: "+id, http.StatusNotFound)
		return
	}

	// rewrite & proxy original request to agent docker remote api
	req.URL.Scheme = "http"
	req.URL.Host = id
	req.URL.Path = req.URL.Path[stripN:]
	req.RequestURI = "" // otherwise: http: Request.RequestURI can't be set in client requests.

	r.proxyAgentHandle(id, req, w)
}

func (r *Server) proxyAgentHandle(id string, req *http.Request, w http.ResponseWriter) {
	resp, err := r.proxyAgent(id, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (r *Server) proxyAgent(id string, req *http.Request) (*http.Response, error) {
	agent := r.driver.ClusterAgent(id)
	if agent == nil {
		return nil, errors.New("no such agent: " + id)
	}

	// rewrite request
	req.Close = true
	req.Header.Set("Connection", "close")
	req.Host = id

	log.Printf("proxying agent request: %s", req.URL.String())

	return agent.Client().Do(req)
}

func (r *Server) getAgentInfo(id string) (*types.SysInfo, error) {
	agentReq, _ := http.NewRequest("GET", fmt.Sprintf("http://%s/sysinfo", id), nil)
	resp, err := r.proxyAgent(id, agentReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("%d - %s", code, string(bs))
	}

	var info *types.SysInfo
	err = json.NewDecoder(resp.Body).Decode(&info)
	return info, err
}
