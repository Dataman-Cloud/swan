package nameserver

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/miekg/dns"
	"golang.org/x/net/context"
)

func NewResolver(config *Config) *Resolver {
	res := &Resolver{
		config: config,
		stopC:  make(chan struct{}),
	}

	rr := RecordGenerator{
		RecordGeneratorChangeChan: make(chan *RecordGeneratorChangeEvent, 1),
	}

	rr.Domain = res.config.Domain
	rr.As = make(map[string]map[string]struct{})
	rr.SRVs = make(map[string]map[string]struct{})

	res.rs = &rr
	res.defaultFwd = NewForwarder(config.Resolvers, exchangers(config.ExchangeTimeout, "udp"))

	go func() {
		res.rs.WatchEvent(context.Background())
	}()

	return res
}

func (res *Resolver) RecordGeneratorChangeChan() chan *RecordGeneratorChangeEvent {
	return res.rs.RecordGeneratorChangeChan
}

func (res *Resolver) Start(ctx context.Context) error {
	return <-res.Run(ctx)
}

func (res *Resolver) Run(ctx context.Context) <-chan error {
	dns.HandleFunc(res.config.Domain+".", res.HandleSwan)
	dns.HandleFunc(".", res.HandleNonSwan(res.defaultFwd))

	startedCh, errCh := res.Serve()
	<-startedCh // when successfully listen

	go func() {
		for {
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			}
		}
	}()

	return errCh
}

type Resolver struct {
	config *Config

	rs         *RecordGenerator
	defaultFwd Forwarder
	stopC      chan struct{}
	startedC   chan struct{}
}

func (res *Resolver) HandleSwan(w dns.ResponseWriter, r *dns.Msg) {
	m := &dns.Msg{MsgHdr: dns.MsgHdr{
		Authoritative:      true,
		RecursionAvailable: res.config.RecurseOn,
	}}
	m.SetReply(r)

	var errs multiError
	rs := res.records()
	name := strings.ToLower(cleanWild(r.Question[0].Name))

	logrus.Debugf("resolve dns hostname %s", name)

	switch r.Question[0].Qtype {
	case dns.TypeSRV:
		errs.Add(res.handleSRV(rs, name, m, r))
	case dns.TypeA:
		errs.Add(res.handleA(rs, name, m))
	case dns.TypeSOA:
		errs.Add(res.handleSOA(m, r))
	case dns.TypeNS:
		errs.Add(res.handleNS(m, r))
	case dns.TypeANY:
		errs.Add(
			res.handleSRV(rs, name, m, r),
			res.handleA(rs, name, m),
			res.handleSOA(m, r),
			res.handleNS(m, r),
		)
	}

	if len(m.Answer) == 0 {
		errs.Add(res.handleEmpty(rs, name, m, r))
	}

	if !errs.Nil() {
		logrus.Errorf(errs.Error())
	}

	reply(w, m)
}

func (res *Resolver) records() *RecordGenerator {
	return res.rs
}

func (res *Resolver) handleSRV(rs *RecordGenerator, name string, m, r *dns.Msg) error {
	var errs multiError
	added := map[string]struct{}{} // track the A RR's we've already added, avoid dups

	for srv := range rs.SRVs[name] {
		srvRR, err := res.formatSRV(r.Question[0].Name, srv)
		if err != nil {
			errs.Add(err)
			continue
		}

		m.Answer = append(m.Answer, srvRR)
		host := strings.Split(srv, ":")[0]
		if _, found := added[host]; found {
			continue
		}

		if len(rs.As[host]) == 0 {
			continue
		}

		if a, ok := rs.As.First(host); ok {
			aRR, err := res.formatA(host, a)
			if err != nil {
				errs.Add(err)
				continue
			}
			m.Extra = append(m.Extra, aRR)
			added[host] = struct{}{}
		}
	}
	return errs
}

