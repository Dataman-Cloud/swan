package janitor

import (
	"net/http"
	"os"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"

	"github.com/Dataman-Cloud/swan/src/config"
	"github.com/Dataman-Cloud/swan/src/janitor/proxy"
	"github.com/Dataman-Cloud/swan/src/janitor/stats"
	"github.com/Dataman-Cloud/swan/src/janitor/upstream"
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
	eventChan    chan *upstream.BackendEvent
	httpd        *http.Server
	httpdTLS     *http.Server
	tcpd         map[string]*proxy.TCPProxyServer // listen -> tcp proxy server
	sync.RWMutex                                  // protect tcpd
}

func NewJanitorServer(cfg *config.Janitor) *JanitorServer {
	s := &JanitorServer{
		config:    cfg,
		eventChan: make(chan *upstream.BackendEvent, 1024),
		tcpd:      make(map[string]*proxy.TCPProxyServer),
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

func (s *JanitorServer) EmitEvent(ev *upstream.BackendEvent) {
	s.eventChan <- ev
}

func (s *JanitorServer) Start() error {
	go s.watchEvent()

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

func (s *JanitorServer) watchEvent() {
	log.Println("proxy listening on backends event ...")

	for ev := range s.eventChan {
		log.Printf("proxy caught event: %s", ev)

		ev.Format()

		switch strings.ToLower(ev.Action) {
		case "add", "change":
			if err := s.upsertBackend(ev.BackendCombined); err != nil {
				log.Errorln("upsert backend error:", err)
			}

		case "del":
			s.removeBackend(ev.BackendCombined)

		default:
			log.Warnln("unrecognized event action", ev.Action)
		}
	}

	panic("event channel closed, never be here")
}

func (s *JanitorServer) upsertBackend(cmb *upstream.BackendCombined) error {
	if err := cmb.Valid(); err != nil {
		return err
	}

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
