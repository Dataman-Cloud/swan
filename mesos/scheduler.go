package mesos

import (
	"encoding/json"
	"errors"
	"fmt"
	//"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"

	"github.com/Dataman-Cloud/swan/mesosproto"
	"github.com/Dataman-Cloud/swan/store"
	"github.com/Dataman-Cloud/swan/types"
)

const (
	reconnectDuration = time.Duration(20 * time.Second)
	resourceTimeout   = time.Duration(360000 * time.Second)
	creationTimeout   = time.Duration(360000 * time.Second)
	deleteTimeout     = time.Duration(360000 * time.Second)
	reconcileInterval = time.Duration(24 * time.Hour)
)

var (
	errResourceNotEnough = errors.New("resource not enough")
	errCreationTimeout   = errors.New("task create timeout")
	errDeletingTimeout   = errors.New("task delete timeout")
)

type SchedulerConfig struct {
	ZKHost []string
	ZKPath string

	ReconciliationInterval  float64
	ReconciliationStep      int64
	ReconciliationStepDelay float64
}

// Scheduler represents a client interacting with mesos master via x-protobuf
type Scheduler struct {
	http      *httpClient
	cfg       *SchedulerConfig
	framework *mesosproto.FrameworkInfo

	eventCh chan *mesosproto.Event // mesos events
	errCh   chan error             // subscriber's error events
	quit    chan struct{}

	//endPoint string // eg: http://master/api/v1/scheduler
	leader  string // mesos leader address
	cluster string // name of mesos cluster

	db store.Store

	sync.RWMutex
	agents map[string]*Agent

	handlers map[mesosproto.Event_Type]eventHandler
	tasks    map[string]*Task

	offerTimeout time.Duration

	heartbeatTimeout time.Duration
	watcher          *time.Timer
	reconcileTimer   *time.Ticker

	strategy Strategy
	filters  []Filter

	eventmgr *eventManager

	launch sync.Mutex
}

// NewScheduler...
func NewScheduler(cfg *SchedulerConfig, db store.Store, strategy Strategy, mgr *eventManager) (*Scheduler, error) {
	s := &Scheduler{
		cfg:          cfg,
		framework:    defaultFramework(),
		errCh:        make(chan error, 1),
		quit:         make(chan struct{}),
		agents:       make(map[string]*Agent),
		tasks:        make(map[string]*Task),
		offerTimeout: time.Duration(10 * time.Second),
		db:           db,
		strategy:     strategy,
		filters:      make([]Filter, 0),
		eventmgr:     mgr,
	}

	if err := s.init(); err != nil {
		return nil, err
	}

	return s, nil
}

// init setup mesos sched api endpoint & cluster name
func (s *Scheduler) init() error {
	state, err := s.MesosState()
	if err != nil {
		return err
	}

	l := state.Leader
	if l == "" {
		return fmt.Errorf("no mesos leader found.")
	}

	s.http = NewHTTPClient(l)
	s.leader = state.Leader

	s.cluster = state.Cluster
	if s.cluster == "" {
		s.cluster = "unnamed" // set default cluster name
	}

	s.handlers = map[mesosproto.Event_Type]eventHandler{
		mesosproto.Event_SUBSCRIBED: s.subscribedHandler,
		mesosproto.Event_OFFERS:     s.offersHandler,
		mesosproto.Event_RESCIND:    s.rescindedHandler,
		mesosproto.Event_UPDATE:     s.updateHandler,
		mesosproto.Event_HEARTBEAT:  s.heartbeatHandler,
		mesosproto.Event_ERROR:      s.errHandler,
		mesosproto.Event_FAILURE:    s.failureHandler,
		mesosproto.Event_MESSAGE:    s.messageHandler,
	}

	if id := s.db.GetFrameworkId(); id != "" {
		s.framework.Id = &mesosproto.FrameworkID{
			Value: proto.String(id),
		}
	}

	//s.reconcileTimer = time.NewTicker(reconcileInterval)
	//// TOOD(nmg): stop timer.
	//go func() {
	//	for range s.reconcileTimer.C {
	//		s.reconcile()
	//	}
	//}()

	return nil
}

