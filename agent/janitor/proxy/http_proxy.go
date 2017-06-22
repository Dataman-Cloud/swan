package proxy

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/Dataman-Cloud/swan/agent/janitor/stats"
	"github.com/Dataman-Cloud/swan/agent/janitor/upstream"
)

const (
	APIGATEWAY = "gateway"
)

// generic http proxy handler
type HTTPProxy struct {
	suffix string
}

func NewHTTPProxyHandler(domain string) http.Handler {
	return &HTTPProxy{
		suffix: "." + APIGATEWAY + "." + domain,
	}
}

// lookup a proper backend according by request
func (p *HTTPProxy) lookup(r *http.Request) (*upstream.BackendCombined, error) {
	remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return nil, fmt.Errorf("request RemoteAddr [%s] unrecognized", r.RemoteAddr)
	}

	if len(r.Host) == 0 {
		return nil, errors.New("request Host empty")
	}

	var (
		host     = strings.Split(r.Host, ":")[0]
		byAlias  bool // flag on looking up by upstream alias or not
		selected *upstream.BackendCombined
	)
	if !strings.HasSuffix(host, p.suffix) {
		byAlias = true
	}

	if byAlias {
		selected = upstream.LookupAlias(remoteIP, host)

	} else {
		trimed := strings.TrimSuffix(host, p.suffix)
		ss := strings.Split(trimed, ".")

		switch len(ss) {
		case 4: // upstream
			ups := fmt.Sprintf("%s-%s-%s-%s", ss[0], ss[1], ss[2], ss[3])
			selected = upstream.Lookup(remoteIP, ups, "")
		case 5: // specified backend
			ups := fmt.Sprintf("%s-%s-%s-%s", ss[1], ss[2], ss[3], ss[4])
			backend := fmt.Sprintf("%s-%s", ss[0], ups)
			selected = upstream.Lookup(remoteIP, ups, backend)
		default:
			return nil, fmt.Errorf("request Host [%s] invalid", host)
		}
	}

	if selected == nil {
		return nil, fmt.Errorf("no matched backends for request [%s]", host)
	}

	log.Debugf("[HTTP] proxy redirecting request [%s] -> [%s-%s] -> [%s-%s]",
		remoteIP, r.Method, r.Host, selected.Backend.ID, selected.Addr(),
	)

	return selected, nil
}

// implements http.Handler interface
func (p *HTTPProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		err  error
		in   int64           // received bytes
		out  int64           // transmitted bytes
		dGlb *stats.DeltaGlb // delta global
	)

	defer func() {
		if err != nil {
			log.Errorf("[HTTP] proxy serve error: %v, received:%d, transmitted:%d", err, in, out)
			dGlb = &stats.DeltaGlb{uint64(in), uint64(out), 1, 1}
		} else {
			log.Printf("[HTTP] proxy serve succeed: received:%d, transmitted:%d", in, out)
			dGlb = &stats.DeltaGlb{uint64(in), uint64(out), 1, 0}
		}
		stats.Incr(nil, dGlb)
	}()

	// lookup a proper backend according by request
	selected, err := p.lookup(r)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}

	// detect & update backend scheme
	if selected.Backend.Scheme == "" {
		https, err := detectHTTPs(selected.Addr())
		if err != nil {
			err = fmt.Errorf("detect selected scheme error: %v", err)
			http.Error(w, err.Error(), 500)
			return
		}

		if https {
			selected.Backend.Scheme = "https"
		} else {
			selected.Backend.Scheme = "http"
		}

		upstream.UpsertBackend(selected)
	}

	var (
		addr    = selected.Backend.Addr()
		sche    = selected.Backend.Scheme
		ups     = selected.Upstream.Name
		backend = selected.Backend.ID
	)

	// obtian the underlying net.Conn
	hj, ok := w.(http.Hijacker)
	if !ok {
		err = fmt.Errorf("not support http hijack: %T", w)
		http.Error(w, err.Error(), 500)
		return
	}

	conn, _, err := hj.Hijack()
	if err != nil {
		err = fmt.Errorf("hijack tcp conn error: %v", err)
		http.Error(w, err.Error(), 500)
		return
	}
	defer conn.Close()

	// do proxy
	stats.Incr(&stats.DeltaBackend{ups, backend, 1, 0, 0, 1}, nil) // conn, active
	in, out, err = p.doRawProxy(conn, r, sche, addr)
	stats.Incr(&stats.DeltaBackend{ups, backend, -1, uint64(in), uint64(out), 0}, nil) // disconnect
}

func (p *HTTPProxy) doRawProxy(src net.Conn, req *http.Request, sche, addr string) (int64, int64, error) {
	var in, out int64

	// dial backend
	dst, err := net.DialTimeout("tcp", addr, time.Second*60)
	if err != nil {
		err = fmt.Errorf("cannot connect to upstream %s: %v", addr, err)
		src.Write([]byte("HTTP/1.0 500 Internal Server Error\r\n\r\n" + err.Error() + "\r\n"))
		return in, out, err
	}
	defer dst.Close()

	// tls wrap and try handshake
	if sche == "https" {
		dst, err = wrapWithTLS(dst)
		if err != nil {
			err = fmt.Errorf("tls handshake with upstream %s error: %v", addr, err)
			src.Write([]byte("HTTP/1.0 500 Internal Server Error\r\n\r\n" + err.Error() + "\r\n"))
			return in, out, err
		}
	}

	err = req.WriteProxy(dst) // send original request
	if err != nil {
		err = fmt.Errorf("copying request to %s error: %v", addr, err)
		src.Write([]byte("HTTP/1.0 500 Internal Server Error\r\n\r\n" + err.Error() + "\r\n"))
		return in, out, err
	}
	in += httpRequestLen(req)

	// io copy between src & dst
	errc := make(chan error, 2)
	cp := func(w io.WriteCloser, r io.Reader, c *int64) {
		defer w.Close()

		n, err := io.Copy(w, r) // TODO caculate each piece of io buffer by real time
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
		src.Write([]byte("HTTP/1.0 500 Internal Server Error\r\n\r\n" + err.Error() + "\r\n"))
		return in, out, err
	}
	return in, out, nil
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

func detectHTTPs(addr string) (https bool, err error) {
	conn, err := net.DialTimeout("tcp", addr, time.Second*60)
	if err != nil {
		return
	}
	defer conn.Close()

	_, err = conn.Write([]byte("GET / HTTP/1.1\r\n\r\n"))
	if err != nil {
		return
	}

	b := make([]byte, 5)
	_, err = conn.Read(b)
	if err != nil {
		return
	}

	https = string(b[:]) != "HTTP/" // or use: b[0] == 21
	return
}

// wrap a plain net.Conn with tls and try tls handshake
func wrapWithTLS(plainConn net.Conn) (net.Conn, error) {
	tlsConn := tls.Client(plainConn, &tls.Config{InsecureSkipVerify: true})

	errCh := make(chan error, 2)
	timer := time.AfterFunc(time.Second*10, func() {
		errCh <- errors.New("timeout on tls handshake")
	})
	defer timer.Stop()

	go func() {
		errCh <- tlsConn.Handshake()
	}()

	if err := <-errCh; err != nil {
		return nil, err
	}
	return tlsConn, nil
}
