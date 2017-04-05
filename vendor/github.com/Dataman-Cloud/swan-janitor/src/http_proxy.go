package janitor

import (
	"errors"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
)

const (
	RESERVED_API_GATEWAY_DOMAIN = "gateway"
)

func newHTTPProxy(t *url.URL, tr http.RoundTripper, flush time.Duration) http.Handler {
	rp := httputil.NewSingleHostReverseProxy(t)
	rp.Transport = tr
	rp.FlushInterval = flush
	rp.Transport = &meteredRoundTripper{tr}
	return rp
}

type meteredRoundTripper struct {
	tr http.RoundTripper
}

func (m *meteredRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	resp, err := m.tr.RoundTrip(r)
	return resp, err
}

// httpProxy is a dynamic reverse proxy for HTTP and HTTPS protocols.
type httpProxy struct {
	tr             http.RoundTripper
	handlerCfg     HttpHandlerCfg
	listenAddr     string
	UpstreamLoader *UpstreamLoader
}

func NewHTTPProxy(tr http.RoundTripper, handlerCfg HttpHandlerCfg, listenAddr string, UpstreamLoader *UpstreamLoader) http.Handler {
	return &httpProxy{
		tr:             tr,
		listenAddr:     listenAddr,
		handlerCfg:     handlerCfg,
		UpstreamLoader: UpstreamLoader,
	}
}

func (p *httpProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Debugf("got request for hostname: %s", r.Host)

	var targetEntry *url.URL
	if len(r.Host) == 0 {
		log.Debugf("header HOST is null")
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	host := strings.Split(r.Host, ":")[0]
	log.Debugf("host [%s] is requested", host)

	if !strings.HasSuffix(host, p.handlerCfg.Domain) {
		log.Debugf("header host doesn't end with %s, abort", p.handlerCfg.Domain)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	domainIndex := strings.Index(host, RESERVED_API_GATEWAY_DOMAIN+"."+p.handlerCfg.Domain)
	if domainIndex == 0 {
		log.Debugf("header host is %s doesn't match [0\\.]app.user.cluster.domain.com abort", host)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	wildcardDomain := host[0 : domainIndex-1]
	slices := strings.Split(wildcardDomain, ".")
	if !(len(slices) == 3 || len(slices) == 4) {
		log.Debugf("slices is %s, header host is %s doesn't match [0\\.]app.user.cluster.domain.com abort", slices, host)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(slices) == 4 {
		taskID := strings.Join(slices, "-")
		appID := strings.Join(slices[1:], "-")
		upstream := p.UpstreamLoader.Get(appID)
		if upstream == nil {
			log.Debugf("fail to found any upstream for %s", host)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		target := upstream.GetTarget(taskID)
		if target == nil {
			log.Debugf("fail to found any target for %s", host)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		targetEntry = target.Entry()
	}

	if len(slices) == 3 {
		appID := strings.Join(slices, "-")
		upstream := p.UpstreamLoader.Get(appID)
		if upstream == nil {
			log.Debugf("fail to found any upstream for %s", host)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		targetEntry = upstream.NextTargetEntry()
	}

	log.Debugf("targetEntry [%s] is found", targetEntry)
	if targetEntry == nil {
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	if err := p.AddHeaders(r); err != nil {
		http.Error(w, "cannot parse "+r.RemoteAddr, http.StatusInternalServerError)
		return
	}

	var h http.Handler
	switch {
	case r.Header.Get("Upgrade") == "websocket":
		h = newRawProxy(targetEntry)

		// To use the filtered proxy use
		// h = newWSProxy(t.URL)

	case r.Header.Get("Accept") == "text/event-stream":
		// use the flush interval for SSE (server-sent events)
		// must be > 0s to be effective
		h = newHTTPProxy(targetEntry, p.tr, p.handlerCfg.FlushInterval)

	default:
		h = newHTTPProxy(targetEntry, p.tr, time.Duration(0))
	}

	//start := time.Now()
	h.ServeHTTP(w, r)
}

func (proxy *httpProxy) AddHeaders(r *http.Request) error {
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
	ip, _, err := net.SplitHostPort(proxy.listenAddr)
	if err == nil && ip != "" {
		fwd += "; by=" + ip
	}
	r.Header.Set("Forwarded", fwd)

	//if cfg.TLSHeader != "" && r.TLS != nil {
	//r.Header.Set(cfg.TLSHeader, cfg.TLSHeaderValue)
	//}

	return nil
}