func (res *Resolver) handleA(rs *RecordGenerator, name string, m *dns.Msg) error {
	var errs multiError
	for a := range rs.As[name] {
		rr, err := res.formatA(name, a)
		if err != nil {
			errs.Add(err)
			continue
		}
		m.Answer = append(m.Answer, rr)
	}
	return errs
}

func (res *Resolver) handleSOA(m, r *dns.Msg) error {
	m.Ns = append(m.Ns, res.formatSOA(r.Question[0].Name))
	return nil
}

func (res *Resolver) handleNS(m, r *dns.Msg) error {
	m.Ns = append(m.Ns, res.formatNS(r.Question[0].Name))
	return nil
}

func (res *Resolver) handleEmpty(rs *RecordGenerator, name string, m, r *dns.Msg) error {
	qType := r.Question[0].Qtype
	switch qType {
	case dns.TypeSOA, dns.TypeNS, dns.TypeSRV:
		return nil
	}

	m.Rcode = dns.RcodeNameError

	// Because we don't implement AAAA records, AAAA queries will always
	// go via this path
	// Unfortunately, we don't implement AAAA queries in Swan-DNS,
	// and although the 'Not Implemented' error code seems more suitable,
	// RFCs do not recommend it: https://tools.ietf.org/html/rfc4074
	// Therefore we always return success, which is synonymous with NODATA
	// to get a positive cache on no records AAAA
	// Further information:
	// PR: https://github.com/mesosphere/mesos-dns/pull/366
	// Issue: https://github.com/mesosphere/mesos-dns/issues/363

	// The second component is just a matter of returning NODATA if we have
	// SRV or A records for the given name, but no neccessarily the given query

	if (qType == dns.TypeAAAA) || (len(rs.SRVs[name])+len(rs.As[name]) > 0) {
		m.Rcode = dns.RcodeSuccess
	}

	m.Ns = append(m.Ns, res.formatSOA(r.Question[0].Name))

	return nil
}

func (res *Resolver) HandleNonSwan(fwd Forwarder) func(dns.ResponseWriter, *dns.Msg) {
	return func(w dns.ResponseWriter, r *dns.Msg) {
		m, err := fwd(r, w.RemoteAddr().Network())
		if err != nil {
			m = new(dns.Msg).SetRcode(r, rcode(err))
		} else if len(m.Answer) == 0 {
			logrus.Infof("no answer found")
		}
		reply(w, m)
	}
}

// reply writes the given dns.Msg out to the given dns.ResponseWriter,
// compressing the message first and truncating it accordingly.
func reply(w dns.ResponseWriter, m *dns.Msg) {
	//m.Compress = true // https://github.com/mesosphere/mesos-dns/issues/{170,173,174}

	//if err := w.WriteMsg(truncate(m, isUDP(w))); err != nil {
	if err := w.WriteMsg(m); err != nil {
		logrus.Errorf("%s", err)
	}
}

func cleanWild(name string) string {
	if strings.Contains(name, ".*") {
		return strings.Replace(name, ".*", "", -1)
	}
	return name
}

func exchangers(timeout time.Duration, protos ...string) map[string]Exchanger {
	exs := make(map[string]Exchanger, len(protos))
	for _, proto := range protos {
		exs[proto] = Decorate(
			&dns.Client{
				Net:          proto,
				DialTimeout:  timeout,
				ReadTimeout:  timeout,
				WriteTimeout: timeout,
			},
		)
	}
	return exs
}

func (res *Resolver) Serve() (startedCh <-chan struct{}, errCh chan error) {
	ch := make(chan struct{})
	server := &dns.Server{
		Addr:              net.JoinHostPort(res.config.Listener, strconv.Itoa(res.config.Port)),
		Net:               "udp",
		TsigSecret:        nil,
		NotifyStartedFunc: func() { close(ch) },
	}

	go func() {
		defer close(errCh)
		err := server.ListenAndServe()
		if err != nil {
			errCh <- fmt.Errorf("Failed to setup %q server: %v", err)
		}
	}()

	return ch, errCh
}

