package mesos

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/Dataman-Cloud/swan/agent/janitor/upstream"
	"github.com/Dataman-Cloud/swan/agent/resolver"
	"github.com/Dataman-Cloud/swan/mole"
	"github.com/Dataman-Cloud/swan/types"
)

func (s *Scheduler) ClusterAgents() map[string]*mole.ClusterAgent {
	return s.clusterMaster.Agents()
}

func (s *Scheduler) ClusterAgent(id string) *mole.ClusterAgent {
	return s.clusterMaster.Agent(id)
}

type broadcastRes struct {
	sync.Mutex
	m [][2]string // agent-id, errmsg
}

func (br *broadcastRes) Error() string {
	bs, _ := json.Marshal(br.m)
	return string(bs)
}

// sync calling agent Api to update agent's proxy & dns records on task healthy events.
func (s *Scheduler) broadcastEventRecords(ev *types.TaskEvent) error {
	var (
		res = &broadcastRes{m: make([][2]string, 0, 0)}
	)

	var wg sync.WaitGroup
	for _, agent := range s.ClusterAgents() {
		wg.Add(1)
		go func(agent *mole.ClusterAgent) {
			var err error

			defer func() {
				if err != nil {
					res.Lock()
					res.m = append(res.m, [2]string{agent.ID(), err.Error()})
					res.Unlock()
				}

				wg.Done()
			}()

			funcDoReq := func(req *http.Request) error {
				resp, err := agent.Client().Do(req)
				if err != nil {
					return err
				}
				defer resp.Body.Close()

				if code := resp.StatusCode; code >= 400 {
					bs, _ := ioutil.ReadAll(resp.Body)
					return fmt.Errorf("%d - %s", code, string(bs))
				}

				return nil
			}

			reqProxy, err := s.buildAgentProxyReq(ev)
			if err != nil {
				return
			}
			reqProxy.Close = true
			reqProxy.Header.Set("Connection", "close")
			reqProxy.Host = agent.ID()
			err = funcDoReq(reqProxy)
			if err != nil {
				return
			}

			reqDNS, err := s.buildAgentDNSReq(ev)
			if err != nil {
				return
			}
			reqDNS.Close = true
			reqDNS.Header.Set("Connection", "close")
			reqDNS.Host = agent.ID()
			err = funcDoReq(reqDNS)
			if err != nil {
				return
			}

		}(agent)
	}

	wg.Wait()

	if len(res.m) == 0 {
		return nil
	}

	return res
}

func (s *Scheduler) buildAgentDNSRecord(ev *types.TaskEvent) *resolver.Record {
	return &resolver.Record{
		ID:          ev.TaskID,
		Parent:      ev.AppID,
		IP:          ev.IP,
		Port:        fmt.Sprintf("%d", ev.Port),
		Weight:      ev.Weight,
		ProxyRecord: false,
	}
}

func (s *Scheduler) buildAgentDNSReq(ev *types.TaskEvent) (*http.Request, error) {
	body := s.buildAgentDNSRecord(ev)

	bs, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	switch typ := ev.Type; typ {
	case types.EventTypeTaskHealthy:
		return http.NewRequest("PUT", "http://xxx/dns/records", bytes.NewBuffer(bs))
	case types.EventTypeTaskUnhealthy:
		return http.NewRequest("DELETE", "http://xxx/dns/records", bytes.NewBuffer(bs))
	case types.EventTypeTaskWeightChange:
		return nil, nil
	default:
		return nil, errors.New("unknown event type: " + typ)
	}
}

func (s *Scheduler) buildAgentProxyRecord(ev *types.TaskEvent) *upstream.BackendCombined {
	return &upstream.BackendCombined{
		Upstream: &upstream.Upstream{
			Name:   ev.AppID,
			Alias:  ev.AppAlias,
			Listen: ev.AppListen,
			Sticky: ev.AppSticky,
		},
		Backend: &upstream.Backend{
			ID:        ev.TaskID,
			IP:        ev.IP,
			Port:      ev.Port,
			Scheme:    "",
			Version:   ev.VersionID,
			Weight:    ev.Weight,
			CleanName: "",
		},
	}
}

func (s *Scheduler) buildAgentProxyReq(ev *types.TaskEvent) (*http.Request, error) {
	body := s.buildAgentProxyRecord(ev)

	bs, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	switch typ := ev.Type; typ {
	case types.EventTypeTaskHealthy, types.EventTypeTaskWeightChange:
		return http.NewRequest("PUT", "http://xxx/proxy/upstreams", bytes.NewBuffer(bs))
	case types.EventTypeTaskUnhealthy:
		return http.NewRequest("DELETE", "http://xxx/proxy/upstreams", bytes.NewBuffer(bs))
	default:
		return nil, errors.New("unknown event type: " + typ)
	}
}
