package event

import (
	"errors"
	"strings"

	"github.com/Dataman-Cloud/swan-janitor/src/janitor"
	"github.com/Dataman-Cloud/swan-janitor/src/upstream"
)

type JanitorSubscriber struct {
	Key      string
	Resolver *janitor.JanitorServer
}

func NewJanitorSubscriber(resolver *janitor.JanitorServer) *JanitorSubscriber {
	janitorSubscriber := &JanitorSubscriber{
		Resolver: resolver,
		Key:      "janitor",
	}
	return janitorSubscriber
}

func (js *JanitorSubscriber) Subscribe(bus *EventBus) error {
	bus.Subscribers[js.Key] = js
	return nil
}

func (js *JanitorSubscriber) Unsubscribe(bus *EventBus) error {
	delete(bus.Subscribers, js.Key)
	return nil
}

func (js *JanitorSubscriber) Write(e *Event) error {
	payload, ok := e.Payload.(*TaskInfo)
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

	js.Resolver.SwanEventChan() <- rgevent
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
