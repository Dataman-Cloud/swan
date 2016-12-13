package service

import (
	"net/http"
	"sync"
	"time"

	"github.com/Dataman-Cloud/swan-janitor/src/upstream"
	"github.com/armon/go-proxyproto"

	log "github.com/Sirupsen/logrus"
)

const (
	SESSION_RENEW_INTERVAL = time.Second * 50
)

type ServicePod struct {
	Key upstream.UpstreamKey

	Manager    *ServiceManager
	HttpServer *http.Server
	Listener   *proxyproto.Listener

	upstream           *upstream.Upstream
	sessionIDWithTTY   string
	sessionRenewTicker *time.Ticker
	stopCh             chan bool
	lock               sync.Mutex
}

func NewServicePod(u *upstream.Upstream, manager *ServiceManager) (*ServicePod, error) {
	pod := &ServicePod{
		Key: u.Key(),

		stopCh:   make(chan bool, 1),
		upstream: u,
		Manager:  manager,
	}
	return pod, nil
}

func NewSingleServicePod(manager *ServiceManager) (*ServicePod, error) {
	pod := &ServicePod{
		stopCh:  make(chan bool, 1),
		Manager: manager,
	}
	return pod, nil
}

func (pod *ServicePod) Run() {
	go func() {
		log.Infof("start runing pod now %s", pod.Key)
		err := pod.HttpServer.Serve(pod.Listener)
		if err != nil {
			log.Errorf("pod run goroutine error  <%s>,  the error is [%s]", pod.Key, err)
		}
		log.Infof("end runing pod now %s", pod.Key)
	}()
}
