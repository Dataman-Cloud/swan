package service

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/Dataman-Cloud/swan-janitor/src/handler"
	"github.com/Dataman-Cloud/swan-janitor/src/listener"
	"github.com/Dataman-Cloud/swan-janitor/src/upstream"

	"golang.org/x/net/context"
)

const (
	SERVICE_ACTIVITIES_PREFIX = "SA"
	SERVICE_ENTRIES_PREFIX    = "SE"
)

type ServiceManager struct {
	servicePods map[upstream.UpstreamKey]*ServicePod

	handlerFactory  *handler.Factory
	listenerManager *listener.Manager
	upstreamLoader  upstream.UpstreamLoader
	ctx             context.Context
	forkMutex       sync.Mutex
	rwMutex         sync.RWMutex
}

func NewServiceManager(ctx context.Context) *ServiceManager {
	serviceManager := &ServiceManager{}

	handlerFactory_ := ctx.Value(handler.HANDLER_FACTORY_KEY)
	serviceManager.handlerFactory = handlerFactory_.(*handler.Factory)

	listenerManager_ := ctx.Value(listener.LISTENER_MANAGER_KEY)
	serviceManager.listenerManager = listenerManager_.(*listener.Manager)

	switch upstream.UpstreamLoaderKey {
	case upstream.SWAN_UPSTREAM_LOADER_KEY:
		serviceManager.upstreamLoader = ctx.Value(upstream.SWAN_UPSTREAM_LOADER_KEY).(*upstream.SwanUpstreamLoader)
	}
	serviceManager.servicePods = make(map[upstream.UpstreamKey]*ServicePod)
	serviceManager.ctx = ctx

	return serviceManager
}

func (manager *ServiceManager) FetchDefaultServicePod() (*ServicePod, error) {
	manager.forkMutex.Lock()
	defer manager.forkMutex.Unlock()
	pod, err := NewSingleServicePod(manager)
	if err != nil {
		fmt.Println("fails to setup default service pod")
		return nil, err
	}
	// fetch default listener then assign it to pod
	pod.Listener = manager.listenerManager.DefaultListener()

	// fetch a http handler then assign it to pod
	var u *upstream.Upstream
	pod.HttpServer = &http.Server{Handler: manager.handlerFactory.HttpHandler(u)}

	manager.servicePods[manager.listenerManager.DefaultUpstreamKey()] = pod
	return pod, nil
}

func (manager *ServiceManager) ForkOrFetchNewServicePod(us *upstream.Upstream) (*ServicePod, error) {
	manager.forkMutex.Lock()
	defer manager.forkMutex.Unlock()

	pod, found := manager.servicePods[us.Key()]
	if found {
		return pod, nil
	}

	pod, err := NewServicePod(us, manager)
	if err != nil {
		return nil, err
	}

	// fetch a listener then assign it to pod
	pod.Listener, err = manager.listenerManager.FetchListener(us.Key())
	if err != nil {
		fmt.Sprintf("[ERRO] fetch a listener error:%s", err.Error())
		return nil, err
	}

	// fetch a http handler then assign it to pod
	pod.HttpServer = &http.Server{Handler: manager.handlerFactory.HttpHandler(us)}

	manager.servicePods[us.Key()] = pod
	return pod, nil
}

func (manager *ServiceManager) KillServicePod(u *upstream.Upstream) error {
	manager.rwMutex.Lock()
	defer manager.rwMutex.Unlock()

	_, found := manager.servicePods[u.Key()]
	if found {
		delete(manager.servicePods, u.Key())
		manager.upstreamLoader.Remove(u)
		manager.listenerManager.Remove(u.Key())
	}
	return nil
}

func (manager *ServiceManager) PortsOccupied() []string {
	ports := make([]string, 0)
	for key, _ := range manager.servicePods {
		ports = append(ports, key.Port)
	}
	return ports
}
