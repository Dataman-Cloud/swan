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

	log.Debugf("proxy redirecting request [%s-%s-%s] -> [%s-%s]",
		remoteIP, r.Method, r.Host,
		selected.TaskID, url,
	)

	return selected, nil
}

func (p *httpProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		err      error
		code     int
		in       int64     // received bytes
		out      int64     // transmitted bytes
		hijacked bool      // if hijacked or not
		dGlb     *deltaGlb // delta global
	)

	defer func() {
		if err != nil {
			if !hijacked { // we've proceed the error if hijacked
				http.Error(w, err.Error(), code)
			}
			log.Errorf("proxy serve error: %d - %v, recived:%d, transmitted:%d", code, err, in, out)
			dGlb = &deltaGlb{uint64(in), uint64(out), 1, 1}
		} else {
			log.Debugf("proxy serve succeed: recived:%d, transmitted:%d", in, out)
			dGlb = &deltaGlb{uint64(in), uint64(out), 1, 0}
		}
		p.stats.incr(nil, dGlb)
	}()

	selected, err := p.lookup(r)
	if err != nil {
		code = 404
		return
	}

	var (
		url, _ = selected.url()
		aid    = selected.AppID
		tid    = selected.TaskID
	)

	p.stats.incr(&deltaApp{aid, tid, 1, 0, 0, 1}, nil) // conn, active
	in, out, err, hijacked = p.doRawProxy(w, r, url)
	if err != nil {
		code = 500
	}
	p.stats.incr(&deltaApp{aid, tid, -1, uint64(in), uint64(out), 0}, nil) // disconnect
}

func (p *httpProxy) doRawProxy(w http.ResponseWriter, r *http.Request, t *url.URL) (int64, int64, error, bool) {
	var (
		in, out  int64
		hijacked bool
	)

	hj, ok := w.(http.Hijacker)
	if !ok {
		return in, out, fmt.Errorf("not a hijacker"), hijacked
	}

	src, _, err := hj.Hijack()
	if err != nil {
		return in, out, fmt.Errorf("hijack error: %v", err), hijacked
	}
	defer src.Close()

	hijacked = true

	dst, err := net.DialTimeout("tcp", t.Host, time.Second*60)
	if err != nil {
		err = fmt.Errorf("cannot connect to upstream %s: %v", t.Host, err)
		src.Write([]byte("HTTP/1.0 500 Internal Server Error\r\n\r\n" + err.Error() + "\n"))
		return in, out, err, hijacked
	}
	defer dst.Close()

	err = r.WriteProxy(dst) // send request to backend firstly
	if err != nil {
		err = fmt.Errorf("copying request to %s error: %v", t, err)
		src.Write([]byte("HTTP/1.0 500 Internal Server Error\r\n\r\n" + err.Error() + "\n"))
		return in, out, err, hijacked
	}
	in += httpRequestLen(r)

	errc := make(chan error, 2)
	cp := func(w io.WriteCloser, r io.Reader, c *int64) {
		defer w.Close()

		n, err := io.Copy(w, r)
		if n > 0 {
			*c += n
		}
		errc <- err
	}

	go cp(dst, src, &in)
	cp(src, dst, &out) // note: hanging wait while copying the response

	err = <-errc
	if err != nil && err != io.EOF {
		err = fmt.Errorf("io copy error: %v", err)
		src.Write([]byte("HTTP/1.0 500 Internal Server Error\r\n\r\n" + err.Error() + "\n"))
		return in, out, err, hijacked
	}
	return in, out, nil, hijacked
}

// try hard to obtain the size of initial raw HTTP request according by RFC7231.
// Note: we can't obtain the actually exact size through *http.Request because some details
// of the initial request are lost while parsing it into *http.Request within golang http.Server
// Note: do NOT try `httputil.DumpRequest` that will lead to poor performance.
func httpRequestLen(r *http.Request) int64 {
	n := int64(len(r.Method) + len(r.URL.Path) + len(r.Proto) + 3)
	n += int64(len(r.Host) + 3)

	for k, vs := range r.Header {
		n += int64(len(k) + 3)
		for _, v := range vs {
			n += int64(len(v))
		}
	}

	if len := r.ContentLength; len > 0 {
		n += len
	}
	return n
}
