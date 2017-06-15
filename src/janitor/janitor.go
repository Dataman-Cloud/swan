package janitor

import (
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/Dataman-Cloud/swan/src/config"
)

func init() {
	rand.Seed(time.Now().UnixNano())

	// disable HTTP/2 server side support, because when `Chrome/Firefox` visit `https://`,
	// http.ResponseWriter is actually implemented by *http.http2responseWriter which
	// does NOT implemented http.Hijacker
	// See: https://github.com/golang/go/issues/14797
	os.Setenv("GODEBUG", "http2server=0")
}

type JanitorServer struct {
	config       *config.Janitor
	upstreams    *Upstreams
	eventChan    chan *TargetChangeEvent
	stats        *Stats
	httpd        *http.Server
	httpdTLS     *http.Server
	tcpd         map[string]*tcpProxyServer // listen -> tcp proxy server
	sync.RWMutex                            // protect tcpd
}

func NewJanitorServer(cfg *config.Janitor) *JanitorServer {
	s := &JanitorServer{
		config:    cfg,
		upstreams: &Upstreams{Upstreams: make([]*Upstream, 0, 0)},
		eventChan: make(chan *TargetChangeEvent, 1024),
		stats:     newStats(),
		tcpd:      make(map[string]*tcpProxyServer),
	}

	s.httpd = &http.Server{
		Addr:    s.config.ListenAddr,
		Handler: s.newHTTPProxyHandler(),
	}

	if s.config.TLSListenAddr != "" {
		s.httpdTLS = &http.Server{
			Addr:    s.config.TLSListenAddr,
			Handler: s.newHTTPProxyHandler(),
		}
	}

	return s
}

func (s *JanitorServer) EmitChange(ev *TargetChangeEvent) {
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

		target := ev.Target.format()

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

func (s *JanitorServer) upsertBackend(target *Target) error {
	if err := target.valid(); err != nil {
		return err
	}

	first, err := s.upstreams.upsertTarget(target)
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

	tcpProxy := s.newTCPProxyServer(l)
	if err := tcpProxy.listen(); err != nil {
		return err
	}

	go tcpProxy.serve()

	s.Lock()
	s.tcpd[l] = tcpProxy
	s.Unlock()

	return nil
}

func (s *JanitorServer) removeBackend(target *Target) {
	onLast := s.upstreams.removeTarget(target)
	s.stats.del(target.AppID, target.TaskID)

	if !onLast {
		return
	}

	l := target.AppListen
	if l == "" {
		return
	}

	s.Lock()
	if tcpProxy, ok := s.tcpd[l]; ok {
		tcpProxy.stop()
	}
	delete(s.tcpd, l)
	s.Unlock()
}
