package janitor

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	proxyproto "github.com/armon/go-proxyproto"
	"golang.org/x/net/context"

	"github.com/Dataman-Cloud/swan/src/config"
)

type JanitorServer struct {
	config *config.Janitor

	UpstreamLoader *UpstreamLoader
	EventChan      chan *TargetChangeEvent

	httpServer *http.Server
	P          *Prometheus
}

func NewJanitorServer(Config *config.Janitor) *JanitorServer {
	s := &JanitorServer{
		config: Config,
	}

	s.EventChan = make(chan *TargetChangeEvent, 1024)
	s.UpstreamLoader = NewUpstreamLoader(s.EventChan)

	s.P = &Prometheus{
		MetricsPath: "/gateway-metrics",
	}
	s.P.registerMetrics(fmt.Sprintf("gateway_%s", strings.Replace(strings.Replace(Config.ListenAddr, ".", "_", -1), ":", "_", -1)))

	s.httpServer = &http.Server{Handler: NewLayer7Proxy(&http.Transport{},
		s.config,
		s.UpstreamLoader,
		s.P,
	)}

	return s
}

func (s *JanitorServer) Start(ctx context.Context, started chan bool) error {
	ln, err := net.Listen("tcp", s.config.ListenAddr)
	if err != nil {
		return err
	}

	errCh := make(chan error)
	go func() {
		errCh <- s.UpstreamLoader.Start(ctx)
	}()

	go func() {
		defer s.httpServer.Close()
		errCh <- s.httpServer.Serve(&proxyproto.Listener{Listener: TcpKeepAliveListener{ln.(*net.TCPListener)}})
	}()

	go func() {
		started <- true
	}()

	return <-errCh
}
