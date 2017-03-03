package event

import (
	"sync"

	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)

type EventBus struct {
	Subscribers map[string]EventSubscriber

	eventChan chan *Event

	stopC chan struct{}
	Lock  sync.Mutex
}

var eventBusInstance *EventBus

func Init() {
	eventBusInstance = &EventBus{
		Subscribers: make(map[string]EventSubscriber),
		eventChan:   make(chan *Event, 1024),
		stopC:       make(chan struct{}, 1),
		Lock:        sync.Mutex{},
	}
}

func Start(ctx context.Context) error {
	for {
		select {
		case e := <-eventBusInstance.eventChan:
			for _, subscriber := range eventBusInstance.Subscribers {
				if subscriber.InterestIn(e) {
					if err := subscriber.Write(e); err != nil {
						logrus.Debugf("write event e %s to %s got error: %s", e, subscriber, err)
					} else {
						logrus.Debugf("write event e %s to %s", e, subscriber)
					}
				} else {
					logrus.Debugf("subscriber %s have no interest in %s", subscriber, e)
				}
			}

		case <-eventBusInstance.stopC:
			return nil

		case <-ctx.Done():
			return nil
		}
	}
}

func Stop() {
	eventBusInstance.stopC <- struct{}{}
}

func WriteEvent(e *Event) {
	eventBusInstance.eventChan <- e
}

func RegistSubscriber(subscriber EventSubscriber) {
	eventBusInstance.Lock.Lock()
	defer eventBusInstance.Lock.Unlock()

	eventBusInstance.Subscribers[subscriber.GetKey()] = subscriber
	return
}

func UnRegistSubcriber(subscriber EventSubscriber) {
	eventBusInstance.Lock.Lock()
	defer eventBusInstance.Lock.Unlock()

	delete(eventBusInstance.Subscribers, subscriber.GetKey())
	return
}
