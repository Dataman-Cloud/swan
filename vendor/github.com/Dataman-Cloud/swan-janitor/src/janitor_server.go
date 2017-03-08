package janitor

import (
	"fmt"
	"net"
	"net/http"

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

func (server *JanitorServer) Init() error {
	log.Info("Janitor Server Initialing")
	var err error
	server.swanUpstreamLoader, err = SwanUpstreamLoaderInit()
	if err != nil {
		log.Fatalf("Setup Upstream Loader Got err: %s", err)
	}

	ln, err := net.Listen("tcp", server.config.ListenAddr)
	if err != nil {
		log.Errorf("%s", err)
		return err
	}

	server.Listener = &proxyproto.Listener{Listener: TcpKeepAliveListener{ln.(*net.TCPListener)}}
	if server.Listener == nil {
		return fmt.Errorf("failed to listen port")
	}
	server.HttpServer = &http.Server{Handler: NewHTTPProxy(&http.Transport{},
		server.config.HttpHandler,
		server.config.ListenAddr,
		server.swanUpstreamLoader)}

	if server.HttpServer == nil {
		return fmt.Errorf("server.HttpServer not initialized")
	}

	return nil
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
