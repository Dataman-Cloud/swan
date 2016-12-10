package nameserver

import (
	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
)

type rrs map[string]map[string]struct{}

func (r rrs) del(name string) bool {
	delete(r, name)
	return true
}

func (r rrs) add(name, host string) bool {
	logrus.Debugf("add new record for %s %s ", name, host)

	if host == "" {
		return false
	}
	v, ok := r[name]
	if !ok {
		v = make(map[string]struct{})
		r[name] = v
	} else {
		// don't overwrite existing values
		_, ok = v[host]
		if ok {
			return false
		}
	}
	v[host] = struct{}{}
	return true
}

func (r rrs) First(name string) (string, bool) {
	for host := range r[name] {
		return host, true
	}
	return "", false
}

type rrsKind string

const (
	// A record types
	A rrsKind = "A"
	// SRV record types
	SRV = "SRV"
)

func (kind rrsKind) rrs(rg *RecordGenerator) rrs {
	switch kind {
	case A:
		return rg.As
	case SRV:
		return rg.SRVs
	default:
		return nil
	}
}

func (rg *RecordGenerator) WatchEvent(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case e := <-rg.RecordGeneratorChangeChan:
			if e.Change == "add" {
				aDomain := e.DomainPrefix + "." + rg.Domain + "."
				rg.As.add(aDomain, e.Ip)

				if e.Type == "srv" {
					rg.SRVs.add(aDomain, aDomain+":"+e.Port)
				}
			}

			if e.Change == "del" {
				aDomain := e.DomainPrefix + "." + rg.Domain + "."
				rg.As.del(aDomain)

				if e.Type == "srv" {
					rg.SRVs.del(aDomain)
				}
			}
		}
	}
}

// RecordGenerator contains DNS records and methods to access and manipulate
// them. TODO(kozyraki): Refactor when discovery id is available.
type RecordGenerator struct {
	Domain                    string
	As                        rrs
	SRVs                      rrs
	SlaveIPs                  map[string]string
	RecordGeneratorChangeChan chan *RecordGeneratorChangeEvent
}
