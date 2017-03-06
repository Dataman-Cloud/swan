package event

import (
	"net/http"

	"github.com/manucorporat/sse"
)

type SSEListener struct {
	key    string
	appId  string
	rw     http.ResponseWriter
	doneCh chan struct{}
}

func NewSSEListener(key string, appId string, rw http.ResponseWriter) (*SSEListener, chan struct{}) {
	sseListener := &SSEListener{
		key:    key,
		appId:  appId,
		rw:     rw,
		doneCh: make(chan struct{}),
	}

	if c, ok := sseListener.rw.(http.CloseNotifier); ok {
		go func() {
			for {
				select {
				case <-c.CloseNotify():
					sseListener.doneCh <- struct{}{}
				}
			}
		}()
	} else { // exit fast
		sseListener.doneCh <- struct{}{}
	}

	return sseListener, sseListener.doneCh
}

func (ssel *SSEListener) Key() string {
	return ssel.key
}

func (ssel *SSEListener) Write(e *Event) error {
	sse.Encode(ssel.rw, sse.Event{
		Id:    e.ID,
		Event: e.Type,
		Data:  e.Payload,
	})

	// TODO
	if f, ok := ssel.rw.(http.Flusher); ok {
		f.Flush()
	}

	return nil
}

func (sse *SSEListener) InterestIn(e *Event) bool {
	if sse.appId != "" {
		if sse.appId == e.AppID {
			return true
		} else {
			return false
		}
	} else {
		return true
	}
}
