package janitor

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	RESERVED_API_GATEWAY_DOMAIN = "gateway"
)

func newHTTPProxy(t *url.URL,
	tr http.RoundTripper,
	flush time.Duration,
	P *Prometheus,
	taskId string,
) http.Handler {
	rp := httputil.NewSingleHostReverseProxy(t)
	rp.FlushInterval = flush
	rp.Transport = &meteredRoundTripper{tr, P, taskId}
	return rp
}

type meteredRoundTripper struct {
	tr     http.RoundTripper
	P      *Prometheus
	taskId string
}

func (m *meteredRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	backendBegin := time.Now()
	resp, err := m.tr.RoundTrip(r)
	if err != nil {
		return resp, err
	}

	if r.Header.Get("X-Forwarded-Proto") == "http" {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		err = resp.Body.Close()
		if err != nil {
			return nil, err
		}
		body := ioutil.NopCloser(bytes.NewReader(b))
		resp.Body = body

		resp.ContentLength = int64(len(b))
		resp.Header.Set("Content-Length", strconv.Itoa(len(b)))

		m.P.ResponseSize.Observe(float64(resp.ContentLength))
	}

	m.P.BackendDuration.Observe(time.Now().Sub(backendBegin).Seconds())

	m.P.RequestCounter.With(prometheus.Labels{
		"source": "UserApp",
		"code":   fmt.Sprintf("%d", resp.StatusCode),
		"method": r.Method,
		"path":   r.URL.RawPath,
		"taskId": m.taskId,
		"reason": "nomral",
	}).Inc()
	return resp, err
}

// httpProxy is a dynamic reverse proxy for HTTP and HTTPS protocols.
type layer7Proxy struct {
	tr             http.RoundTripper
	config         Config
	UpstreamLoader *UpstreamLoader
	P              *Prometheus
}

func NewLayer7Proxy(tr http.RoundTripper,
	Config Config,
	UpstreamLoader *UpstreamLoader,
	P *Prometheus) http.Handler {

	return &layer7Proxy{
		tr:             tr,
		config:         Config,
		UpstreamLoader: UpstreamLoader,
		P:              P,
	}
}

func (p *layer7Proxy) FailByGateway(w http.ResponseWriter, r *http.Request, httpCode int, reason string) {
	log.Debugf(reason)
	p.P.RequestCounter.With(prometheus.Labels{
		"source": "GATEWAY",
		"code":   fmt.Sprintf("%d", httpCode),
		"method": r.Method,
		"path":   r.URL.RawPath,
		"reason": reason,
		"taskId": "",
	}).Inc()
}

