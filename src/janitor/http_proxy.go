package janitor

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
)

const (
	RESERVED_API_GATEWAY_DOMAIN = "gateway"
)

type httpProxy struct {
	domain    string
	upstreams *Upstreams
	stats     *Stats
	suffix    string
}

func NewHTTPProxy(domain string, ups *Upstreams, sta *Stats) http.Handler {
	return &httpProxy{
		upstreams: ups,
		stats:     sta,
		suffix:    "." + RESERVED_API_GATEWAY_DOMAIN + "." + domain,
	}
}

func (p *httpProxy) lookup(r *http.Request) (*Target, error) {
	remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return nil, fmt.Errorf("request RemoteAddr [%s] unrecognized", r.RemoteAddr)
	}

	if len(r.Host) == 0 {
		return nil, errors.New("request Host empty")
	}

	var (
		host     = strings.Split(r.Host, ":")[0]
		byAlias  bool // flag on looking up by target alias or not
		selected *Target
	)
	if !strings.HasSuffix(host, p.suffix) {
		byAlias = true
	}

	if byAlias {
		selected = p.upstreams.lookupAlias(remoteIP, host)

	} else {
		trimed := strings.TrimSuffix(host, p.suffix)
		ss := strings.Split(trimed, ".")

		switch len(ss) {
		case 4: // app
			appID := fmt.Sprintf("%s-%s-%s-%s", ss[0], ss[1], ss[2], ss[3])
			selected = p.upstreams.lookup(remoteIP, appID, "")
		case 5: // task
			appID := fmt.Sprintf("%s-%s-%s-%s", ss[1], ss[2], ss[3], ss[4])
			taskID := fmt.Sprintf("%s-%s", ss[0], appID)
			selected = p.upstreams.lookup(remoteIP, appID, taskID)
		default:
			return nil, fmt.Errorf("request Host [%s] invalid", host)
		}
	}

	if selected == nil {
		return nil, fmt.Errorf("not found any matched targets for request Host [%s]", host)
	}

	url, err := selected.url()
	if err != nil {
		return nil, err
	}

	log.Debugf("proxy redirect request [%s-%s-%s] -> [%s-%s]",
		remoteIP, r.Method, r.Host,
		selected.TaskID, url,
	)

	return selected, nil
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
			log.Errorln(err)
			gGlb.fail = 1
		}
		p.stats.incr(nil, gGlb)
	}()

	selected, err := p.lookup(r)
	if err != nil {
		code = 404
		return
	}

	url, _ := selected.url()

	p.stats.incr(&deltaApp{selected.AppID, selected.TaskID, 1, 0, 0}, nil)
	err = p.rawHTTPProxy(w, r, url)
	if err != nil {
		code = 500
	}
	p.stats.incr(&deltaApp{selected.AppID, selected.TaskID, -1, 0, 0}, nil)
}

func (proxy *httpProxy) rawHTTPProxy(w http.ResponseWriter, r *http.Request, t *url.URL) error {
	hj, ok := w.(http.Hijacker)
	if !ok {
		return fmt.Errorf("not a hijacker")
	}

	in, _, err := hj.Hijack()
	if err != nil {
		return fmt.Errorf("hijack error: %v", err)
	}
	defer in.Close()

	out, err := net.DialTimeout("tcp", t.Host, time.Second*60)
	if err != nil {
		return fmt.Errorf("cannot connect to upstream %s", t.Host)
	}
	defer out.Close()

	err = r.Write(out) // send request to backend firstly
	if err != nil {
		return fmt.Errorf("copying request to %s error: %v", t, err)
	}

	errc := make(chan error, 2)
	cp := func(dst io.Writer, src io.Reader) {
		_, err := io.Copy(dst, src)
		errc <- err
	}

	go cp(out, in)
	go cp(in, out)
	err = <-errc
	if err != nil && err != io.EOF {
		return fmt.Errorf("io copy error: %v", err)
	}
	return nil
}
