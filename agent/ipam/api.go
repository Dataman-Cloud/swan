package ipam

import (
	"encoding/json"
	"net/http"
)

func (m *IPAM) ListSubNets(w http.ResponseWriter, r *http.Request) {
	subnets, err := m.store.ListSubNets()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// wrap each subnet with ips
	ret := make(map[string]interface{})
	for _, subnet := range subnets {
		ips, err := m.store.ListIPs(subnet.ID)
		if err != nil {
			ret[subnet.ID] = map[string]interface{}{
				"subnet": subnet,
				"ips":    nil,
				"err":    err.Error(),
			}
		} else {
			ret[subnet.ID] = map[string]interface{}{
				"subnet": subnet,
				"ips":    ips,
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ret)
}

func (m *IPAM) SetSubNetPool(w http.ResponseWriter, r *http.Request) {
	var pool *IPPoolRange
	if err := json.NewDecoder(r.Body).Decode(&pool); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	if err := pool.Valid(); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	if err := m.SetIPPool(pool); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}