type multiError []error

func (e *multiError) Add(err ...error) {
	for _, e1 := range err {
		if me, ok := e1.(multiError); ok {
			*e = append(*e, me...)
		} else if e1 != nil {
			*e = append(*e, e1)
		}
	}
}

func (e multiError) Error() string {
	errs := make([]string, len(e))
	for i := range errs {
		if e[i] != nil {
			errs[i] = e[i].Error()
		}
	}
	return strings.Join(errs, "; ")
}

func (e multiError) Nil() bool {
	for _, err := range e {
		if err != nil {
			return false
		}
	}
	return true
}

func (res *Resolver) formatSRV(name string, target string) (*dns.SRV, error) {
	ttl := uint32(res.config.TTL)

	h, port, err := net.SplitHostPort(target)
	if err != nil {
		return nil, errors.New("invalid target")
	}
	p, err := strconv.Atoi(port)
	if err != nil {
		return nil, errors.New("invalid target port")
	}

	return &dns.SRV{
		Hdr: dns.RR_Header{
			Name:   name,
			Rrtype: dns.TypeSRV,
			Class:  dns.ClassINET,
			Ttl:    ttl,
		},
		Priority: 0,
		Weight:   0,
		Port:     uint16(p),
		Target:   h,
	}, nil
}

func (res *Resolver) formatA(dom string, target string) (*dns.A, error) {
	ttl := uint32(res.config.TTL)

	a := net.ParseIP(target)
	if a == nil {
		return nil, errors.New("invalid target")
	}

	return &dns.A{
		Hdr: dns.RR_Header{
			Name:   dom,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    ttl},
		A: a.To4(),
	}, nil
}

func (res *Resolver) formatSOA(dom string) *dns.SOA {
	ttl := uint32(res.config.TTL)

	return &dns.SOA{
		Hdr: dns.RR_Header{
			Name:   dom,
			Rrtype: dns.TypeSOA,
			Class:  dns.ClassINET,
			Ttl:    ttl,
		},
		Ns:      res.config.SOARname,
		Mbox:    res.config.SOAMname,
		Serial:  atomic.LoadUint32(&res.config.SOASerial),
		Refresh: res.config.SOARefresh,
		Retry:   res.config.SOARetry,
		Expire:  res.config.SOAExpire,
		Minttl:  ttl,
	}
}

func (res *Resolver) formatNS(dom string) *dns.NS {
	ttl := uint32(res.config.TTL)

	return &dns.NS{
		Hdr: dns.RR_Header{
			Name:   dom,
			Rrtype: dns.TypeNS,
			Class:  dns.ClassINET,
			Ttl:    ttl,
		},
		Ns: res.config.SOAMname,
	}
}

func rcode(err error) int {
	switch err.(type) {
	case *ForwardError:
		return dns.RcodeRefused
	default:
		return dns.RcodeServerFailure
	}
}

func truncate(m *dns.Msg, udp bool) *dns.Msg {
	max := dns.MinMsgSize
	if !udp {
		max = dns.MaxMsgSize
	} else if opt := m.IsEdns0(); opt != nil {
		max = int(opt.UDPSize())
	}

	furtherTruncation := m.Len() > max
	m.Truncated = m.Truncated || furtherTruncation

	if !furtherTruncation {
		return m
	}

	m.Extra = nil // Drop all extra records first
	if m.Len() < max {
		return m
	}
	answers := m.Answer[:]
	left, right := 0, len(m.Answer)
	for {
		if left == right {
			break
		}
		mid := (left + right) / 2
		m.Answer = answers[:mid]
		if m.Len() < max {
			left = mid + 1
			continue
		}
		right = mid
	}
	return m
}

func isUDP(w dns.ResponseWriter) bool {
	return strings.HasPrefix(w.RemoteAddr().Network(), "udp")
}