func (p *layer7Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestDurationBegin := time.Now()
	log.Debugf("got request for hostname: %s", r.Host)

	var selectedTarget *Target
	if len(r.Host) == 0 {
		p.FailByGateway(w, r, http.StatusBadGateway, "header host blank")
		return
	}

	host := strings.Split(r.Host, ":")[0]

	if !strings.HasSuffix(host, p.config.Domain) {
		p.FailByGateway(w, r, http.StatusBadRequest, fmt.Sprintf("header host doesn't end with %s, abort", p.config.Domain))
		return
	}

	domainIndex := strings.Index(host, RESERVED_API_GATEWAY_DOMAIN+"."+p.config.Domain)
	if domainIndex <= 0 {
		p.FailByGateway(w, r, http.StatusBadRequest, fmt.Sprintf("header host is %s doesn't match [0\\.]app.user.cluster.domain.com abort", host))
		return
	}

	wildcardDomain := host[0 : domainIndex-1]
	slices := strings.SplitN(wildcardDomain, ".", 2)
	if len(slices) != 2 {
		p.FailByGateway(w, r, http.StatusBadRequest, fmt.Sprintf("header host is %s doesn't match [0\\.]app.user.cluster.domain.com abort", host))
		return
	}

	digitRegexp := regexp.MustCompile("[0-9]+")
	if digitRegexp.MatchString(slices[0]) {
		upstream := p.UpstreamLoader.Get(slices[1])
		if upstream == nil {
			p.FailByGateway(w, r, http.StatusNotFound, fmt.Sprintf("fail to found any upstream for %s", host))
			return
		}

		target := upstream.GetTarget(wildcardDomain)
		if target == nil {
			p.FailByGateway(w, r, http.StatusNotFound, fmt.Sprintf("fail to found any target for %s", host))
			return
		}

		selectedTarget = target
	} else {
		upstream := p.UpstreamLoader.Get(wildcardDomain)
		if upstream == nil {
			p.FailByGateway(w, r, http.StatusNotFound, fmt.Sprintf("fail to found any upstream for %s", host))
			return
		}

		selectedTarget = upstream.NextTargetEntry()
	}

	if selectedTarget == nil {
		p.FailByGateway(w, r, http.StatusBadGateway, fmt.Sprintf("selectedTarget [%s] was found", selectedTarget.Entry()))
		return
	}

	if err := p.AddHeaders(r, selectedTarget); err != nil {
		p.FailByGateway(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	var h http.Handler
	switch {
	case r.Header.Get("Upgrade") == "websocket":
		h = newRawProxy(selectedTarget.Entry())

	case r.Header.Get("Accept") == "text/event-stream":
		h = newHTTPProxy(selectedTarget.Entry(), p.tr, p.config.FlushInterval, p.P, selectedTarget.TaskID)

	default:
		h = newHTTPProxy(selectedTarget.Entry(), p.tr, time.Duration(0), p.P, selectedTarget.TaskID)
	}

	h.ServeHTTP(w, r)
	p.P.RequestDuration.Observe(time.Now().Sub(requestDurationBegin).Seconds())
}

func (proxy *layer7Proxy) AddHeaders(r *http.Request, t *Target) error {
	remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return errors.New("cannot parse " + r.RemoteAddr)
	}

	// set configurable ClientIPHeader
	// X-Real-Ip is set later and X-Forwarded-For is set
	// by the Go HTTP reverse proxy.
	//if proxy.handlerCfg.ClientIPHeader != "" &&
	//proxy.handlerCfg.ClientIPHeader != "X-Forwarded-For" &&
	//proxy.handlerCfg.ClientIPHeader != "X-Real-Ip" {
	//r.Header.Set(proxy.handlerCfg.ClientIPHeader, remoteIP)
	//}

	if r.Header.Get("X-Swan-Gateway-Addr") == "" {
		r.Header.Set("X-Swan-Gateway-Addr", proxy.config.ListenAddr)
	}

	if r.Header.Get("X-Swan-AppID") == "" {
		r.Header.Set("X-Swan-AppID", t.AppID)
	}

	if r.Header.Get("X-Swan-TaskID") == "" {
		r.Header.Set("X-Swan-TaskID", t.TaskID)
	}

	if r.Header.Get("X-Swan-TaskIP") == "" {
		r.Header.Set("X-Swan-TaskIP", t.TaskIP)
	}

	if r.Header.Get("X-Swan-TaskPort") == "" {
		r.Header.Set("X-Swan-TaskPort", fmt.Sprintf("%d", t.TaskPort))
	}

	if r.Header.Get("X-Swan-PortName") == "" {
		r.Header.Set("X-Swan-PortName", t.PortName)
	}

	if r.Header.Get("X-Swan-Weight") == "" {
		r.Header.Set("X-Swan-Weight", fmt.Sprintf("%f", t.Weight))
	}

	if r.Header.Get("X-Real-Ip") == "" {
		r.Header.Set("X-Real-Ip", remoteIP)
	}

	// set the X-Forwarded-For header for websocket
	// connections since they aren't handled by the
	// http proxy which sets it.
	ws := r.Header.Get("Upgrade") == "websocket"
	if ws {
		r.Header.Set("X-Forwarded-For", remoteIP)
	}

	if r.Header.Get("X-Forwarded-Proto") == "" {
		switch {
		case ws && r.TLS != nil:
			r.Header.Set("X-Forwarded-Proto", "wss")
		case ws && r.TLS == nil:
			r.Header.Set("X-Forwarded-Proto", "ws")
		case r.TLS != nil:
			r.Header.Set("X-Forwarded-Proto", "https")
		default:
			r.Header.Set("X-Forwarded-Proto", "http")
		}
	}

	if r.Header.Get("X-Forwarded-Port") == "" {
		r.Header.Set("X-Forwarded-Port", localPort(r))
	}

	fwd := r.Header.Get("Forwarded")
	if fwd == "" {
		fwd = "for=" + remoteIP
		switch {
		case ws && r.TLS != nil:
			fwd += "; proto=wss"
		case ws && r.TLS == nil:
			fwd += "; proto=ws"
		case r.TLS != nil:
			fwd += "; proto=https"
		default:
			fwd += "; proto=http"
		}
	}
	ip, _, err := net.SplitHostPort(proxy.config.ListenAddr)
	if err == nil && ip != "" {
		fwd += "; by=" + ip
	}
	r.Header.Set("Forwarded", fwd)

	//if cfg.TLSHeader != "" && r.TLS != nil {
	//r.Header.Set(cfg.TLSHeader, cfg.TLSHeaderValue)
	//}

	return nil
}
