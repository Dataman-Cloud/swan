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
	eventChan    chan *upstream.TargetChangeEvent
	httpd        *http.Server
	httpdTLS     *http.Server
	tcpd         map[string]*proxy.TCPProxyServer // listen -> tcp proxy server
	sync.RWMutex                                  // protect tcpd
}

func NewJanitorServer(cfg *config.Janitor) *JanitorServer {
	s := &JanitorServer{
		config:    cfg,
		eventChan: make(chan *upstream.TargetChangeEvent, 1024),
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

func (s *JanitorServer) EmitChange(ev *upstream.TargetChangeEvent) {
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
	log.Println("proxy listening on app event ...")

	for ev := range s.eventChan {
		log.Printf("proxy caught event: %s", ev)

		target := ev.Target.Format()

		switch strings.ToLower(ev.Change) {
		case "add", "change":
			if err := s.upsertBackend(target); err != nil {
				log.Errorln("upsert backend error:", err)
			}

		case "del":
			s.removeBackend(target)

		default:
			log.Warnln("unrecognized event change type", ev.Change)
		}
	}

	panic("event channel closed, never be here")
}

func (s *JanitorServer) upsertBackend(target *upstream.Target) error {
	if err := target.Valid(); err != nil {
		return err
	}

	first, err := upstream.UpsertTarget(target)
	if err != nil {
		return err
	}

	if !first {
		return nil
	}

	l := target.AppListen
	if l == "" {
		return nil
	}

	tcpProxy := proxy.NewTCPProxyServer(l)
	if err := tcpProxy.Listen(); err != nil {
		return err
	}

	go tcpProxy.Serve()

	s.Lock()
	s.tcpd[l] = tcpProxy
	s.Unlock()

	return nil
}

func (s *JanitorServer) removeBackend(target *upstream.Target) {
	onLast := upstream.RemoveTarget(target)
	stats.Del(target.AppID, target.TaskID)

	if !onLast {
		return
	}

	l := target.AppListen
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
