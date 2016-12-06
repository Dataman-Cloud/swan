package scheduler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Dataman-Cloud/swan/src/manager/framework/event"
	"github.com/Dataman-Cloud/swan/src/mesosproto/mesos"
	"github.com/Dataman-Cloud/swan/src/mesosproto/sched"
	"github.com/Dataman-Cloud/swan/src/util"

	"github.com/Sirupsen/logrus"
	"github.com/andygrunwald/megos"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
)

type Scheduler struct {
	// mesos framework related
	ClusterId        string
	master           string
	client           *MesosHttpClient
	mastersUrls      []string
	lastHearBeatTime time.Time
	config           util.Scheduler

	MesosCallChan chan *sched.Call

	// TODO make sure this chan doesn't explode
	MesosEventChan chan *event.MesosEvent
	Framework      *mesos.FrameworkInfo
}

func NewScheduler(config util.Scheduler) *Scheduler {
	return &Scheduler{
		mastersUrls:    []string{"192.168.1.175:5050"},
		config:         config,
		MesosEventChan: make(chan *event.MesosEvent, 1024), // make this unbound in future
		MesosCallChan:  make(chan *sched.Call, 1024),
	}
}

// start starts the scheduler and subscribes to event stream
func (s *Scheduler) ConnectToMesosAndAcceptEvent() error {
	var err error
	s.Framework, err = createOrLoadFrameworkInfo(s.config)
	state, err := stateFromMasters(strings.Split(s.config.MesosMasters, ","))
	if err != nil {
		logrus.Errorf("%s, check your mesos mastger configuration", err)
		return err
	}

	s.master = state.Leader
	cluster := state.Cluster
	if cluster == "" {
		cluster = "Unnamed"
	}
	s.ClusterId = cluster
	s.client = NewHTTPClient(state.Leader, "/api/v1/scheduler")

	if err := s.subscribe(); err != nil {
		logrus.Error(err)
		return err
	}

	return nil
}

func (s *Scheduler) subscribe() error {
	logrus.Infof("Subscribe with mesos master %s", s.master)
	call := &sched.Call{
		Type: sched.Call_SUBSCRIBE.Enum(),
		Subscribe: &sched.Call_Subscribe{
			FrameworkInfo: s.Framework,
		},
	}

	if s.Framework.Id != nil {
		call.FrameworkId = &mesos.FrameworkID{
			Value: proto.String(s.Framework.Id.GetValue()),
		}
	}

	resp, err := s.Send(call)
	if err != nil {
		return err
	}

	// http might now be the default transport in future release
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Subscribe with unexpected response status: %d", resp.StatusCode)
	}

	logrus.Info(s.client.StreamID)
	go s.handleEvents(resp)

	return nil
}

func (s *Scheduler) handleEvents(resp *http.Response) {
	defer func() {
		resp.Body.Close()
	}()

	r := NewReader(resp.Body)
	dec := json.NewDecoder(r)

	for {
		event := new(sched.Event)
		if err := dec.Decode(event); err != nil {
			logrus.Errorf("Deocde event failed: %s", err)
			return
		}

		switch event.GetType() {
		case sched.Event_SUBSCRIBED:
			s.AddEvent(sched.Event_SUBSCRIBED, event)
		case sched.Event_OFFERS:
			s.AddEvent(sched.Event_OFFERS, event)
		case sched.Event_RESCIND:
			s.AddEvent(sched.Event_RESCIND, event)
		case sched.Event_UPDATE:
			s.AddEvent(sched.Event_UPDATE, event)
		case sched.Event_MESSAGE:
			s.AddEvent(sched.Event_MESSAGE, event)
		case sched.Event_FAILURE:
			s.AddEvent(sched.Event_FAILURE, event)
		case sched.Event_ERROR:
			s.AddEvent(sched.Event_ERROR, event)
		case sched.Event_HEARTBEAT:
			s.AddEvent(sched.Event_HEARTBEAT, event)
		}
	}
}

// create frameworkInfo on initial start
// OR load preexisting frameworkId make mesos believe it's a RESTART of framework
func createOrLoadFrameworkInfo(config util.Scheduler) (*mesos.FrameworkInfo, error) {
	fw := &mesos.FrameworkInfo{
		User:            proto.String(config.MesosFrameworkUser),
		Name:            proto.String("swan"),
		Hostname:        proto.String(config.Hostname),
		FailoverTimeout: proto.Float64(60 * 60 * 24 * 7),
	}

	return fw, nil
}

func stateFromMasters(masters []string) (*megos.State, error) {
	masterUrls := make([]*url.URL, 0)
	for _, master := range masters {
		masterUrl, _ := url.Parse(fmt.Sprintf("http://%s", master))
		masterUrls = append(masterUrls, masterUrl)
	}

	mesos := megos.NewClient(masterUrls, nil)
	return mesos.GetStateFromCluster()
}

func (s *Scheduler) Send(call *sched.Call) (*http.Response, error) {
	payload, err := proto.Marshal(call)
	if err != nil {
		return nil, err
	}
	return s.client.Send(payload)
}

func (s *Scheduler) AddEvent(eventType sched.Event_Type, e *sched.Event) {
	s.MesosEventChan <- &event.MesosEvent{EventType: eventType, Event: e}
}

func (s *Scheduler) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case call := <-s.MesosCallChan:
			logrus.WithFields(logrus.Fields{"sending-call": sched.Call_Type_name[int32(*call.Type)]}).Debugf("")
			resp, err := s.Send(call)
			if err != nil {
				logrus.Errorf("%s", err)
			}
			if resp.StatusCode != 202 {
				logrus.Infof("send response not 202 but %d", resp.StatusCode)
			}
		}
	}
}
