package event

import (
	"github.com/Sirupsen/logrus"
)

type EventBus struct {
	Subscribers map[string]EventSubscriber

	EventChan chan *Event
}

func New() *EventBus {
	bus := &EventBus{
		Subscribers: make(map[string]EventSubscriber),
		EventChan:   make(chan *Event, 1024),
	}

	return bus
}

func (bus *EventBus) Start() error {
	for {
		select {
		case e := <-bus.EventChan:
			for _, subscriber := range bus.Subscribers {
				if subscriber.InterestIn(e) {
					subscriber.Write(e)
					logrus.Infof("write event e %s to %s", e, subscriber)
				}
			}
		}
	}
}
