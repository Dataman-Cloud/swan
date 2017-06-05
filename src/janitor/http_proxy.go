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
	suffix    string
}

func NewHTTPProxy(cfg *config.Janitor, ups *Upstreams) http.Handler {
	return &httpProxy{
		config:    cfg,
		upstreams: ups,
		suffix:    "." + RESERVED_API_GATEWAY_DOMAIN + "." + cfg.Domain,
	}
}

func (p *httpProxy) FailByGateway(code int, reason string) {
	log.Warnln(code, reason)
}

func (p *httpProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// requestDurationBegin := time.Now()

	log.Debugf("proxy http request [%s] - [%s]", r.Method, r.Host)

	if len(r.Host) == 0 {
		p.FailByGateway(502, "header host empty")
		return
	}

	host := strings.Split(r.Host, ":")[0]
	if !strings.HasSuffix(host, p.suffix) {
		p.FailByGateway(400, fmt.Sprintf("request Host [%s] should end with %s", host, p.suffix))
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
		selected = p.upstreams.nextTarget(appID)

	case 5: // task
		appID := fmt.Sprintf("%s-%s-%s-%s", slices[1], slices[2], slices[3], slices[4])
		taskID := fmt.Sprintf("%s-%s", slices[0], appID)
		selected = p.upstreams.getTarget(appID, taskID)

	default:
		p.FailByGateway(400, fmt.Sprintf("request Host [%s] invalid", host))
		return
	}

	if selected == nil {
		p.FailByGateway(404, fmt.Sprintf("not found any matched targets for request Host [%s]", host))
		return
	}

	if err := p.AddHeaders(r, selected); err != nil {
		p.FailByGateway(500, err.Error())
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

	h.ServeHTTP(w, r)
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
