package ipam

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Dataman-Cloud/swan/api/utils"
	ipamanger "github.com/Dataman-Cloud/swan/ipam"
)

func (r *Router) AllocateIP(w http.ResponseWriter, req *http.Request) error {
	if err := utils.CheckForJSON(req); err != nil {
		return err
	}

	if err := req.ParseForm(); err != nil {
		return err
	}

	ipStr := req.Form.Get("ip")
	if ipStr == "" {
		return errors.New("no ip specified")
	}

	ip, err := r.ipam.AllocateIp(ipamanger.IP{Ip: ipStr})
	if err != nil {
		return err
	}

	return json.NewEncoder(w).Encode(ip)
}

func (r *Router) AllocateNextAvailable(w http.ResponseWriter, req *http.Request) error {
	if err := utils.CheckForJSON(req); err != nil {
		return err
	}

	ip, err := r.ipam.AllocateNextAvailableIP()
	if err != nil {
		return err
	}

	return json.NewEncoder(w).Encode(ip)
}

func (r *Router) ListAvailableIps(w http.ResponseWriter, req *http.Request) error {
	if err := utils.CheckForJSON(req); err != nil {
		return err
	}

	list, err := r.ipam.IPsAvailable()
	if err != nil {
		return err
	}

	return json.NewEncoder(w).Encode(list)
}

func (r *Router) ListAllocatedIps(w http.ResponseWriter, req *http.Request) error {
	if err := utils.CheckForJSON(req); err != nil {
		return err
	}

	list, err := r.ipam.IPsAllocated()
	if err != nil {
		return err
	}

	return json.NewEncoder(w).Encode(list)
}

func (r *Router) ReleaseIP(w http.ResponseWriter, req *http.Request) error {
	if err := utils.CheckForJSON(req); err != nil {
		return err
	}

	if err := req.ParseForm(); err != nil {
		return err
	}

	var param struct {
		IP string `json:"ip"`
	}

	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&param); err != nil {
		return err
	}

	err := r.ipam.Release(ipamanger.IP{Ip: param.IP})
	if err != nil {
		return err
	}

	return nil
}

func (r *Router) RefillIPs(w http.ResponseWriter, req *http.Request) error {
	if err := utils.CheckForJSON(req); err != nil {
		return err
	}

	if err := req.ParseForm(); err != nil {
		return err
	}

	var param struct {
		IPs []string `json:"ips"`
	}

	var ips []ipamanger.IP
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&param); err != nil {
		return err
	}

	for _, ipStr := range param.IPs {
		ips = append(ips, ipamanger.IP{Ip: ipStr})
	}

	err := r.ipam.Refill(ips)
	if err != nil {
		return err
	}

	return nil
}

func (r *Router) ListIPs(w http.ResponseWriter, req *http.Request) error {
	if err := utils.CheckForJSON(req); err != nil {
		return err
	}

	list, err := r.ipam.AllIPs()
	if err != nil {
		return err
	}

	return json.NewEncoder(w).Encode(list)
}
