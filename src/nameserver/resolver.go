package nameserver

import (
	"errors"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/miekg/dns"

	"github.com/Dataman-Cloud/swan/src/config"
)

const (
	RESERVED_API_GATEWAY_DOMAIN = "gateway"
)

type Resolver struct {
	recordChangeChan chan *RecordChangeEvent

	recordHolder *RecordHolder
	config       *config.DNS
	defaultFwd   Forwarder
}

func NewResolver(config *config.DNS) *Resolver {
	resolver := &Resolver{
		recordChangeChan: make(chan *RecordChangeEvent, 1),

		config:       config,
		defaultFwd:   NewForwarder(config.Resolvers, exchangers(config.ExchangeTimeout, "udp")),
		recordHolder: NewRecordHolder(config.Domain),
	}

	go resolver.watchEvent()

	return resolver
}

func (resolver *Resolver) EmitChange(ev *RecordChangeEvent) {
	resolver.recordChangeChan <- ev
}

func (resolver *Resolver) watchEvent() {
	for e := range resolver.recordChangeChan {
		switch e.Change {
		case "del":
			resolver.recordHolder.Del(e.record())
		case "add":
			resolver.recordHolder.Add(e.record())
		}
	}
}

func (res *Resolver) Start() error {
	dns.HandleFunc(res.config.Domain, res.HandleSwan)
	dns.HandleFunc(".", res.HandleNonSwan(res.defaultFwd))

	server := &dns.Server{
		Addr:       res.config.ListenAddr,
		Net:        "udp",
		TsigSecret: nil,
	}

	return server.ListenAndServe()
}

func (res *Resolver) HandleSwan(w dns.ResponseWriter, req *dns.Msg) {
	msg := &dns.Msg{MsgHdr: dns.MsgHdr{
		Authoritative:      true,
		RecursionAvailable: res.config.RecurseOn,
	}}
	msg.SetReply(req)

	var errs multiError
	name := strings.ToLower(req.Question[0].Name)

	logrus.Debugf("resolve dns hostname %s", name)

	switch req.Question[0].Qtype {
	case dns.TypeSRV:
		errs.Add(res.handleSRV(name, msg, req))
	case dns.TypeA:
		errs.Add(res.handleA(name, msg))
	case dns.TypeANY:
		errs.Add(
			res.handleSRV(name, msg, req),
			res.handleA(name, msg),
		)
	}

	if len(msg.Answer) == 0 {
		errs.Add(errors.New("no record found"))
	}

	if !errs.Nil() {
		logrus.Errorln(errs.Error())
	}

	reply(w, msg)
}

func (res *Resolver) handleSRV(name string, m, r *dns.Msg) error {
	var errs multiError
	for _, record := range res.recordHolder.GetSRV(name) {
		rr, err := res.buildSRV(name, record)
		if err != nil {
			errs.Add(err)
		} else {
			m.Answer = append(m.Answer, rr)
		}
	}

	return errs
}

func (res *Resolver) handleA(name string, m *dns.Msg) error {
	var errs multiError
	for _, record := range res.recordHolder.GetA(name) {
		rr, err := res.buildA(name, record)
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
			logrus.Debugf("no answer found")
		}
		reply(w, m)
	}
}

// reply writes the given dns.Msg out to the given dns.ResponseWriter,
// compressing the message first and truncating it accordingly.
func reply(w dns.ResponseWriter, m *dns.Msg) {
	if err := w.WriteMsg(m); err != nil {
		logrus.Errorln(err)
	}
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

func (res *Resolver) buildSRV(name string, record *Record) (*dns.SRV, error) {
	ttl := uint32(res.config.TTL)

	p, err := strconv.Atoi(record.Port)
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
		Target:   record.WithSlotDomain() + "." + res.config.Domain + ".",
	}, nil
}

func (res *Resolver) buildA(name string, record *Record) (*dns.A, error) {
	ttl := uint32(res.config.TTL)

	a := net.ParseIP(record.Ip)
	if a == nil {
		return nil, errors.New("invalid target")
	}

	return &dns.A{
		Hdr: dns.RR_Header{
			Name:   name,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    ttl},
		A: a.To4(),
	}, nil
}

func rcode(err error) int {
	switch err.(type) {
	case *ForwardError:
		return dns.RcodeRefused
	default:
		return dns.RcodeServerFailure
	}
}
