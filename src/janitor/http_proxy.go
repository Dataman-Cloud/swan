package janitor

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/Dataman-Cloud/swan/src/config"
)

const (
	RESERVED_API_GATEWAY_DOMAIN = "gateway"
)

type httpProxy struct {
	config    *config.Janitor
	upstreams *Upstreams
	stats     *Stats
	suffix    string
}

func NewHTTPProxy(cfg *config.Janitor, ups *Upstreams, sta *Stats) http.Handler {
	return &httpProxy{
		config:    cfg,
		upstreams: ups,
		stats:     sta,
		suffix:    "." + RESERVED_API_GATEWAY_DOMAIN + "." + cfg.Domain,
	}
}

func (p *httpProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		err  error
		code int
		gGlb = &deltaGlb{0, 0, 1, 0}
	)

	defer func() {
		if err != nil {
			http.Error(w, err.Error(), code)
			log.Errorln("proxy serve error:", err)
			gGlb.fail = 1
		}
		p.stats.incr(nil, gGlb)
	}()

	if len(r.Host) == 0 {
		code, err = 400, errors.New("request Host empty")
		return
	}

	host := strings.Split(r.Host, ":")[0]
	if !strings.HasSuffix(host, p.suffix) {
		code, err = 400, fmt.Errorf("request Host [%s] should end with %s", host, p.suffix)
		return
	}

	remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		code, err = 400, fmt.Errorf("request RemoteAddr [%s] unrecognized", r.RemoteAddr)
		return
	}

	var (
		wildcardDomain = strings.TrimSuffix(host, p.suffix)
		slices         = strings.Split(wildcardDomain, ".")
		selected       *Target
	)

	switch len(slices) {

	case 4: // app
		appID := fmt.Sprintf("%s-%s-%s-%s", slices[0], slices[1], slices[2], slices[3])
		selected = p.upstreams.lookup(remoteIP, appID, "")

	case 5: // task
		appID := fmt.Sprintf("%s-%s-%s-%s", slices[1], slices[2], slices[3], slices[4])
		taskID := fmt.Sprintf("%s-%s", slices[0], appID)
		selected = p.upstreams.lookup(remoteIP, appID, taskID)

	default:
		code, err = 400, fmt.Errorf("request Host [%s] invalid", host)
		return
	}

	if selected == nil {
		code, err = 404, fmt.Errorf("not found any matched targets for request Host [%s]", host)
		return
	}

	log.Debugf("proxy redirect request [%s-%s-%s] -> [%s-%s]",
		remoteIP, r.Method, r.Host,
		selected.TaskID, selected.url(),
	)

	if err := p.AddHeaders(r, selected); err != nil {
		code, err = 500, fmt.Errorf("add header error: %v", err)
		return
	}

	var h http.Handler
	switch {
	case r.Header.Get("Upgrade") == "websocket":
		h = newRawProxy(selected.url())

	case r.Header.Get("Accept") == "text/event-stream":
		h = newHTTPProxy(selected.url(), p.config.FlushInterval)

	default:
		h = newHTTPProxy(selected.url(), time.Duration(0))
	}

	p.stats.incr(&deltaApp{selected.AppID, selected.TaskID, 1, 0, 0}, nil)
	h.ServeHTTP(w, r)
	p.stats.incr(&deltaApp{selected.AppID, selected.TaskID, -1, 0, 0}, nil)
}

func (proxy *httpProxy) AddHeaders(r *http.Request, t *Target) error {
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

func localPort(r *http.Request) string {
	if r == nil {
		return ""
	}
	n := strings.Index(r.Host, ":")
	if n > 0 && n < len(r.Host)-1 {
		return r.Host[n+1:]
	}
	if r.TLS != nil {
		return "443"
	}
	return "80"
}

func newHTTPProxy(t *url.URL, flush time.Duration) http.Handler {
	p := httputil.NewSingleHostReverseProxy(t)
	p.FlushInterval = flush
	return p
}
