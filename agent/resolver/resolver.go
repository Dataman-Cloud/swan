package resolver

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

var (
	isDigitPrefix = regexp.MustCompile(`^[0-9]+\.`) // prefix with DigitsAndDot
)

type Resolver struct {
	config           *config.DNS
	base             string               // base domain suffix
	gwbase           string               // gateway base domain suffix
	m                map[string][]*Record // records store:  parents -> []records
	stats            *Stats               // stats & traffic
	sync.RWMutex                          // protect m
	dnsClient        *dns.Client          // for forwarders
	forwardAddrs     []string             // for forwarders
	proxyAdvertiseIP string               // for gateway.{base.domain} resolve request
}

func NewResolver(cfg *config.DNS, AdvertiseIP string) *Resolver {
	base := cfg.Domain
	if !strings.HasSuffix(base, ".") {
		base = base + "."
	}

	resolver := &Resolver{
		config: cfg,
		base:   base,
		gwbase: base,
		m:      make(map[string][]*Record),
		stats:  newStats(),
		dnsClient: &dns.Client{
			Net:          "udp",
			DialTimeout:  cfg.ExchangeTimeout,
			ReadTimeout:  cfg.ExchangeTimeout,
			WriteTimeout: cfg.ExchangeTimeout,
		},
		proxyAdvertiseIP: AdvertiseIP,
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

func (r *Resolver) Start() error {
	log.Println("agent resolver in serving ...")

	// init with local proxy dns record
	rr := &Record{
		ID:          "local_proxy",
		Parent:      "PROXY",
		IP:          r.proxyAdvertiseIP, // we've verified before
		Port:        "80",
		Weight:      0,
		ProxyRecord: true,
	}
	r.Upsert(rr)

	// serving
	dns.HandleFunc(r.base, r.handleLocal)
	dns.HandleFunc(".", r.handleForward)

	server := &dns.Server{
		Addr: r.config.ListenAddr,
		Net:  "udp",
	}

	return server.ListenAndServe()
}

func (r *Resolver) handleLocal(w dns.ResponseWriter, req *dns.Msg) {
	var (
		parent string
		delta  = &Counter{Requests: 1, Authority: 1, Forward: 0}
	)

	defer func() {
		if parent != "" {
			r.stats.Incr(parent, delta)
		}
	}()

	msg := &dns.Msg{
		MsgHdr: dns.MsgHdr{
			Authoritative:      true,
			RecursionAvailable: r.config.RecurseOn,
		},
	}
	msg.SetReply(req)

	var (
		name    = strings.ToLower(req.Question[0].Name)
		ttl     = r.config.TTL
		records []*Record
	)

	switch typ := req.Question[0].Qtype; typ {

	case dns.TypeA:
		parent, records = r.search(name)
		for _, record := range records {
			a := record.buildA(name, ttl)
			msg.Answer = append(msg.Answer, a)
		}
		delta.TypeA = 1

	case dns.TypeSRV:
		parent, records = r.search(name)
		for _, record := range records {
			srv, ext := record.buildSRV(name, ttl)
			msg.Answer = append(msg.Answer, srv)
			msg.Extra = append(msg.Extra, ext)
		}
		delta.TypeSRV = 1
	}

	if len(msg.Answer) == 0 {
		delta.Fails = 1
		log.Warnf("resolve [%s] got non of matched records", name)
	} else {
		log.Debugf("resolve [%s] -> [%s]", name, msg.Answer)
	}

	// write reply whatever...
	if err := w.WriteMsg(msg); err != nil {
		delta.Fails = 1
		log.Errorln("resolve [%s] error on dns reply: %v", name, err)
	}

}

func (r *Resolver) handleForward(w dns.ResponseWriter, req *dns.Msg) {
	var (
		delta = &Counter{Requests: 1, Authority: 0, Forward: 1}
	)

	defer func() {
		r.stats.Incr("", delta)
	}()

	m, err := r.Forward(req)
	if err != nil {
		delta.Fails = 1
		log.Errorln("forwarder:", err)
		m = new(dns.Msg).SetRcode(req, dns.RcodeServerFailure)
	} else if len(m.Answer) == 0 {
		log.Debugf("forwarder: no answer found")
	}

	if err := w.WriteMsg(m); err != nil {
		delta.Fails = 1
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

func (r *Resolver) allRecords() map[string][]*Record {
	r.RLock()
	defer r.RUnlock()
	return r.m
}

func (r *Resolver) Upsert(record *Record) error {
	log.Printf("dns upserting record: %s", record)

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
		r.m[parent][idx] = record
		return nil
	}

	r.m[parent] = append(r.m[parent], record)
	return nil
}

func (r *Resolver) remove(record *Record) (onLast bool) {
	log.Printf("dns removing record: %s", record)

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
		onLast = true
		delete(r.m, parent)
	}

	return
}

func (r *Resolver) search(name string) (string, []*Record) {
	r.RLock()
	defer r.RUnlock()

	// dns query for gateway
	if strings.HasSuffix(name, r.gwbase) {
		return "PROXY", r.m["PROXY"]
	}

	parent := strings.TrimSuffix(name, "."+r.base)

	// by parent
	if !isDigitPrefix.MatchString(name) {
		return parent, r.m[parent] // all sub records
	}

	// by index
	fields := strings.SplitN(parent, ".", 2)
	if len(fields) == 2 {
		parent = fields[1]
	}

	// specified index record
	for _, record := range r.m[parent] {
		if record.CleanName == name {
			return parent, []*Record{record}
		}
	}

	return "", nil
}
