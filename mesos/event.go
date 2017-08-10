package mesos

import (
	"io"
	"net/http"
	"sync"

	log "github.com/Sirupsen/logrus"
)

type event interface {
	Format() []byte
}

type eventClient struct {
	w io.Writer
	f http.Flusher
	n http.CloseNotifier

	wait chan struct{}
	recv chan []byte
}

type eventManager struct {
	sync.RWMutex                         // protect m
	m            map[string]*eventClient // store of online event clients
	max          int                     // max nb of clients, avoid bomber
}

func NewEventManager() *eventManager {
	return &eventManager{
		m:   make(map[string]*eventClient),
		max: 1024,
	}
}

func (em *eventManager) clients() map[string]*eventClient {
	em.RLock()
	defer em.RUnlock()

	return em.m
}

// broadcast message to all event clients
func (em *eventManager) broadcast(e event) error {
	for _, c := range em.clients() {
		select {
		case c.recv <- e.Format():
		default:
		}
	}
	return nil
}

// subscribe() add an event client
func (em *eventManager) subscribe(remoteAddr string, w io.Writer) {
	c := &eventClient{
		w: w,
		f: w.(http.Flusher),
		n: w.(http.CloseNotifier),

		wait: make(chan struct{}),
		recv: make(chan []byte, 1024),
	}

	go func(em *eventManager, c *eventClient, remoteAddr string) {
		defer em.evict(remoteAddr)
		for {
			select {
			case <-c.n.CloseNotify():
				return
			case msg := <-c.recv:
				if _, err := c.w.Write(msg); err != nil {
					log.Errorf("write event message to client [%s] error: [%v]", remoteAddr, err)
					return
				}
				c.f.Flush()
			}
		}
	}(em, c, remoteAddr)

	em.Lock()
	em.m[remoteAddr] = c
	em.Unlock()
}

func (em *eventManager) wait(remoteAddr string) {
	em.RLock()
	c, ok := em.m[remoteAddr]
	em.RUnlock()
	if !ok {
		return
	}
	<-c.wait
}

func (em *eventManager) evict(remoteAddr string) {
	log.Debugln("evict event listener ", remoteAddr)

	em.Lock()
	defer em.Unlock()

	c, ok := em.m[remoteAddr]
	if !ok {
		return
	}

	close(c.wait)
	delete(em.m, remoteAddr)
}

func (em *eventManager) Full() bool {
	return em.size() >= int(em.max)
}

func (em *eventManager) size() int {
	em.RLock()
	defer em.RUnlock()

	return len(em.m)
}
