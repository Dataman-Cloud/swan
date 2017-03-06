package event

import (
	"encoding/json"
	"sync"

	"github.com/Dataman-Cloud/swan-resolver/nameserver"
	"github.com/Dataman-Cloud/swan/src/types"
	"github.com/Sirupsen/logrus"
)

type DNSListener struct {
	key          string
	acceptors    map[string]types.ResolverAcceptor
	acceptorLock sync.RWMutex
}

func NewDNSListener() *DNSListener {
	listener := &DNSListener{
		key:       "dns",
		acceptors: make(map[string]types.ResolverAcceptor),
	}

	return listener
}

func (listener *DNSListener) Key() string {
	return listener.key
}

func (listener *DNSListener) AddAcceptor(acceptor types.ResolverAcceptor) {
	listener.acceptorLock.Lock()
	listener.acceptors[acceptor.ID] = acceptor
	listener.acceptorLock.Unlock()
}

func (listener *DNSListener) RemoveAcceptor(ID string) {
	listener.acceptorLock.Lock()
	delete(listener.acceptors, ID)
	listener.acceptorLock.Unlock()
}

func (listener *DNSListener) Write(e *Event) error {
	rgEvent, err := BuildResolverEvent(e)
	if err != nil {
		return err
	}

	go listener.pushResloverEvent(rgEvent)

	return nil
}

func (listener *DNSListener) InterestIn(e *Event) bool {
	if e.Type == EventTypeTaskHealthy {
		return true
	}

	if e.Type == EventTypeTaskUnhealthy {
		return true
	}

	return false
}

func (listener *DNSListener) pushResloverEvent(event *nameserver.RecordGeneratorChangeEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		logrus.Infof("marshal reslover event got error: %s", err.Error())
		return
	}

	listener.acceptorLock.RLock()
	for _, acceptor := range listener.acceptors {
		if err := SendEventByHttp(acceptor.RemoteAddr, "POST", data); err != nil {
			logrus.Infof("send reslover event by http to %s got error: %s", acceptor.RemoteAddr, err.Error())
		} else {
			logrus.Debugf("send reslover event by http to %s success", acceptor.RemoteAddr)
		}
	}
	listener.acceptorLock.RUnlock()
}
