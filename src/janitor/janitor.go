package janitor

import (
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	proxyproto "github.com/armon/go-proxyproto"

	"github.com/Dataman-Cloud/swan/src/config"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type JanitorServer struct {
	upstreams  *Upstreams
	eventChan  chan *TargetChangeEvent
	stats      *Stats
	httpServer *http.Server
	config     *config.Janitor
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

	s.httpServer = &http.Server{
		Handler: NewHTTPProxy(cfg.Domain, s.upstreams, s.stats)}

	return s
}

func (s *JanitorServer) EmitChange(ev *TargetChangeEvent) {
	s.eventChan <- ev
}

func (s *JanitorServer) Start() error {
	ln, err := net.Listen("tcp", s.config.ListenAddr)
	if err != nil {
		return err
	}

	go s.watchEvent()

	defer s.httpServer.Close()
	return s.httpServer.Serve(&proxyproto.Listener{
		Listener: TcpKeepAliveListener{ln.(*net.TCPListener)}})
}

func (s *JanitorServer) watchEvent() {
	log.Println("proxy listening on app event ...")

	for ev := range s.eventChan {
		log.Printf("proxy caught event: %s", ev)

		target := &ev.Target

		switch strings.ToLower(ev.Change) {
		case "add", "change":
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

// TcpKeepAliveListener enable TCP-KEEPALIVE for each incomming tcp conn
type TcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln TcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	if err = tc.SetKeepAlive(true); err != nil {
		return
	}
	if err = tc.SetKeepAlivePeriod(1 * time.Minute); err != nil {
		return
	}
	return tc, nil
}
