package janitor

import (
	"net/http"
	"os"
	"sync"

	"github.com/Dataman-Cloud/swan/agent/janitor/proxy"
	"github.com/Dataman-Cloud/swan/agent/janitor/stats"
	"github.com/Dataman-Cloud/swan/agent/janitor/upstream"
	"github.com/Dataman-Cloud/swan/config"
)

func init() {
	// disable HTTP/2 server side support, because when `Chrome/Firefox` visit `https://`,
	// http.ResponseWriter is actually implemented by *http.http2responseWriter which
	// does NOT implemented http.Hijacker
	// See: https://github.com/golang/go/issues/14797
	os.Setenv("GODEBUG", "http2server=0")
}

type JanitorServer struct {
	config       *config.Janitor
	httpd        *http.Server
	httpdTLS     *http.Server
	tcpd         map[string]*proxy.TCPProxyServer // listen -> tcp proxy server
	sync.RWMutex                                  // protect tcpd
}

func NewJanitorServer(cfg *config.Janitor) *JanitorServer {
	s := &JanitorServer{
		config: cfg,
		tcpd:   make(map[string]*proxy.TCPProxyServer),
	}

	s.httpd = &http.Server{
		Addr:    s.config.ListenAddr,
		Handler: proxy.NewHTTPProxyHandler(cfg.Domain),
	}

	if s.config.TLSListenAddr != "" {
		s.httpdTLS = &http.Server{
			Addr:    s.config.TLSListenAddr,
			Handler: proxy.NewHTTPProxyHandler(cfg.Domain),
		}
	}

	return s
}

func (s *JanitorServer) Start() error {
	errCh := make(chan error, 2)

	go func() {
		defer s.httpd.Close()
		errCh <- s.httpd.ListenAndServe()
	}()

	go func() {
		if s.httpdTLS != nil {
			defer s.httpdTLS.Close()
			errCh <- s.httpdTLS.ListenAndServeTLS(s.config.TLSCertFile, s.config.TLSKeyFile)
		}
	}()

	return <-errCh
}

func (s *JanitorServer) UpsertBackend(cmb *upstream.BackendCombined) error {
	if err := cmb.Valid(); err != nil {
		return err
	}

	cmb.Format()

	first, err := upstream.UpsertBackend(cmb)
	if err != nil {
		return err
	}

	if !first {
		return nil
	}

	l := cmb.Upstream.Listen
	if l == "" {
		return nil
	}

	tcpProxy := proxy.NewTCPProxyServer(l)
	if err := tcpProxy.Listen(); err != nil {
		upstream.RemoveBackend(cmb) // roll back
		return err
	}

	go tcpProxy.Serve()

	s.Lock()
	s.tcpd[l] = tcpProxy
	s.Unlock()

	return nil
}

func (s *JanitorServer) removeBackend(cmb *upstream.BackendCombined) {
	u := upstream.GetUpstream(cmb.Upstream.Name)
	if u == nil {
		return
	}

	onLast := upstream.RemoveBackend(cmb)
	stats.Del(cmb.Upstream.Name, cmb.Backend.ID)

	if !onLast {
		return
	}

	l := u.Listen
	if l == "" {
		return
	}

	s.Lock()
	if tcpProxy, ok := s.tcpd[l]; ok {
		tcpProxy.Stop()
	}
	delete(s.tcpd, l)
	s.Unlock()
}
