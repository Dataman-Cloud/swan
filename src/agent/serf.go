package agent

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/hashicorp/serf/serf"
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

	fields := strings.Split(ss.SerfListenAddr, ":")
	_bindAddr, _bindPort := fields[0], fields[1]
	ss.serfConfig.MemberlistConfig.BindAddr = _bindAddr

	bindPort, _ := strconv.Atoi(_bindPort)
	ss.serfConfig.MemberlistConfig.BindPort = bindPort

	hostname, _ := os.Hostname()
	ss.serfConfig.NodeName = fmt.Sprintf("%s-%d", hostname, bindPort)

	return ss
}

func (ss *SerfServer) Start() error {
	var err error

	ss.SerfNode, err = serf.Create(ss.serfConfig)
	if err != nil {
		return err
	}

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
