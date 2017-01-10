package event

import (
	"encoding/json"
	"errors"
	"strings"
	"sync"

	"github.com/Dataman-Cloud/swan/src/types"
	"github.com/Sirupsen/logrus"

	"github.com/Dataman-Cloud/swan-janitor/src/upstream"
)

type JanitorSubscriber struct {
	Key          string
	acceptors    map[string]types.JanitorAcceptor
	acceptorLock sync.RWMutex
}

func NewJanitorSubscriber() *JanitorSubscriber {
	janitorSubscriber := &JanitorSubscriber{
		Key:       "janitor",
		acceptors: make(map[string]types.JanitorAcceptor),
	}
	return janitorSubscriber
}

func (js *JanitorSubscriber) Subscribe(bus *EventBus) error {
	bus.Lock.Lock()
	defer bus.Lock.Unlock()

	bus.Subscribers[js.Key] = js
	return nil
}

func (js *JanitorSubscriber) Unsubscribe(bus *EventBus) error {
	bus.Lock.Lock()
	defer bus.Lock.Unlock()

	delete(bus.Subscribers, js.Key)
	return nil
}

func (js *JanitorSubscriber) AddAcceptor(acceptor types.JanitorAcceptor) {
	js.acceptorLock.Lock()
	js.acceptors[acceptor.ID] = acceptor
	js.acceptorLock.Unlock()
}

func (js *JanitorSubscriber) Write(e *Event) error {
	payload, ok := e.Payload.(*TaskInfoEvent)
	if !ok {
		return errors.New("payload type error")
	}

	rgevent := &upstream.TargetChangeEvent{}
	if e.Type == EventTypeTaskAdd {
		rgevent.Change = "add"
	} else {
		rgevent.Change = "del"
	}

	rgevent.TargetIP = payload.Ip
	rgevent.TargetPort = payload.Port
	rgevent.TargetName = strings.ToLower(strings.Replace(payload.TaskId, "-", ".", -1))

	go js.pushJanitorEvent(rgevent)

	return nil
}

func (js *JanitorSubscriber) InterestIn(e *Event) bool {
	if e.Type == EventTypeTaskAdd {
		return true
	}

	if e.Type == EventTypeTaskRm {
		return true
	}

	return false
}

func (js *JanitorSubscriber) pushJanitorEvent(event *upstream.TargetChangeEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		logrus.Infof("marshal janitor event got error: %s", err.Error())
		return
	}

	js.acceptorLock.RLock()
	for _, acceptor := range js.acceptors {
		if err := sendEventByHttp(acceptor.RemoteAddr, "POST", data); err != nil {
			logrus.Infof("send janitor event by http to %s got error: %s", acceptor.RemoteAddr, err.Error())
		}
	}
	js.acceptorLock.RUnlock()
}
