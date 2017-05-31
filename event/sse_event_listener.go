package event

import (
	"errors"
	"net/http"
	"time"

	"github.com/manucorporat/sse"
)

type SSEListener struct {
	key    string
	appId  string
	rw     http.ResponseWriter
	doneCh chan struct{}
}

func NewSSEListener(key, appId string, rw http.ResponseWriter) (*SSEListener, error) {
	sseListener := &SSEListener{
		key:    key,
		appId:  appId,
		rw:     rw,
		doneCh: make(chan struct{}),
	}

	c, ok := sseListener.rw.(http.CloseNotifier)
	if !ok {
		return nil, errors.New("not implemented http CloseNotifier")
	}
	go func() {
		<-c.CloseNotify()
		close(sseListener.doneCh)
	}()

	return sseListener, nil
}

func (ssel *SSEListener) Wait() {
	<-ssel.doneCh
}

func (ssel *SSEListener) Key() string {
	return ssel.key
}

func (ssel *SSEListener) Write(e *Event) error {
	ch := make(chan struct{})
	go func() {
		sse.Encode(ssel.rw, sse.Event{
			Id:    e.ID,
			Event: e.Type,
			Data:  e.Payload,
		})
		close(ch)
	}()

	// with timeout check to avoid block
	select {
	case <-time.After(time.Second * 10):
		return errors.New("write timeout")
	case <-ch:
	}

	if f, ok := ssel.rw.(http.Flusher); ok {
		f.Flush()
	}

	return nil
}

func (sse *SSEListener) InterestIn(e *Event) bool {
	if sse.appId == "" { // subcribe all events
		return true
	}

	return sse.appId == e.AppID // subcribe specified app events
}
