package agent

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/hashicorp/serf/serf"
	"golang.org/x/net/context"
)

type SerfServer struct {
	serfConfig *serf.Config
	SerfNode   *serf.Serf
	EventCh    chan serf.Event

	SerfListenAddr string
	SerfJoinAddr   string
}

func NewSerfServer(listenerAddr, joinAddr string) *SerfServer {
	ss := &SerfServer{
		SerfListenAddr: listenerAddr,
		SerfJoinAddr:   joinAddr,
	}

	ss.serfConfig = serf.DefaultConfig()

	ss.EventCh = make(chan serf.Event, 64)
	ss.serfConfig.EventCh = ss.EventCh

	_bindAddr, _bindPort := strings.Split(ss.SerfListenAddr, ":")[0], strings.Split(ss.SerfListenAddr, ":")[1]
	ss.serfConfig.MemberlistConfig.BindAddr = _bindAddr

	bindPort, _ := strconv.Atoi(_bindPort)
	ss.serfConfig.MemberlistConfig.BindPort = bindPort

	hostname, _ := os.Hostname()
	ss.serfConfig.NodeName = fmt.Sprintf("%s-%d", hostname, bindPort)

	return ss
}

func (ss *SerfServer) Start(ctx context.Context, started chan bool) error {
	var err error
	ss.SerfNode, err = serf.Create(ss.serfConfig)
	if err != nil {
		return err
	}

	go func() {
		started <- true
	}()

	if ss.SerfJoinAddr != "" {
		_, err := ss.SerfNode.Join([]string{ss.SerfJoinAddr}, false)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ss *SerfServer) Publish(eventType string, payload []byte) {
	ss.SerfNode.UserEvent(eventType, payload, true)
}

//for {
//event := <-eventCh
//userEvent, ok := event.(serf.UserEvent)
//if ok {
//var taskInfoEvent types.TaskInfoEvent
//err = json.Unmarshal(userEvent.Payload, &taskInfoEvent)
//if err != nil {
//logrus.Errorf("unmarshal taskInfoEvent go error: %s", err.Error())
//}

//agent.Resolver.RecordGeneratorChangeChan() <- recordGeneratorChangeEventFromTaskInfoEvent(userEvent.Name, &taskInfoEvent)
//agent.Janitor.SwanEventChan() <- janitorTargetgChangeEventFromTaskInfoEvent(userEvent.Name, &taskInfoEvent)
//}
//}
