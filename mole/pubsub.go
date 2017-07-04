package mole

import (
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
)

var (
	pub *publisher // global publisher
)

func init() {
	pub = newPublisher(time.Second*5, 1024)
}

type publisher struct {
	mux sync.RWMutex            // protect m
	m   map[subcriber]topicFunc // hold all of subcribers

	timeout   time.Duration // send topic timeout
	bufferLen int           // buffer length of each subcriber channel
}
type subcriber chan interface{}
type topicFunc func(v interface{}) bool

func newPublisher(timeout time.Duration, bufferLen int) *publisher {
	return &publisher{
		m:         make(map[subcriber]topicFunc),
		timeout:   timeout,
		bufferLen: bufferLen,
	}
}

// SubcribeAll adds a new subscriber that receive all messages.
func (p *publisher) SubcribeAll() subcriber {
	return p.Subcribe(nil)
}

// Subcribe adds a new subscriber that filters messages sent by a topic.
func (p *publisher) Subcribe(tf topicFunc) subcriber {
	ch := make(subcriber, p.bufferLen)
	p.mux.Lock()
	p.m[ch] = tf
	p.mux.Unlock()
	return ch
}

// Evict removes the specified subscriber from receiving any more messages.
func (p *publisher) Evict(sub subcriber) {
	p.mux.Lock()
	delete(p.m, sub)
	close(sub)
	p.mux.Unlock()
}

func (p *publisher) Publish(v interface{}) {
	p.mux.RLock()
	defer p.mux.RUnlock()

	var wg sync.WaitGroup
	wg.Add(len(p.m))
	// broadcasting with concurrency
	for sub, tf := range p.m {
		go func(sub subcriber, v interface{}, tf topicFunc) {
			defer wg.Done()
			p.send(sub, v, tf)
		}(sub, v, tf)
	}
	wg.Wait()
}

func (p *publisher) send(sub subcriber, v interface{}, tf topicFunc) {
	// if a subcriber setup topic filter func and not matched by the topic filter
	// skip send message to this subcriber
	if tf != nil && !tf(v) {
		return
	}

	// send with timeout
	if p.timeout > 0 {
		select {
		case sub <- v:
		case <-time.After(p.timeout):
			log.Println("send to subcriber timeout after", p.timeout.String())
		}
		return
	}

	// directely send
	sub <- v
}
