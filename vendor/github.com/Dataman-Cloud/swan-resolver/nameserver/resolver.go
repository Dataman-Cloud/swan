package nameserver

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/miekg/dns"
	"golang.org/x/net/context"
)

const (
	RESERVED_API_GATEWAY_DOMAIN = "gateway"
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
	rr.As = make(map[string][]string)
	rr.SRVs = make(map[string][]string)
	rr.ProxiesAs = make(map[string][]string)

	res.rg = &rr
	res.defaultFwd = NewForwarder(config.Resolvers, exchangers(config.ExchangeTimeout, "udp"))

	go func() {
		res.rg.WatchEvent(context.Background())
	}()

	return res
}

func (res *Resolver) RecordGeneratorChangeChan() chan *RecordGeneratorChangeEvent {
	return res.rg.RecordGeneratorChangeChan
}

func (res *Resolver) Start(ctx context.Context) error {
	return <-res.Run(ctx)
}

func (res *Resolver) Run(ctx context.Context) <-chan error {
	dns.HandleFunc(res.rg.Domain, res.HandleSwan)
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

	rg         *RecordGenerator
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
	case dns.TypeANY:
		errs.Add(
			res.handleSRV(rs, name, m, r),
			res.handleA(rs, name, m),
		)
	}

	if len(m.Answer) == 0 {
		errs.Add(errors.New("no record found"))
	}

	if !errs.Nil() {
		logrus.Errorf(errs.Error())
	}

	reply(w, m)
}

func (res *Resolver) records() *RecordGenerator {
	return res.rg
}

func (res *Resolver) handleSRV(rs *RecordGenerator, name string, m, r *dns.Msg) error {
	var errs multiError
	added := map[string]struct{}{} // track the A RR's we've already added, avoid dups
	srvs, ok := rs.SRVs[name]
	if !ok {
		errs.Add(errors.New("srvs not found"))
	} else {
		for _, srv := range srvs {
			//get srv resource record
			srvRR, err := res.formatSRV(r.Question[0].Name, srv)
			if err != nil {
				errs.Add(err)
				continue
			}

			m.Answer = append(m.Answer, srvRR)

			//get ip
			name := strings.Split(srv, ":")[0]
			if _, found := added[name]; found {
				continue
			}

			hosts, ok := rs.As[name]
			if !ok {
				errs.Add(errors.New(fmt.Sprintf("%s is not found in rrs", name)))
				continue
			}

			if len(hosts) == 0 {
				continue
			}
			for _, host := range hosts {
				aRR, err := res.formatA(name, host)
				if err != nil {
					errs.Add(err)
					continue
				}
				m.Extra = append(m.Extra, aRR)
				added[name] = struct{}{}
			}
		}
	}

	return errs
}

func (res *Resolver) handleA(rs *RecordGenerator, name string, m *dns.Msg) error {
	var errs multiError
	var records = make([]string, 0)
	var isDigit = regexp.MustCompile("\\d+")
	tokens := strings.Split(strings.TrimRight(name, res.rg.Domain), ".")

	if tokens[len(tokens)-1] == RESERVED_API_GATEWAY_DOMAIN { // api gateway resolve with higher priority
		for k, hosts := range rs.ProxiesAs {
			ok := strings.HasSuffix(k, res.rg.Domain+".")
			if ok {
				records = append(records, hosts...)
			}
		}
	} else {
		if len(tokens) == 4 && isDigit.MatchString(tokens[0]) {
			for k, hosts := range rs.As {
				if name == k {
					records = append(records, hosts...)
				}
			}
		} else if len(tokens) == 3 { // nginx.xcm.foobar  -  .swan.com
			for k, hosts := range rs.As {
				if sliceEqual(strings.Split(strings.TrimRight(k, res.rg.Domain), ".")[1:], tokens) {
					records = append(records, hosts...)
				}
			}
		}
	}

	recordsAdded := make([]string, 0)
	for _, host := range records {
		if stringInSlice(host, recordsAdded) { // make sure no duplicated record added to A
			continue
		}
		recordsAdded = append(recordsAdded, host)

		rr, err := res.formatA(name, host)
		if err != nil {
			errs.Add(err)
		} else {
			m.Answer = append(m.Answer, rr)
		}
	}

	return errs
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

func rcode(err error) int {
	switch err.(type) {
	case *ForwardError:
		return dns.RcodeRefused
	default:
		return dns.RcodeServerFailure
	}
}

func isUDP(w dns.ResponseWriter) bool {
	return strings.HasPrefix(w.RemoteAddr().Network(), "udp")
}

func sliceEqual(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i, _ := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
