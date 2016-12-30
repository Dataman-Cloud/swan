package event

import (
	"errors"
	"strings"

	"github.com/Dataman-Cloud/swan-resolver/nameserver"
)

type DNSSubscriber struct {
	Key      string
	Resolver *nameserver.Resolver
}

func NewDNSSubscriber(resolver *nameserver.Resolver) *DNSSubscriber {
	subscriber := &DNSSubscriber{
		Resolver: resolver,
		Key:      "dns",
	}

	return subscriber
}

func (subscriber *DNSSubscriber) Subscribe(bus *EventBus) error {
	bus.Subscribers[subscriber.Key] = subscriber
	return nil
}

func (subscriber *DNSSubscriber) Unsubscribe(bus *EventBus) error {
	delete(bus.Subscribers, subscriber.Key)
	return nil
}

func (subscriber *DNSSubscriber) Write(e *Event) error {
	payload, ok := e.Payload.(*TaskInfo)
	if !ok {
		return errors.New("payload type error")
	}

	rgEvent := &nameserver.RecordGeneratorChangeEvent{}
	if e.Type == EventTypeTaskAdd {
		rgEvent.Change = "add"
	} else {
		rgEvent.Change = "del"
	}

	rgEvent.Type = payload.Type
	rgEvent.Ip = payload.Ip
	rgEvent.Port = payload.Port
	rgEvent.DomainPrefix = strings.ToLower(strings.Replace(payload.TaskId, "-", ".", -1))

	subscriber.Resolver.RecordGeneratorChangeChan() <- rgEvent
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
