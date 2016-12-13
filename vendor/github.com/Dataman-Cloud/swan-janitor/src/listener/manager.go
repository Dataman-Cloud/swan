package listener

import (
	"net"

	"github.com/Dataman-Cloud/swan-janitor/src/config"
	"github.com/Dataman-Cloud/swan-janitor/src/upstream"
	"github.com/armon/go-proxyproto"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)

const (
	LISTENER_MANAGER_KEY = "listener_manager"
)

type Manager struct {
	Mode      string
	Listeners map[upstream.UpstreamKey]*proxyproto.Listener
	Config    config.Listener
}

func InitManager(Config config.Listener) (*Manager, error) {
	manager := &Manager{}
	manager.Listeners = make(map[upstream.UpstreamKey]*proxyproto.Listener)
	manager.Config = Config

	switch manager.Config.Mode {
	case config.SINGLE_LISTENER_MODE:
		setupSingleListener(manager)
	case config.MULTIPORT_LISTENER_MODE:
		// Do nothing
	}

	return manager, nil
}

func ManagerFromContext(ctx context.Context) *Manager {
	manager := ctx.Value(LISTENER_MANAGER_KEY)
	return manager.(*Manager)
}

func (manager *Manager) Shutdown() {
	for _, listener := range manager.Listeners {
		listener.Close()
	}
}

func (manager *Manager) DefaultUpstreamKey() upstream.UpstreamKey {
	return upstream.UpstreamKey{Ip: manager.Config.IP, Port: manager.Config.DefaultPort}
}

func (manager *Manager) DefaultListener() *proxyproto.Listener {
	return manager.Listeners[manager.DefaultUpstreamKey()]
}

func setupSingleListener(manager *Manager) error {
	ln, err := net.Listen("tcp", net.JoinHostPort(manager.Config.IP, manager.Config.DefaultPort))
	if err != nil {
		log.Errorf("%s", err)
		return err
	}

	manager.Listeners[manager.DefaultUpstreamKey()] = &proxyproto.Listener{Listener: TcpKeepAliveListener{ln.(*net.TCPListener)}}
	return nil
}

func (manager *Manager) FetchListener(key upstream.UpstreamKey) (*proxyproto.Listener, error) {
	listener := manager.Listeners[key]
	if listener == nil {
		ln, err := net.Listen("tcp", net.JoinHostPort(key.Ip, key.Port))
		if err != nil {
			log.Errorf("%s", err)
			return nil, err
		}

		manager.Listeners[key] = &proxyproto.Listener{Listener: TcpKeepAliveListener{ln.(*net.TCPListener)}}
	}

	return manager.Listeners[key], nil
}

func (manager *Manager) Remove(key upstream.UpstreamKey) {
	l, ok := manager.Listeners[key]
	if ok {
		err := l.Close()
		if err != nil {
			log.Error("close a already closed listener")
		}
	}

	delete(manager.Listeners, key)
}

func (manager *Manager) ListeningPorts() []string {
	var ports []string
	for key, _ := range manager.Listeners {
		ports = append(ports, key.Port)
	}

	return ports
}
