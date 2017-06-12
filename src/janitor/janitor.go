package janitor

import (
	"math/rand"
	"net/http"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/Dataman-Cloud/swan/src/config"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type JanitorServer struct {
	upstreams *Upstreams
	eventChan chan *TargetChangeEvent
	stats     *Stats
	httpd     *http.Server
	httpdTLS  *http.Server
	config    *config.Janitor
}

func NewJanitorServer(cfg *config.Janitor) *JanitorServer {
	s := &JanitorServer{
		config:    cfg,
		eventChan: make(chan *TargetChangeEvent, 1024),
		stats:     newStats(),
		upstreams: &Upstreams{
			Upstreams: make([]*Upstream, 0, 0),
		},
	}

	s.httpd = &http.Server{
		Addr:    s.config.ListenAddr,
		Handler: NewHTTPProxy(cfg.Domain, s.upstreams, s.stats),
	}

	if s.config.TLSListenAddr != "" {
		s.httpdTLS = &http.Server{
			Addr:    s.config.TLSListenAddr,
			Handler: NewHTTPProxy(cfg.Domain, s.upstreams, s.stats),
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

		target := &ev.Target

		switch strings.ToLower(ev.Change) {
		case "add", "change":
			if err := target.valid(); err != nil {
				log.Errorln("invalid event target:", err)
				continue
			}
			if err := s.upstreams.upsertTarget(target); err != nil {
				log.Errorln("upstream upsert error:", err)
			}

		case "del":
			s.upstreams.removeTarget(target)
			s.stats.del(target.AppID, target.TaskID)

		default:
			log.Warnln("unrecognized event change type", ev.Change)
		}
	}

	panic("event channel closed, never be here")
}
