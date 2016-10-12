package scheduler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Dataman-Cloud/swan/client"
	"github.com/Dataman-Cloud/swan/mesosproto/mesos"
	sched "github.com/Dataman-Cloud/swan/mesosproto/sched"
	"github.com/Sirupsen/logrus"
	"github.com/gogo/protobuf/proto"
	//consul "github.com/hashicorp/consul/api"
)

// Scheduler represents a Mesos scheduler
type Scheduler struct {
	master    string
	framework *mesos.FrameworkInfo
	registry  Registry

	client   *client.Client
	doneChan chan struct{}
	events   Events
}

// New returns a pointer to new Scheduler
//func New(master string, fw *mesos.FrameworkInfo, registry Registry, consul *consul.Client) *Scheduler {
func New(master string, fw *mesos.FrameworkInfo, registry Registry) *Scheduler {
	return &Scheduler{
		master:    master,
		client:    client.New(master, "/api/v1/scheduler"),
		framework: fw,
		registry:  registry,
		doneChan:  make(chan struct{}),
		events: Events{
			sched.Event_OFFERS: make(chan *sched.Event, 64),
		},
	}
}

// start starts the scheduler and subscribes to event stream
// returns a channel to wait for completion.
func (s *Scheduler) Start() <-chan struct{} {
	if err := s.subscribe(); err != nil {
		logrus.Fatal(err)
	}
	return s.doneChan
}

func (s *Scheduler) stop() {
	for _, event := range s.events {
		close(event)
	}
}

func (s *Scheduler) send(call *sched.Call) (*http.Response, error) {
	payload, err := proto.Marshal(call)
	if err != nil {
		return nil, err
	}
	return s.client.Send(payload)
}

func (s *Scheduler) Registry() Registry {
	return s.registry
}

// Subscribe subscribes the scheduler to the Mesos cluster.
// It keeps the http connection opens with the Master to stream
// subsequent events.
func (s *Scheduler) subscribe() error {
	logrus.WithFields(logrus.Fields{"master": s.master}).Info("Subscribe with mesos")
	call := &sched.Call{
		Type: sched.Call_SUBSCRIBE.Enum(),
		Subscribe: &sched.Call_Subscribe{
			FrameworkInfo: s.framework,
		},
	}

	if s.framework.Id != nil {
		call.FrameworkId = &mesos.FrameworkID{
			Value: proto.String(s.framework.Id.GetValue()),
		}
	}

	resp, err := s.send(call)
	if err != nil {
		return err
	}
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
		close(s.doneChan)
		for _, event := range s.events {
			close(event)
		}
	}()
	dec := json.NewDecoder(resp.Body)
	for {
		event := new(sched.Event)
		if err := dec.Decode(event); err != nil {
			if err == io.EOF {
				return
			}
			continue
		}

		switch event.GetType() {
		case sched.Event_SUBSCRIBED:
			sub := event.GetSubscribed()
			//s.framework.Id = sub.FrameworkId
			logrus.WithFields(logrus.Fields{"FrameworkId": sub.FrameworkId.GetValue()}).Info("Subscription successful.")
			if registered, _ := s.registry.FrameworkIDHasRegistered(sub.FrameworkId.GetValue()); !registered {
				if err := s.registry.RegisterFrameworkID(sub.FrameworkId.GetValue()); err != nil {
					logrus.Errorf("Register framework id in consul failed: %s", err)
					return
				}
			}
		case sched.Event_OFFERS:
			s.AddEvent(sched.Event_OFFERS, event)

		case sched.Event_RESCIND:
			logrus.Info("Received rescind offers")

		case sched.Event_UPDATE:
			status := event.GetUpdate().GetStatus()
			go s.status(status)

		case sched.Event_MESSAGE:
			logrus.Info("Received message event")

		case sched.Event_FAILURE:
			logrus.Error("Received failure event")
			fail := event.GetFailure()
			if fail.ExecutorId != nil {
				logrus.Error(
					"Executor ", fail.ExecutorId.GetValue(), " terminated ",
					" with status ", fail.GetStatus(),
					" on agent ", fail.GetAgentId().GetValue(),
				)
			} else {
				if fail.GetAgentId() != nil {
					logrus.Error("Agent ", fail.GetAgentId().GetValue(), " failed ")
				}
			}

		case sched.Event_ERROR:
			err := event.GetError().GetMessage()
			logrus.Error(err)

		case sched.Event_HEARTBEAT:

		}

	}
}
