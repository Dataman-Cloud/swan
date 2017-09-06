package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

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

func (r *Server) queryAgentID(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var (
		ips = req.Form.Get("ips")
	)
	if ips == "" {
		http.Error(w, "query parameter ips required", http.StatusBadRequest)
		return
	}

	state, err := r.driver.MesosState()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, slave := range state.Slaves {
		for _, ip := range strings.Split(ips, ",") {
			if ip == slave.Hostname {
				w.Write([]byte(slave.ID))
				return
			}
		}
	}

	http.Error(w, "not found matched mesos slaves", http.StatusNotFound)
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

func (r *Server) redirectAgentDocker(w http.ResponseWriter, req *http.Request) {
	n := len(`/v1/agents/docker/`) + 39 // FIX LATER
	r.redirectAgent(n, w, req)
}

func (r *Server) redirectAgentProxy(w http.ResponseWriter, req *http.Request) {
	n := len(`/v1/agents/`) + 39
	r.redirectAgent(n, w, req)
}

func (r *Server) redirectAgentDNS(w http.ResponseWriter, req *http.Request) {
	n := len(`/v1/agents/`) + 39
	r.redirectAgent(n, w, req)
}

func (r *Server) redirectAgentIPAM(w http.ResponseWriter, req *http.Request) {
	n := len(`/v1/agents/`) + 39
	r.redirectAgent(n, w, req)
}

func (r *Server) getAppDNS(w http.ResponseWriter, req *http.Request) {
	var (
		appId = mux.Vars(req)["app_id"]
		ret   = make(map[string]interface{})
	)

	for id := range r.driver.ClusterAgents() {
		info, err := r.getAppDNSInfo(id, appId)
		if err != nil {
			ret[id] = err.Error()
		} else {
			ret[id] = info
		}
	}

	writeJSON(w, http.StatusOK, ret)
}

func (r *Server) getAppDNSTraffics(w http.ResponseWriter, req *http.Request) {
	var (
		appId = mux.Vars(req)["app_id"]
		ret   = make(map[string]interface{})
	)

	for id := range r.driver.ClusterAgents() {
		info, err := r.getAppDNSTrafficInfo(id, appId)
		if err != nil {
			ret[id] = err.Error()
		} else {
			ret[id] = info
		}
	}

	writeJSON(w, http.StatusOK, ret)
}

func (r *Server) getAppProxy(w http.ResponseWriter, req *http.Request) {
	var (
		appId = mux.Vars(req)["app_id"]
		ret   = make(map[string]interface{})
	)

	for id := range r.driver.ClusterAgents() {
		info, err := r.getAppProxyInfo(id, appId)
		if err != nil {
			ret[id] = err.Error()
		} else {
			ret[id] = info
		}
	}

	writeJSON(w, http.StatusOK, ret)
}

func (r *Server) getAppTraffics(w http.ResponseWriter, req *http.Request) {
	var (
		appId = mux.Vars(req)["app_id"]
		ret   = make(map[string]interface{})
	)

	for id := range r.driver.ClusterAgents() {
		info, err := r.getAppProxyTrafficInfo(id, appId)
		if err != nil {
			ret[id] = err.Error()
		} else {
			ret[id] = info
		}
	}

	writeJSON(w, http.StatusOK, ret)
}

// utils
//
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

func (r *Server) getAgentsListenings() []int64 {
	ls := make([]int64, 0)
	for id := range r.driver.ClusterAgents() {
		info, err := r.getAgentInfo(id)
		if err != nil {
			continue
		}
		ls = append(ls, info.Listenings...)
	}

	// make uniq
	seen := map[int64]bool{}
	for _, l := range ls {
		if _, ok := seen[l]; !ok {
			ls[len(seen)] = l
			seen[l] = true
		}
	}

	return ls[:len(seen)] // re-slice
}

func (r *Server) getAgentInfo(agentId string) (*types.SysInfo, error) {
	agentReq, _ := http.NewRequest("GET", fmt.Sprintf("http://%s/sysinfo", agentId), nil)
	var info *types.SysInfo
	err := r.requestAgentResource(agentId, agentReq, 200, &info)
	return info, err
}

func (r *Server) getAppDNSInfo(agentId, appId string) ([]interface{}, error) {
	agentReq, _ := http.NewRequest("GET", fmt.Sprintf("http://%s/dns/records/%s", agentId, appId), nil)
	var info []interface{}
	err := r.requestAgentResource(agentId, agentReq, 200, &info)
	return info, err
}

func (r *Server) getAppProxyInfo(agentId, appId string) (map[string]interface{}, error) {
	agentReq, _ := http.NewRequest("GET", fmt.Sprintf("http://%s/proxy/upstreams/%s", agentId, appId), nil)
	var info map[string]interface{}
	err := r.requestAgentResource(agentId, agentReq, 200, &info)
	return info, err
}

func (r *Server) getAppProxyTrafficInfo(agentId, appId string) (map[string]interface{}, error) {
	agentReq, _ := http.NewRequest("GET", fmt.Sprintf("http://%s/proxy/stats/%s", agentId, appId), nil)
	var info map[string]interface{}
	err := r.requestAgentResource(agentId, agentReq, 200, &info)
	return info, err
}

func (r *Server) getAppDNSTrafficInfo(agentId, appId string) (map[string]interface{}, error) {
	agentReq, _ := http.NewRequest("GET", fmt.Sprintf("http://%s/dns/stats/%s", agentId, appId), nil)
	var info map[string]interface{}
	err := r.requestAgentResource(agentId, agentReq, 200, &info)
	return info, err
}

func (r *Server) requestAgentResource(id string, req *http.Request, expectCode int, data interface{}) error {
	resp, err := r.proxyAgent(id, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != expectCode {
		bs, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("%d - %s", code, string(bs))
	}

	return json.NewDecoder(resp.Body).Decode(&data)
}