func (s *Scheduler) InitFilters(filters []Filter) {
	s.filters = filters
}

// Cluster return current mesos cluster's name
func (s *Scheduler) ClusterName() string {
	return s.cluster
}

func (s *Scheduler) FrameworkId() *mesosproto.FrameworkID {
	return s.framework.Id
}

// Send send mesos request against the mesos master's scheduler api endpoint.
// NOTE it's the caller's responsibility to deal with the Send() error
func (s *Scheduler) Send(call *mesosproto.Call) (*http.Response, error) {
	payload, err := proto.Marshal(call)
	if err != nil {
		return nil, err
	}

	return s.http.send(payload)
}

func (s *Scheduler) connect() (*http.Response, error) {
	call := &mesosproto.Call{
		Type: mesosproto.Call_SUBSCRIBE.Enum(),
		Subscribe: &mesosproto.Call_Subscribe{
			FrameworkInfo: s.framework,
		},
	}

	if s.framework.Id != nil {
		call.FrameworkId = &mesosproto.FrameworkID{
			Value: proto.String(s.framework.Id.GetValue()),
		}
	}

	resp, err := s.Send(call)
	if err != nil {
		return nil, fmt.Errorf("subscribe to mesos leader [%s] error [%v]", s.leader, err)
	}

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("subscribe with unexpected response [%d] - [%s]", code, string(bs))
	}

	return resp, nil
}

// Subscribe ...
func (s *Scheduler) Subscribe() error {
	log.Infof("Subscribing to mesos leader: %s", s.leader)

	resp, err := s.connect()
	if err != nil {
		return err
	}

	go s.watchEvents(resp)

	return nil
}

func (s *Scheduler) Unsubscribe() error {
	log.Println("Unscribing from mesos leader:", s.leader)
	close(s.quit)
	return nil
}

func (s *Scheduler) reconnect() {
	s.http.Reset()

	var (
		resp *http.Response
		err  error
	)

	for {
		log.Printf("Reconnecting to mesos leader: %s", s.leader)

		resp, err = s.connect()
		if err == nil {
			go s.watchEvents(resp)
			return
		}

		time.Sleep(2 * time.Second)
	}
}

func (s *Scheduler) watchEvents(resp *http.Response) {
	defer func() {
		resp.Body.Close()
	}()

	r := NewReader(resp.Body)
	dec := json.NewDecoder(r)

	for {
		select {
		case <-s.quit:
			return
		default:
			ev := new(mesosproto.Event)
			if err := dec.Decode(ev); err != nil {
				log.Error("mesos events subscriber decode events error:", err)

				// stop tasks reconciliation befor re-connect.
				s.stopReconcile()

				go s.reconnect()

				return
			}

			s.handleEvent(ev)
		}
	}
}

func (s *Scheduler) handleEvent(ev *mesosproto.Event) {
	var (
		typ     = ev.GetType()
		handler = s.handlers[typ]
	)
	if handler == nil {
		log.Error("without any proper event handler for mesos event:", typ)
		return
	}

	// TODO panic protection on each event handling ?
	handler(ev)
}

func (s *Scheduler) addOffer(offer *mesosproto.Offer) {
	a, ok := s.agents[offer.AgentId.GetValue()]
	if !ok {
		return
	}

	a.addOffer(offer)
}

func (s *Scheduler) removeOffer(offer *mesosproto.Offer) bool {
	log.Debugf("Removing offer %s", offer.Id.GetValue())

	a := s.getAgent(offer.AgentId.GetValue())
	if a == nil {
		return false
	}

	found := a.removeOffer(offer.Id.GetValue())
	if a.empty() {
		s.removeAgent(offer.AgentId.GetValue())
	}

	return found
}

