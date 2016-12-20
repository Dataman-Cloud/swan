package nameserver

import (
	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
)

type rrs map[string][]string

func (r rrs) del(name string, host string) bool {
	if host != "" {
		// remove one host in target r[name]
		hosts, ok := r[name]
		if !ok {
			return false
		} else {
			index := -1
			for i, h := range hosts {
				if h == host {
					index = i
					break
				}

			}
			if index > -1 {
				hosts = append(hosts[:index], hosts[index+1:]...)
			}
			return true
		}
	} else {
		delete(r, name)
		return true
	}
}

func (r rrs) add(name, host string) bool {
	logrus.Debugf("add new record for %s %s ", name, host)

	if host == "" {
		return false
	}
	var hosts []string
	hosts, ok := r[name]
	if !ok {
		hosts = append(hosts, host)
		r[name] = hosts
	} else {
		hostDuplicated := stringInSlice(host, hosts)
		if !hostDuplicated {
			hosts = append(hosts, host)
			r[name] = hosts
		}
	}
	return true
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func (rg *RecordGenerator) WatchEvent(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case e := <-rg.RecordGeneratorChangeChan:
			if e.Change == "add" {
				aDomain := e.DomainPrefix + "." + rg.Domain + "."
				if e.Type == "srv" {
					rg.SRVAs.add(aDomain, e.Ip)
					rg.SRVs.add(aDomain, aDomain+":"+e.Port)
				}
				if e.Type == "a" {
					//for proxy rr
					rg.As.add(aDomain, e.Ip)
				}
			}

			if e.Change == "del" {
				aDomain := e.DomainPrefix + "." + rg.Domain + "."
				if e.Type == "srv" {
					rg.SRVAs.del(aDomain, "")
					rg.SRVs.del(aDomain, "")
				}
				if e.Type == "a" {
					rg.As.del(aDomain, e.Ip)
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
	SRVAs                     rrs //a rrs for srv type
	SRVs                      rrs
	SlaveIPs                  map[string]string
	RecordGeneratorChangeChan chan *RecordGeneratorChangeEvent
}
