package nameserver

import (
	"errors"
	"net"
	"regexp"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/miekg/dns"

	"github.com/Dataman-Cloud/swan/config"
)

const (
	GATEWAY = "gateway"
)

var (
	isDigitPrefix = regexp.MustCompile(`^[0-9]+\.`) // prefix with DigitsAndDot
)

type Resolver struct {
	config       *config.DNS
	base         string               // base domain suffix
	gwbase       string               // gateway base domain suffix
	m            map[string][]*Record // records store:  parents -> []records
	sync.RWMutex                      // protect m
	dnsClient    *dns.Client          // for forwarders
	forwardAddrs []string             // for forwarders
}

func NewResolver(cfg *config.DNS) *Resolver {
	base := cfg.Domain
	if !strings.HasSuffix(base, ".") {
		base = base + "."
	}

	resolver := &Resolver{
		config: cfg,
		base:   base,
		gwbase: GATEWAY + "." + base,
		m:      make(map[string][]*Record),
		dnsClient: &dns.Client{
			Net:          "udp",
			DialTimeout:  cfg.ExchangeTimeout,
			ReadTimeout:  cfg.ExchangeTimeout,
			WriteTimeout: cfg.ExchangeTimeout,
		},
	}

	resolver.forwardAddrs = make([]string, len(cfg.Resolvers))
	for i, addr := range cfg.Resolvers {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			host = addr
			port = "53"
		}
		resolver.forwardAddrs[i] = net.JoinHostPort(host, port)
	}

	return resolver
}

func (r *Resolver) allRecords() map[string][]*Record {
	r.RLock()
	defer r.RUnlock()
	return r.m
}

func (r *Resolver) upsert(record *Record) error {
	var (
		parent = record.Parent
		id     = record.ID
	)

	// verify & rewrite
	if err := record.rewrite(r.base); err != nil {
		log.Warnf("resolver veriy & rewrite record error: %v", err)
		return err
	}

	r.Lock()
	defer r.Unlock()

	records, ok := r.m[parent]
	if !ok {
		r.m[parent] = []*Record{record}
		return nil
	}

	var idx int = -1
	for i, record := range records {
		if record.ID == id {
			idx = i
			break
		}
	}

	if idx >= 0 {
		return nil
	}

	r.m[parent] = append(r.m[parent], record)
	return nil
}

func (r *Resolver) remove(record *Record) {
	var (
		parent = record.Parent
		id     = record.ID
	)

	r.Lock()
	defer r.Unlock()

	records, ok := r.m[parent]
	if !ok {
		return
	}

	var idx int = -1
	for i, record := range records {
		if record.ID == id {
			idx = i
			break
		}
	}

	if idx < 0 {
		return
	}

	r.m[parent] = append(r.m[parent][:idx], r.m[parent][idx+1:]...)
	if len(r.m[parent]) == 0 {
		delete(r.m, parent)
	}
}

func (r *Resolver) search(name string) []*Record {
	r.RLock()
	defer r.RUnlock()

	// dns query for gateway
	if strings.HasSuffix(name, r.gwbase) {
		return r.m["PROXY"]
	}

	parent := strings.TrimSuffix(name, "."+r.base)

	// by parent
	if !isDigitPrefix.MatchString(name) {
		return r.m[parent] // all sub records
	}

	// by index
	fields := strings.SplitN(parent, ".", 2)
	if len(fields) == 2 {
		parent = fields[1]
	}

	// specified index record
	for _, record := range r.m[parent] {
		if record.CleanName == name {
			return []*Record{record}
		}
	}
	return nil
}

func (r *Resolver) Start() error {
	dns.HandleFunc(r.base, r.handleLocal)
	dns.HandleFunc(".", r.handleForward)

	server := &dns.Server{
		Addr: r.config.ListenAddr,
		Net:  "udp",
	}

	return server.ListenAndServe()
}

func (r *Resolver) handleLocal(w dns.ResponseWriter, req *dns.Msg) {
	msg := &dns.Msg{
		MsgHdr: dns.MsgHdr{
			Authoritative:      true,
			RecursionAvailable: r.config.RecurseOn,
		}}
	msg.SetReply(req)

	var (
		name = strings.ToLower(req.Question[0].Name)
		ttl  = r.config.TTL
	)

	switch typ := req.Question[0].Qtype; typ {

	case dns.TypeA:
		for _, record := range r.search(name) {
			a := record.buildA(name, ttl)
			msg.Answer = append(msg.Answer, a)
		}

	case dns.TypeSRV:
		for _, record := range r.search(name) {
			srv, ext := record.buildSRV(name, ttl)
			msg.Answer = append(msg.Answer, srv)
			msg.Extra = append(msg.Extra, ext)
		}
	}

	if len(msg.Answer) == 0 {
		log.Warnf("resolve [%s] got non of matched records", name)
	} else {
		log.Debugf("resolve [%s] -> [%s]", name, msg.Answer)
	}

	// write reply whatever...
	if err := w.WriteMsg(msg); err != nil {
		log.Errorln("resolve [%s] error on dns reply: %v", name, err)
	}
}

func (r *Resolver) handleForward(w dns.ResponseWriter, req *dns.Msg) {
	m, err := r.Forward(req)
	if err != nil {
		log.Errorln("forwarder:", err)
		m = new(dns.Msg).SetRcode(req, dns.RcodeServerFailure)
	} else if len(m.Answer) == 0 {
		log.Debugf("forwarder: no answer found")
	}

	if err := w.WriteMsg(m); err != nil {
		log.Errorln(err)
	}
}

func (r *Resolver) Forward(req *dns.Msg) (reply *dns.Msg, err error) {
	if len(r.forwardAddrs) == 0 {
		err = errors.New("no avaliable forwarders")
		return
	}

	for _, addr := range r.forwardAddrs {
		reply, _, err = r.dnsClient.Exchange(req, addr)
		if err == nil {
			break
		}
	}

	return reply, err
}