func (s *Scheduler) declineOffer(offer *mesosproto.Offer) error {
	call := &mesosproto.Call{
		FrameworkId: s.FrameworkId(),
		Type:        mesosproto.Call_DECLINE.Enum(),
		Decline: &mesosproto.Call_Decline{
			OfferIds: []*mesosproto.OfferID{
				{
					Value: offer.GetId().Value,
				},
			},
			Filters: &mesosproto.Filters{
				RefuseSeconds: proto.Float64(1),
			},
		},
	}

	resp, err := s.Send(call)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("%d", resp.StatusCode)
	}

	return nil

}

func (s *Scheduler) rescindOffer(offerId string) {
}

func (s *Scheduler) addAgent(agent *Agent) {
	s.Lock()
	defer s.Unlock()

	s.agents[agent.id] = agent
}

func (s *Scheduler) getAgent(agentId string) *Agent {
	s.RLock()
	defer s.RUnlock()

	a, ok := s.agents[agentId]
	if !ok {
		return nil
	}

	return a
}

func (s *Scheduler) removeAgent(agentId string) {
	s.Lock()
	defer s.Unlock()

	delete(s.agents, agentId)
}

func (s *Scheduler) getAgents() []*Agent {
	s.RLock()
	defer s.RUnlock()

	agents := make([]*Agent, 0)
	for _, agent := range s.agents {
		agents = append(agents, agent)
	}

	return agents
}

func (s *Scheduler) addTask(task *Task) {
	s.Lock()
	defer s.Unlock()

	s.tasks[task.TaskId.GetValue()] = task
}

func (s *Scheduler) removeTask(taskID string) bool {
	s.Lock()
	defer s.Unlock()
	found := false
	_, found = s.tasks[taskID]
	if found {
		delete(s.tasks, taskID)
	}
	return found
}

func (s *Scheduler) LaunchTask(t *Task) error {
	log.Printf("launching task %s", t.ID())

	s.launch.Lock()

	s.addTask(t)
	defer func() {
		s.removeTask(t.TaskId.GetValue())
	}()

	filtered, err := s.applyFilters(t.cfg)
	if err != nil {
		s.launch.Unlock()

		return err
	}

	candidates := s.strategy.RankAndSort(filtered)

	chosen := candidates[0]

	var offer *mesosproto.Offer
	for _, ofr := range chosen.getOffers() {
		offer = ofr
	}

	t.Build(offer)

	appId := strings.SplitN(t.GetName(), ".", 2)[1]

	task, err := s.db.GetTask(appId, t.GetTaskId().GetValue())
	if err != nil {
		s.launch.Unlock()
		return fmt.Errorf("find task from zk got error: %v", err)
	}

	task.AgentId = t.AgentId.GetValue()
	task.IP = t.cfg.IP

	if t.cfg.Network == "host" || t.cfg.Network == "bridge" {
		task.IP = offer.GetHostname()
	}

	task.Port = t.cfg.Port

	if err := s.db.UpdateTask(appId, task); err != nil {
		s.launch.Unlock()
		return fmt.Errorf("update task status error: %v", err)
	}

	call := &mesosproto.Call{
		FrameworkId: s.FrameworkId(),
		Type:        mesosproto.Call_ACCEPT.Enum(),
		Accept: &mesosproto.Call_Accept{
			OfferIds: []*mesosproto.OfferID{
				offer.GetId(),
			},
			Operations: []*mesosproto.Offer_Operation{
				&mesosproto.Offer_Operation{
					Type: mesosproto.Offer_Operation_LAUNCH.Enum(),
					Launch: &mesosproto.Offer_Operation_Launch{
						TaskInfos: []*mesosproto.TaskInfo{&t.TaskInfo},
					},
				},
			},
			Filters: &mesosproto.Filters{RefuseSeconds: proto.Float64(1)},
		},
	}

	// send call
	resp, err := s.Send(call)
	if err != nil {
		s.launch.Unlock()
		return err
	}

	if code := resp.StatusCode; code != http.StatusAccepted {
		s.launch.Unlock()
		return fmt.Errorf("launch call send but the status code not 202 got %d", code)
	}

	s.removeOffer(offer)

	s.launch.Unlock()

	wait := time.NewTicker(creationTimeout)
	defer wait.Stop()

	first := true

	// waitting for task's update events here until task finished or met error.
	for {
		select {
		case status := <-t.GetStatus():
			if t.IsDone(status) {
				return t.DetectError(status)
			}
		case <-wait.C:
			if first {
				m := make(map[*mesosproto.TaskID]*mesosproto.AgentID)
				m[t.TaskId] = t.AgentId

				if err := s.reconcileTasks(m); err != nil {
					log.Errorf("reconcile tasks got error: %v", err)
					return errCreationTimeout
				}

				first = false

				continue
			}

			return errCreationTimeout
		}
	}
}

