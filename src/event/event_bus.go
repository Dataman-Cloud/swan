package event

import (
	"sync"

	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)

const (
	numMaxClients int = 10240 // avoid bomber
)

type EventBus struct {
	listeners map[string]EventListener
	eventChan chan *Event

	sync.RWMutex
}

var eventBusInstance *EventBus

func init() {
	eventBusInstance = &EventBus{
		listeners: make(map[string]EventListener),
		eventChan: make(chan *Event, 1024),
	}
}

func Start(ctx context.Context) error {
	for {
		select {
		case e := <-eventBusInstance.eventChan: // TODO make event broadcasting concurrency
			for _, listener := range Listeners() {
				if listener.InterestIn(e) {
					if err := listener.Write(e); err != nil {
						logrus.Debugf("write event %s to %s got error: %v", e, listener.Key(), err)
					} else {
						logrus.Debugf("write event %s to %s", e, listener.Key())
					}
				} else {
					logrus.Debugf("listener %s have no interest in %s", listener.Key(), e)
				}
			}

		case <-ctx.Done():
			logrus.Info("eventbus shutdown goroutine by ctx cancel")
			return ctx.Err()
		}
	}
}

func WriteEvent(e *Event) {
	eventBusInstance.eventChan <- e
}

func Listeners() map[string]EventListener {
	eventBusInstance.RLock()
	defer eventBusInstance.RUnlock()
	return eventBusInstance.listeners
}

func size() int {
	return len(Listeners())
}

func Full() bool {
	return size() >= numMaxClients
}

func AddListener(listener EventListener) {
	eventBusInstance.Lock()
	defer eventBusInstance.Unlock()

	eventBusInstance.listeners[listener.Key()] = listener
	return
}

func RemoveListener(listener EventListener) {
	eventBusInstance.Lock()
	defer eventBusInstance.Unlock()

	delete(eventBusInstance.listeners, listener.Key())
	return
}
