package scheduler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Dataman-Cloud/swan/client"
	"github.com/Dataman-Cloud/swan/health"
	"github.com/Dataman-Cloud/swan/mesosproto/mesos"
	sched "github.com/Dataman-Cloud/swan/mesosproto/sched"
	"github.com/Dataman-Cloud/swan/types"
	"github.com/Sirupsen/logrus"
	"github.com/gogo/protobuf/proto"
)

// Scheduler represents a Mesos scheduler
type Scheduler struct {
	master    string
	framework *mesos.FrameworkInfo
	registry  Registry

	client       *client.Client
	DoneChan     chan struct{}
	ReschedQueue chan types.ReschedulerMsg
	events       Events

	taskLaunched int

	// Status indicated scheduler's state is idle or busy.
	Status string

	ClusterId string

	HealthCheckManager *health.HealthCheckManager
}

// New returns a pointer to new Scheduler
func New(master string, fw *mesos.FrameworkInfo, registry Registry, clusterId string,
	health *health.HealthCheckManager, queue chan types.ReschedulerMsg) *Scheduler {
	return &Scheduler{
		master:    master,
		client:    client.New(master, "/api/v1/scheduler"),
		framework: fw,
		registry:  registry,
		DoneChan:  make(chan struct{}),
		events: Events{
			sched.Event_SUBSCRIBED: make(chan *sched.Event, 64),
			sched.Event_OFFERS:     make(chan *sched.Event, 64),
			sched.Event_RESCIND:    make(chan *sched.Event, 64),
			sched.Event_UPDATE:     make(chan *sched.Event, 64),
			sched.Event_MESSAGE:    make(chan *sched.Event, 64),
			sched.Event_FAILURE:    make(chan *sched.Event, 64),
			sched.Event_ERROR:      make(chan *sched.Event, 64),
			sched.Event_HEARTBEAT:  make(chan *sched.Event, 64),
		},
		Status:             "idle",
		ClusterId:          clusterId,
		HealthCheckManager: health,
		ReschedQueue:       queue,
	}
}

// start starts the scheduler and subscribes to event stream
// returns a channel to wait for completion.
func (s *Scheduler) Start() <-chan struct{} {
	if err := s.subscribe(); err != nil {
		logrus.Fatal(err)
	}
	return s.DoneChan
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
	logrus.Infof("Subscribe with mesos master %s", s.master)
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
		close(s.DoneChan)
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
			logrus.Infof("Subscription successful with frameworkId %s", sub.FrameworkId.GetValue())
			if registered, _ := s.registry.FrameworkIDHasRegistered(sub.FrameworkId.GetValue()); !registered {
				if err := s.registry.RegisterFrameworkID(sub.FrameworkId.GetValue()); err != nil {
					logrus.Errorf("Register framework id in consul failed: %s", err)
					return
				}
			}
			s.AddEvent(sched.Event_SUBSCRIBED, event)

			go func() {
				s.HealthCheckManager.Init()
				s.HealthCheckManager.Start()
			}()

			go func() {
				s.ReschedulerTask()
			}()
		case sched.Event_OFFERS:
			if s.Status == "idle" {
				// Refused all offers when scheduler is idle.
				for _, offer := range event.Offers.Offers {
					s.DeclineResource(offer.GetId().Value)
				}
			} else {
				// Accept all offers when scheduler is busy.
				s.AddEvent(sched.Event_OFFERS, event)
			}
		case sched.Event_RESCIND:
			logrus.Info("Received rescind offers")
			s.AddEvent(sched.Event_RESCIND, event)

		case sched.Event_UPDATE:
			status := event.GetUpdate().GetStatus()
			go func() {
				s.status(status)
			}()

			s.AddEvent(sched.Event_UPDATE, event)
		case sched.Event_MESSAGE:
			logrus.Info("Received message event")
			s.AddEvent(sched.Event_MESSAGE, event)

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

			s.AddEvent(sched.Event_FAILURE, event)
		case sched.Event_ERROR:
			err := event.GetError().GetMessage()
			logrus.Error(err)
			s.AddEvent(sched.Event_ERROR, event)

		case sched.Event_HEARTBEAT:
			s.AddEvent(sched.Event_HEARTBEAT, event)
		}

	}
}