func (s *Scheduler) KillTask(taskId, agentId string) error {
	log.Info("killing task ", taskId)

	defer func() {
		s.removeTask(taskId)
	}()

	t := NewTask(nil, taskId, taskId)

	s.addTask(t)

	call := &mesosproto.Call{
		FrameworkId: s.FrameworkId(),
		Type:        mesosproto.Call_KILL.Enum(),
		Kill: &mesosproto.Call_Kill{
			TaskId: &mesosproto.TaskID{
				Value: proto.String(taskId),
			},
			AgentId: &mesosproto.AgentID{
				Value: proto.String(agentId),
			},
		},
	}

	// send call
	resp, err := s.Send(call)
	if err != nil {
		return err
	}

	if code := resp.StatusCode; code != http.StatusAccepted {
		return fmt.Errorf("kill call send but the status code not 202 got %d", code)
	}

	timeout := time.NewTicker(deleteTimeout)
	defer timeout.Stop()

	first := true
	// waitting for task's update events here until task finished or met error.
	for {
		select {
		case status := <-t.GetStatus():
			if t.IsTerminated(status) {
				return nil
			}
		case <-timeout.C:
			if first {
				m := make(map[*mesosproto.TaskID]*mesosproto.AgentID)
				m[&mesosproto.TaskID{Value: proto.String(taskId)}] = &mesosproto.AgentID{Value: proto.String(agentId)}

				if err := s.reconcileTasks(m); err != nil {
					log.Errorf("reconcile tasks got error: %v", err)
					return errDeletingTimeout
				}

				first = false

				continue
			}

			return errDeletingTimeout
		}
	}
}

func (s *Scheduler) applyFilters(config *types.TaskConfig) ([]*Agent, error) {
	filtered := make([]*Agent, 0)

	timeout := time.After(resourceTimeout)
	for {
		select {
		case <-timeout:
			return nil, errResourceNotEnough
		default:
			filtered = ApplyFilters(s.filters, config, s.getAgents())
			if len(filtered) > 0 {
				return filtered, nil
			}
			time.Sleep(1 * time.Second)
		}
	}
}

func (s *Scheduler) reconcileTasks(tasks map[*mesosproto.TaskID]*mesosproto.AgentID) error {
	call := &mesosproto.Call{
		FrameworkId: s.FrameworkId(),
		Type:        mesosproto.Call_RECONCILE.Enum(),
		Reconcile: &mesosproto.Call_Reconcile{
			Tasks: []*mesosproto.Call_Reconcile_Task{},
		},
	}

	for t, a := range tasks {
		call.Reconcile.Tasks = append(call.Reconcile.Tasks, &mesosproto.Call_Reconcile_Task{
			TaskId:  t,
			AgentId: a,
		})
	}

	resp, err := s.Send(call)
	if err != nil {
		return err
	}

	if code := resp.StatusCode; code != http.StatusAccepted {
		return fmt.Errorf("send reconcile call got %d not 202", code)
	}

	return nil
}

