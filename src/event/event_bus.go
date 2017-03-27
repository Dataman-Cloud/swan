package event

import (
	"errors"
	"sync"

	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)

type EventBus struct {
	listeners map[string]EventListener

	eventChan chan *Event

	stopC chan struct{}
	Lock  sync.Mutex
}

var initOnce sync.Once
var eventBusInstance *EventBus

func Instance() *EventBus {
	return eventBusInstance
}

func Init() *EventBus {
	initOnce.Do(func() {
		eventBusInstance = &EventBus{
			listeners: make(map[string]EventListener),
			eventChan: make(chan *Event, 1024),
			stopC:     make(chan struct{}, 1),
			Lock:      sync.Mutex{},
		}
	})

	return eventBusInstance
}

func Start(ctx context.Context) error {
	for {
		select {
		case e := <-eventBusInstance.eventChan:
			for _, listener := range eventBusInstance.listeners {
				if listener.InterestIn(e) {
					if err := listener.Write(e); err != nil {
						logrus.Debugf("write event e %s to %s got error: %s", e, listener, err)
					} else {
						logrus.Debugf("write event e %s to %s", e, listener)
					}
				} else {
					logrus.Debugf("listener %s have no interest in %s", listener, e)
				}
			}

		case <-eventBusInstance.stopC:
			return errors.New("eventBusInstance bye")

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func Stop() {
	eventBusInstance.stopC <- struct{}{}
}

func WriteEvent(e *Event) {
	eventBusInstance.eventChan <- e
}

func AddListener(listener EventListener) {
	eventBusInstance.Lock.Lock()
	defer eventBusInstance.Lock.Unlock()

	eventBusInstance.listeners[listener.Key()] = listener
	return
}

func RemoveListener(listener EventListener) {
	eventBusInstance.Lock.Lock()
	defer eventBusInstance.Lock.Unlock()

	delete(eventBusInstance.listeners, listener.Key())
	return
}
