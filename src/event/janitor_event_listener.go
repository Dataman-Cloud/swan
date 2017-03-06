package event

import (
	"encoding/json"
	"sync"

	"github.com/Dataman-Cloud/swan/src/types"

	"github.com/Dataman-Cloud/swan-janitor/src"
	"github.com/Sirupsen/logrus"
)

type JanitorListener struct {
	key          string
	acceptors    map[string]types.JanitorAcceptor
	acceptorLock sync.RWMutex
}

func NewJanitorListener() *JanitorListener {
	janitorListener := &JanitorListener{
		key:       "janitor",
		acceptors: make(map[string]types.JanitorAcceptor),
	}
	return janitorListener
}

func (jl *JanitorListener) Key() string {
	return jl.key
}

func (jl *JanitorListener) AddAcceptor(acceptor types.JanitorAcceptor) {
	jl.acceptorLock.Lock()
	jl.acceptors[acceptor.ID] = acceptor
	jl.acceptorLock.Unlock()
}

func (jl *JanitorListener) RemoveAcceptor(ID string) {
	jl.acceptorLock.Lock()
	delete(jl.acceptors, ID)
	jl.acceptorLock.Unlock()
}

func (jl *JanitorListener) Write(e *Event) error {
	janitorEvent, err := BuildJanitorEvent(e)
	if err != nil {
		return err
	}

	go jl.pushJanitorEvent(janitorEvent)

	return nil
}

func (jl *JanitorListener) InterestIn(e *Event) bool {
	if e.AppMode != "replicates" {
		return false
	}

	if e.Type == EventTypeTaskHealthy {
		return true
	}

	if e.Type == EventTypeTaskUnhealthy {
		return true
	}

	return false
}

func (jl *JanitorListener) pushJanitorEvent(event *janitor.TargetChangeEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		logrus.Infof("marshal janitor event got error: %s", err.Error())
		return
	}

	jl.acceptorLock.RLock()
	for _, acceptor := range jl.acceptors {
		if err := SendEventByHttp(acceptor.RemoteAddr, "POST", data); err != nil {
			logrus.Infof("send janitor event by http to %s got error: %s", acceptor.RemoteAddr, err.Error())
		}
	}
	jl.acceptorLock.RUnlock()
}
