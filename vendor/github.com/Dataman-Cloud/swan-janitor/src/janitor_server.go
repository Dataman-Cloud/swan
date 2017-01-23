package janitor

import (
	"net"
	"net/http"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/armon/go-proxyproto"
	"golang.org/x/net/context"
)

type JanitorServer struct {
	swanUpstreamLoader *SwanUpstreamLoader
	HttpServer         *http.Server
	Listener           *proxyproto.Listener

	ctx    context.Context
	config Config
}

func NewJanitorServer(Config Config) *JanitorServer {
	server := &JanitorServer{
		config: Config,
		ctx:    context.Background(),
	}
	return server
}

func (server *JanitorServer) ServerInit() *JanitorServer {
	log.Info("Janitor Server Initialing")
	var err error
	server.swanUpstreamLoader, err = SwanUpstreamLoaderInit()
	if err != nil {
		log.Fatalf("Setup Upstream Loader Got err: %s", err)
	}

	ln, err := net.Listen("tcp", net.JoinHostPort(server.config.Listener.IP, server.config.Listener.DefaultPort))
	if err != nil {
		log.Errorf("%s", err)
		return nil
	}

	server.Listener = &proxyproto.Listener{Listener: TcpKeepAliveListener{ln.(*net.TCPListener)}}
	if server.Listener == nil {
		log.Fatalf("failed to listen port")
		os.Exit(1)
	}
	server.HttpServer = &http.Server{Handler: NewHTTPProxy(&http.Transport{},
		server.config.HttpHandler,
		server.config.Listener,
		server.swanUpstreamLoader)}

	if server.HttpServer == nil {
		log.Fatalf("failed to listen port")
		os.Exit(1)
	}

	return server
}

func (server *JanitorServer) UpstreamLoader() *SwanUpstreamLoader {
	return server.swanUpstreamLoader
}

func (server *JanitorServer) SwanEventChan() chan<- *TargetChangeEvent {
	return server.swanUpstreamLoader.SwanEventChan()
}

func (server *JanitorServer) Run() {
	server.HttpServer.Serve(server.Listener)
}
