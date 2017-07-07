package resolver

import (
	"errors"
	"net"
	"strconv"
	"strings"

	"github.com/miekg/dns"
)

type Record struct {
	ID          string  `json:"id"`
	Parent      string  `json:"parent"`
	IP          string  `json:"ip"`
	Port        string  `json:"port"`
	Weight      float64 `json:"weight"`
	ProxyRecord bool    `json:"proxy_record"`
	CleanName   string  `json:"clean_name"`

	ip    net.IP
	portN int
}

func (r *Record) rewrite(base string) error {
	ip := net.ParseIP(r.IP)
	if ip == nil {
		return errors.New("dns-record: invlaid IP: " + r.IP)
	}
	r.ip = ip

	port, err := strconv.Atoi(r.Port)
	if err != nil {
		return errors.New("dns-record: invalid Port: " + r.Port)
	}
	r.portN = port

	fields := strings.SplitN(r.ID, ".", 2)
	if len(fields) == 2 {
		r.CleanName = fields[1] + "." + base
	}

	return nil
}

func (r *Record) buildA(name string, ttl int) *dns.A {
	return &dns.A{
		Hdr: dns.RR_Header{
			Name:   name,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    uint32(ttl),
		},
		A: r.ip.To4(),
	}
}

func (r *Record) buildSRV(name string, ttl int) (*dns.SRV, *dns.A) {
	srv := &dns.SRV{
		Hdr: dns.RR_Header{
			Name:   name,
			Rrtype: dns.TypeSRV,
			Class:  dns.ClassINET,
			Ttl:    uint32(ttl),
		},
		Priority: 0,
		Weight:   uint16(r.Weight),
		Port:     uint16(r.portN),
		Target:   r.CleanName,
	}

	a := r.buildA(r.CleanName, ttl) // note: use clean name to build A

	return srv, a
}
