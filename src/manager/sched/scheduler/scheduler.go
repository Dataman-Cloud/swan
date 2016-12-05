package scheduler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Dataman-Cloud/swan/src/manager/sched/client"
	"github.com/Dataman-Cloud/swan/src/manager/store"
	"github.com/Dataman-Cloud/swan/src/mesosproto/mesos"
	sched "github.com/Dataman-Cloud/swan/src/mesosproto/sched"
	"github.com/Dataman-Cloud/swan/src/types"
	"github.com/Dataman-Cloud/swan/src/util"

	"github.com/Sirupsen/logrus"
	"github.com/andygrunwald/megos"
	"github.com/golang/protobuf/proto"
	"github.com/samuel/go-zookeeper/zk"
)

// Scheduler represents a Mesos scheduler
type Scheduler struct {
	master    string
	framework *mesos.FrameworkInfo
	store     store.Store

	client       *client.Client
	doneChan     chan struct{}
	ReschedQueue chan types.ReschedulerMsg
	events       Events

	TaskLaunched int

	// Status indicated scheduler's state is idle or busy.
	Status string

	ClusterId string
	config    util.Scheduler
}

// NewScheduler returns a pointer to new Scheduler
func NewScheduler(config util.Scheduler, store store.Store) *Scheduler {
	s := &Scheduler{
		config:   config,
		store:    store,
		doneChan: make(chan struct{}),
		events: Events{
			sched.Event_OFFERS: make(chan *sched.Event, 64),
		},
		Status: "idle",
	}

	// TODO, the follow statements are only used for passing test,
	// should resovled to make sure test pass
	s.client = client.New("foobar", "make test pass")
	return s
}

// start starts the scheduler and subscribes to event stream
func (s *Scheduler) Start() error {
	var err error
	s.framework, err = buildFrameworkInfo(s.config, s.store)
	masters := make([]string, 0)
	if strings.HasPrefix(s.config.MesosMasters, "zk://") {
		master, err := getMasterFromZK(s.config.MesosMasters)
		if err != nil {
			logrus.Error(err)
			return err
		}
		masters = append(masters, master)
	} else {
		masters = strings.Split(s.config.MesosMasters, ",")
	}

	state, err := stateFromMasters(masters)
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
	s.client = client.New(state.Leader, "/api/v1/scheduler")

	if err := s.subscribe(); err != nil {
		logrus.Error(err)
		return err
	}

	return nil
}

// build frameworkInfo on initial start
func buildFrameworkInfo(config util.Scheduler, store store.Store) (*mesos.FrameworkInfo, error) {
	fw := &mesos.FrameworkInfo{
		User:            proto.String(config.MesosFrameworkUser),
		Name:            proto.String("swan"),
		Hostname:        proto.String(config.Hostname),
		FailoverTimeout: proto.Float64(60 * 60 * 24 * 7),
	}

	frameworkId, err := store.FetchFrameworkID()
	if err != nil {
		logrus.Errorf("Fetch framework id failed: %s", err)
		return nil, err
	}

	if frameworkId != "" {
		fw.Id = &mesos.FrameworkID{
			Value: proto.String(frameworkId),
		}
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

func getMasterFromZK(zks string) (string, error) {
	masterInfo := new(mesos.MasterInfo)

	url, err := url.Parse(zks)
	if err != nil {
		return "", err
	}

	conn, _, err := zk.Connect(strings.Split(url.Host, ","), time.Second)
	defer conn.Close()
	if err != nil {
		return "", fmt.Errorf("Couldn't connect to zookeeper:%s", err.Error())
	}

	// find mesos master
	children, _, err := conn.Children(url.Path)
	if err != nil {
		return "", fmt.Errorf("Couldn't connect to zookeeper:%s", err.Error())
	}

	for _, name := range children {
		data, _, err := conn.Get(url.Path + "/" + name)
		if err != nil {
			return "", fmt.Errorf("%s", err.Error())
		}
		if err := proto.Unmarshal(data, masterInfo); err != nil {
			break
		}
	}

	return fmt.Sprintf("%s:%d", masterInfo.GetHostname(), masterInfo.GetPort()), nil
}

func (s *Scheduler) Stop() {
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

	// http might now be the default transport in future release
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Subscribe with unexpected response status: %d", resp.StatusCode)
	}

	logrus.Info(s.client.StreamID)
	go s.handleEvents(resp)

	return nil
}

// main loop of a scheduler module
func (s *Scheduler) handleEvents(resp *http.Response) {
	defer func() {
		resp.Body.Close()
		close(s.doneChan)
		for _, event := range s.events {
			close(event)
		}
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
			sub := event.GetSubscribed()
			logrus.Infof("Subscription successful with frameworkId %s", sub.FrameworkId.GetValue())
			if registered, _ := s.store.HasFrameworkID(); !registered {
				if err := s.store.SaveFrameworkID(sub.FrameworkId.GetValue()); err != nil {
					logrus.Errorf("Register framework id in db failed: %s", err)
					return
				}
			}

			if s.framework.Id == nil {
				s.framework.Id = sub.FrameworkId
			}

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

		case sched.Event_UPDATE:
			status := event.GetUpdate().GetStatus()
			// why status & AddEvent together, as initial thought would be
			// status should be notified on the other side of events
			go func() {
				s.status(status)
			}()

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
			logrus.Debug("Received heartbeat msg")
		}

	}
}
