package event

import (
	"encoding/json"
	"errors"
	"strings"
	"sync"

	"github.com/Dataman-Cloud/swan-resolver/nameserver"
	"github.com/Dataman-Cloud/swan/src/types"
	"github.com/Sirupsen/logrus"
)

type DNSSubscriber struct {
	Key          string
	acceptors    map[string]types.ResolverAcceptor
	acceptorLock sync.RWMutex
}

func NewDNSSubscriber() *DNSSubscriber {
	subscriber := &DNSSubscriber{
		Key:       "dns",
		acceptors: make(map[string]types.ResolverAcceptor),
	}

	return subscriber
}

func (subscriber *DNSSubscriber) Subscribe(bus *EventBus) error {
	bus.Lock.Lock()
	defer bus.Lock.Unlock()

	bus.Subscribers[subscriber.Key] = subscriber
	return nil
}

func (subscriber *DNSSubscriber) Unsubscribe(bus *EventBus) error {
	bus.Lock.Lock()
	defer bus.Lock.Unlock()

	delete(bus.Subscribers, subscriber.Key)
	return nil
}

func (subscriber *DNSSubscriber) AddAcceptor(acceptor types.ResolverAcceptor) {
	subscriber.acceptorLock.Lock()
	subscriber.acceptors[acceptor.ID] = acceptor
	subscriber.acceptorLock.Unlock()
}

func (subscriber *DNSSubscriber) Write(e *Event) error {
	payload, ok := e.Payload.(*TaskInfoEvent)
	if !ok {
		return errors.New("payload type error")
	}

	rgEvent := &nameserver.RecordGeneratorChangeEvent{}
	if e.Type == EventTypeTaskAdd {
		rgEvent.Change = "add"
	} else {
		rgEvent.Change = "del"
	}

	rgEvent.Type = "srv"
	rgEvent.Ip = payload.Ip
	rgEvent.Port = payload.Port
	rgEvent.DomainPrefix = strings.ToLower(strings.Replace(payload.TaskId, "-", ".", -1))

	go subscriber.pushResloverEvent(rgEvent)
	return nil
}

func (subscriber *DNSSubscriber) InterestIn(e *Event) bool {
	if e.Type == EventTypeTaskAdd {
		return true
	}

	if e.Type == EventTypeTaskRm {
		return true
	}

	return false
}

func (subscriber *DNSSubscriber) pushResloverEvent(event *nameserver.RecordGeneratorChangeEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		logrus.Infof("marshal reslover event got error: %s", err.Error())
		return
	}

	subscriber.acceptorLock.RLock()
	for _, acceptor := range subscriber.acceptors {
		if err := sendEventByHttp(acceptor.RemoteAddr, "POST", data); err != nil {
			logrus.Infof("send reslover event by http to %s got error: %s", acceptor.RemoteAddr, err.Error())
		}
	}
	subscriber.acceptorLock.RUnlock()
}