func (s *Scheduler) DetectError(status *mesosproto.TaskStatus) error {
	var (
		state = status.GetState()
		//data  = status.GetData() // docker container inspect result
	)

	switch state {
	case mesosproto.TaskState_TASK_FAILED,
		mesosproto.TaskState_TASK_ERROR,
		mesosproto.TaskState_TASK_LOST,
		mesosproto.TaskState_TASK_DROPPED,
		mesosproto.TaskState_TASK_UNREACHABLE,
		mesosproto.TaskState_TASK_GONE,
		mesosproto.TaskState_TASK_GONE_BY_OPERATOR,
		mesosproto.TaskState_TASK_UNKNOWN:
		bs, _ := json.Marshal(map[string]interface{}{
			"state":   state.String(),
			"message": status.GetMessage(),
			"source":  status.GetSource().String(),
			"reason":  status.GetReason().String(),
			"healthy": status.GetHealthy(),
		})
		return errors.New(string(bs))
	}

	return nil
}

func (s *Scheduler) AckUpdateEvent(status *mesosproto.TaskStatus) error {
	if status.GetUuid() != nil {
		call := &mesosproto.Call{
			FrameworkId: s.FrameworkId(),
			Type:        mesosproto.Call_ACKNOWLEDGE.Enum(),
			Acknowledge: &mesosproto.Call_Acknowledge{
				AgentId: status.GetAgentId(),
				TaskId:  status.GetTaskId(),
				Uuid:    status.GetUuid(),
			},
		}
		resp, err := s.Send(call)
		if err != nil {
			return err
		}

		if code := resp.StatusCode; code != http.StatusAccepted {
			return fmt.Errorf("send ack got %d not 202", code)
		}
	}

	return nil
}

func (s *Scheduler) SubscribeEvent(w http.ResponseWriter, remote string) error {
	if s.eventmgr.Full() {
		return fmt.Errorf("%s", "too many event clients")
	}

	s.eventmgr.subscribe(remote, w)
	s.eventmgr.wait(remote)

	return nil
}

func (s *Scheduler) watchConn(interval float64) {
	du := interval + 5
	s.heartbeatTimeout = time.Duration(du) * time.Second
	//s.watcher = time.AfterFunc(s.heartbeatTimeout, s.connect)
}

func (s *Scheduler) reconcile() {
	log.Println("Start task reconciliation with the Mesos master")

	apps, err := s.db.ListApps()
	if err != nil {
		log.Errorf("List app got error for task reconcile. %v", err)
		return
	}

	m := make(map[*mesosproto.TaskID]*mesosproto.AgentID)
	for _, app := range apps {
		for _, task := range app.Tasks {
			m[&mesosproto.TaskID{Value: proto.String(task.ID)}] = &mesosproto.AgentID{Value: proto.String(task.AgentId)}
		}
	}
	if err := s.reconcileTasks(m); err != nil {
		log.Errorf("reconcile tasks got error: %v", err)
	}
}

func (s *Scheduler) startReconcile() {
	var (
		interval = time.Duration(s.cfg.ReconciliationInterval) * time.Second
		step     = int(s.cfg.ReconciliationStep)
		delay    = time.Duration(s.cfg.ReconciliationStepDelay) * time.Second
	)

	s.reconcileTimer = time.NewTicker(interval)
	go func() {
		for range s.reconcileTimer.C {
			apps, err := s.db.ListApps()
			if err != nil {
				log.Errorf("List app got error for task reconcile. %v", err)
				return
			}

			ch := make(chan *types.Task, step)
			go func() {
				for _, app := range apps {
					for _, task := range app.Tasks {
						ch <- task
					}
				}
			}()

			m := make(map[*mesosproto.TaskID]*mesosproto.AgentID)
			for _, app := range apps {
				for _, task := range app.Tasks {
					m[&mesosproto.TaskID{Value: proto.String(task.ID)}] = &mesosproto.AgentID{Value: proto.String(task.AgentId)}

					if len(m) >= step {
						if err := s.reconcileTasks(m); err != nil {
							log.Errorf("reconcile tasks got error: %v", err)
						}

						m = make(map[*mesosproto.TaskID]*mesosproto.AgentID)

						time.Sleep(delay)
					}
				}
			}
		}
	}()
}

func (s *Scheduler) stopReconcile() {
	s.reconcileTimer.Stop()
}
