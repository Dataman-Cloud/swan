package mesos

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"

	"github.com/Dataman-Cloud/swan/mesosproto"
	"github.com/Dataman-Cloud/swan/mole"
	"github.com/Dataman-Cloud/swan/store"
	"github.com/Dataman-Cloud/swan/types"
)

const (
	reconnectDuration = time.Duration(20 * time.Second)
	resourceTimeout   = time.Duration(360000 * time.Second)
	creationTimeout   = time.Duration(360000 * time.Second)
	deleteTimeout     = time.Duration(360000 * time.Second)
	reconcileInterval = time.Duration(24 * time.Hour)

	statusConnecting = "connecting"
	statusConnected  = "connected"
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

	HeartbeatTimeout float64
	MaxTasksPerOffer int
}

// Scheduler represents a client interacting with mesos master via x-protobuf
type Scheduler struct {
	http      *httpClient // mesos scheduler http client
	cfg       *SchedulerConfig
	framework *mesosproto.FrameworkInfo

	quit chan struct{}

	leader  string
	cluster string // name of mesos cluster

	db store.Store

	handlers map[mesosproto.Event_Type]eventHandler

	sync.RWMutex                   // protect followings two
	agents       map[string]*Agent // holding offers (agents)
	tasks        map[string]*Task

	offerTimeout time.Duration

	reconcileTimer *time.Ticker

	strategy Strategy
	filters  []Filter

	eventmgr *eventManager

	clusterMaster *mole.Master

	sem chan struct{}
}

// NewScheduler...
func NewScheduler(cfg *SchedulerConfig, db store.Store, strategy Strategy, clusterMaster *mole.Master) (*Scheduler, error) {
	s := &Scheduler{
		cfg:           cfg,
		framework:     defaultFramework(),
		quit:          make(chan struct{}),
		agents:        make(map[string]*Agent),
		tasks:         make(map[string]*Task),
		offerTimeout:  time.Duration(10 * time.Second),
		db:            db,
		strategy:      strategy,
		filters:       make([]Filter, 0),
		eventmgr:      NewEventManager(),
		clusterMaster: clusterMaster,
		sem:           make(chan struct{}, 1),
	}

	if err := s.init(); err != nil {
		return nil, err
	}

	return s, nil
}

// init setup mesos sched api endpoint & cluster name
func (s *Scheduler) init() error {
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

	id, mtime := s.db.GetFrameworkId()
	if id != "" {
		if time.Now().Unix()-(mtime/1000) >= DefaultFrameworkFailoverTimeout {
			log.Warnln("framework failover time exceed")
			return nil
		}
		// attach framework with given id to subscribe with mesos
		s.framework.Id = &mesosproto.FrameworkID{
			Value: proto.String(id),
		}
	}

	return nil
}

func (s *Scheduler) detectMesosState() (string, string, error) {
	state, err := s.MesosState()
	if err != nil {
		return "", "", err
	}

	return state.Leader, state.Cluster, nil
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
	l, c, err := s.detectMesosState()
	if err != nil {
		return nil, err
	}

	if l == "" {
		return nil, fmt.Errorf("no mesos leader found.")
	}
	s.leader = l

	if c == "" {
		c = "unnamed"
	}
	s.cluster = c

	s.http = NewHTTPClient(l)

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

	log.Printf("Subscribing to mesos leader %s", l)
	resp, err := s.Send(call)
	if err != nil {
		return nil, fmt.Errorf("subscribe to mesos leader [%s] error [%v]", l, err)
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
	resp, err := s.connect()
	if err != nil {
		return err
	}

	go s.watchEvents(resp)

	return nil
}

func (s *Scheduler) Unsubscribe() error {
	log.Println("Unscribing from mesos leader %s", s.leader)
	return nil
}

func (s *Scheduler) reconnect() {
	// Empty Mesos-Stream-Id for new connect.
	s.http.Reset()

	var (
		err  error
		resp *http.Response
	)

	for {
		resp, err = s.connect()
		if err == nil {
			go s.watchEvents(resp)

			return
		}

		time.Sleep(2 * time.Second)
	}
}

func (s *Scheduler) watchEvents(resp *http.Response) {
	var (
		ev  *mesosproto.Event
		err error
	)

	dec := json.NewDecoder(NewReader(resp.Body))

	for {
		if err = dec.Decode(&ev); err != nil {
			log.Errorf("mesos events subscriber decode events error: %v", err)
			resp.Body.Close()
			go s.reconnect()
			return
		}

		s.handleEvent(ev)
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

	if typ == mesosproto.Event_UPDATE {
		var (
			status = ev.GetUpdate().GetStatus()
			taskId = status.TaskId.GetValue()
			state  = status.GetState()
		)

		// ack firstly
		go func() {
			if err := s.AckUpdateEvent(status); err != nil {
				log.Errorf("send status update %s for task %s error: %v", status.GetState(), taskId, err)
			}
		}()

		// emit event status to ongoing task
		log.Debugf("Finding task %s to send status %s", taskId, state.String())
		if task := s.getTask(taskId); task != nil {
			log.Debugf("Sending status %s to task %s", state.String(), taskId)
			task.SendStatus(status)
		}

		handler(ev)

		return
	}

	go handler(ev)
}

func (s *Scheduler) addOffer(offer *mesosproto.Offer) {
	a := s.getAgent(offer.AgentId.GetValue())
	if a == nil {
		return
	}

	f := newOffer(offer)

	log.Debugf("Received offer %s with resource cpus:[%.2f] mem:[%.2fG] disk:[%.2fG] ports:%v from agent %s",
		f.GetId(), f.GetCpus(), f.GetMem()/1024, f.GetDisk()/1024, f.GetPortRange(), f.GetHostname())

	a.addOffer(f)
	time.AfterFunc(time.Second*5, func() { // release the offer later
		if s.removeOffer(f) {
			s.declineOffers([]*Offer{f})
		}
	})
}

func (s *Scheduler) removeOffer(offer *Offer) bool {
	log.Debugln("Removing offer ", offer.GetId())

	a := s.getAgent(offer.GetAgentId())
	if a == nil {
		return false
	}

	found := a.removeOffer(offer.GetId())
	if a.empty() {
		s.removeAgent(offer.GetAgentId())
	}

	return found
}

func (s *Scheduler) declineOffers(offers []*Offer) error {
	call := &mesosproto.Call{
		FrameworkId: s.FrameworkId(),
		Type:        mesosproto.Call_DECLINE.Enum(),
		Decline: &mesosproto.Call_Decline{
			OfferIds: []*mesosproto.OfferID{},
			Filters: &mesosproto.Filters{
				RefuseSeconds: proto.Float64(1),
			},
		},
	}

	for _, offer := range offers {
		call.Decline.OfferIds = append(call.Decline.OfferIds, &mesosproto.OfferID{
			Value: proto.String(offer.GetId()),
		})

		log.Debugf("Prepare to decline offer %s", offer.GetId())
	}

	log.Debugf("Decline %d offer(s)", len(offers))

	resp, err := s.Send(call)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("%d", resp.StatusCode)
	}

	return nil

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

func (s *Scheduler) addTask(t *Task) {
	s.Lock()
	defer s.Unlock()

	s.tasks[t.TaskId.GetValue()] = t
}

func (s *Scheduler) getTask(taskId string) *Task {
	s.RLock()
	defer s.RUnlock()

	task, ok := s.tasks[taskId]
	if !ok {
		return nil
	}

	return task
}

func (s *Scheduler) removeTask(taskId string) {
	s.Lock()
	defer s.Unlock()

	_, ok := s.tasks[taskId]
	if !ok {
		return
	}

	delete(s.tasks, taskId)
}

func (s *Scheduler) KillTasks(tasks []*types.Task) map[string]error {
	var (
		wg   sync.WaitGroup
		p    sync.Mutex
		errs = map[string]error{}
	)

	for _, task := range tasks {
		wg.Add(1)
		go func(task *types.Task, errs map[string]error) {
			defer wg.Done()

			var (
				taskId  = task.ID
				agentId = task.AgentId
			)

			if agentId == "" {
				log.Debugf("agentId of task %s is empty, ignore", taskId)
				p.Lock()
				errs[taskId] = nil
				p.Unlock()
				return
			}

			log.Debugf("Killing task %s with agentId %s", taskId, agentId)

			t := NewTask(nil, taskId, taskId)

			log.Debugf("Adding task %s", taskId)
			s.addTask(t)
			defer s.removeTask(taskId)

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
				p.Lock()
				errs[taskId] = err
				p.Unlock()
				return
			}

			if code := resp.StatusCode; code != http.StatusAccepted {
				p.Lock()
				errs[taskId] = fmt.Errorf("kill call send but the status code not 202 got %d", code)
				p.Unlock()
				return
			}

			log.Debugf("Waiting for task %s to be killed by mesos", taskId)
			for status := range t.GetStatus() {
				log.Debugf("Receiving status %s for task %s", status.GetState().String(), taskId)
				if t.IsKilled(status) {
					log.Debugf("Task %s killed", taskId)
					p.Lock()
					errs[taskId] = nil
					p.Unlock()
					return
				}
			}
		}(task, errs)
	}

	wg.Wait()

	return errs

}

func (s *Scheduler) KillTask(taskId, agentId string) error {
	if agentId == "" {
		log.Debugf("agentId of task %s is empty, ignore", taskId)
		return nil
	}

	log.Debugf("Killing task %s with agentId %s", taskId, agentId)

	t := NewTask(nil, taskId, taskId)

	log.Debugf("Adding task %s", taskId)
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

	log.Debugf("Waiting for task %s to be killed by mesos", taskId)
	for status := range t.GetStatus() {
		log.Debugf("Receiving status %s for task %s", status.GetState().String(), taskId)
		if t.IsKilled(status) {
			log.Debugf("Task %s killed", taskId)
			s.removeTask(taskId)
			break
		}
	}

	return nil
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

func (s *Scheduler) reconcile() {
	log.Println("Start task reconciliation with the Mesos master")

	apps, err := s.db.ListApps()
	if err != nil {
		log.Errorf("List app got error for task reconcile. %v", err)
		return
	}

	m := make(map[*mesosproto.TaskID]*mesosproto.AgentID)
	for _, app := range apps {
		tasks, err := s.db.ListTasks(app.ID)
		if err != nil {
			log.Errorf("List tasks got error: %v", err)
			continue
		}
		for _, task := range tasks {
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

			tasks := make([]*types.Task, 0)

			for _, app := range apps {
				tasks, err := s.db.ListTasks(app.ID)
				if err != nil {
					log.Errorf("List tasks got error: %v", err)
					continue
				}

				for _, task := range tasks {
					tasks = append(tasks, task)
				}
			}

			var (
				total = len(tasks)
				send  = 0
			)

			m := make(map[*mesosproto.TaskID]*mesosproto.AgentID)

			for _, task := range tasks {
				m[&mesosproto.TaskID{Value: proto.String(task.ID)}] = &mesosproto.AgentID{Value: proto.String(task.AgentId)}

				if len(m) >= step || (len(m)+send) >= total {
					if err := s.reconcileTasks(m); err != nil {
						log.Errorf("reconcile tasks got error: %v", err)
					}

					send += len(m)

					m = make(map[*mesosproto.TaskID]*mesosproto.AgentID)

					time.Sleep(delay)
				}
			}
		}
	}()
}

func (s *Scheduler) stopReconcile() {
	s.reconcileTimer.Stop()
}

func (s *Scheduler) Dump() interface{} {
	s.RLock()
	defer s.RUnlock()

	return map[string]interface{}{
		"agents":       s.agents,
		"config":       s.cfg,
		"cluster":      s.cluster,
		"mesos_leader": s.leader,
	}
}

func (s *Scheduler) LaunchTasks(tasks []*Task) (map[string]error, error) {
	var (
		wg   sync.WaitGroup
		l    sync.RWMutex
		rets = map[string]error{}
	)

	var (
		groups = [][]*Task{}
		count  = len(tasks)
		step   = s.cfg.MaxTasksPerOffer
	)

	for i := 0; i < count; i = i + step {
		m := i + step
		if m > count {
			m = count
		}
		groups = append(groups, tasks[i:m])
	}

	for _, group := range groups {
		offers := []*Offer{}
		for {
			s.lock()
			candidates := s.strategy.RankAndSort(s.getAgents())
			if len(candidates) <= 0 {
				s.unlock()
				log.Debugln("No enough agent to run tasks, waiting...")
				time.Sleep(10 * time.Millisecond)
				continue
			}

			agent := candidates[0]
			offers = agent.getOffers()

			if len(offers) <= 0 { //?
				s.unlock()
				log.Debugln("No enough resource to run tasks, waiting...")
				time.Sleep(10 * time.Millisecond)
				continue
			}

			for _, task := range group {
				s.addTask(task)
			}

			for _, offer := range offers {
				s.removeOffer(offer)
			}
			s.unlock()
			break
		}

		if err := s.launch(offers, group); err != nil {
			// TODO:handler error
		}

		for _, task := range group {
			wg.Add(1)
			go func(task *Task) {
				defer wg.Done()
				log.Debugf("Waiting for task %s to be running", task.ID())
				for status := range task.GetStatus() {
					log.Debugf("Receiving status %s for task %s", status.GetState().String(), task.ID())
					if task.IsDone(status) {
						l.Lock()
						if err := task.DetectError(status); err != nil {
							rets[task.ID()] = err
						}
						l.Unlock()
						s.removeTask(task.ID())

						return
					}
				}
			}(task)
		}
	}

	return rets, nil
}

func (s *Scheduler) launch(offers []*Offer, tasks []*Task) error {
	ports := make([]uint64, 0)
	for _, offer := range offers {
		ports = append(ports, offer.GetPorts()...)
	}

	for i, task := range tasks {
		if i < len(ports) {
			task.cfg.Port = ports[i]
		}

		task.AgentId = &mesosproto.AgentID{
			Value: proto.String(offers[0].GetAgentId()),
		}

		task.Build()
	}

	appId := strings.SplitN(tasks[0].GetName(), ".", 2)[1]

	for _, t := range tasks {
		task, err := s.db.GetTask(appId, t.GetTaskId().GetValue())
		if err != nil {
			log.Errorln("get task got error: %v", err)
			continue
		}

		task.AgentId = t.AgentId.GetValue()
		task.IP = t.cfg.IP

		if t.cfg.Network == "host" || t.cfg.Network == "bridge" {
			task.IP = offers[0].GetHostname()
		}

		task.Port = t.cfg.Port

		if err := s.db.UpdateTask(appId, task); err != nil {
			log.Errorln("update task got error: %v", err)
			continue
		}
	}

	var (
		offerIds  = []*mesosproto.OfferID{}
		taskInfos = []*mesosproto.TaskInfo{}
	)

	for _, offer := range offers {
		offerIds = append(offerIds, &mesosproto.OfferID{
			Value: proto.String(offer.GetId()),
		})
	}

	for _, task := range tasks {
		taskInfos = append(taskInfos, &task.TaskInfo)
	}

	call := &mesosproto.Call{
		FrameworkId: s.FrameworkId(),
		Type:        mesosproto.Call_ACCEPT.Enum(),
		Accept: &mesosproto.Call_Accept{
			OfferIds: offerIds,
			Operations: []*mesosproto.Offer_Operation{
				&mesosproto.Offer_Operation{
					Type: mesosproto.Offer_Operation_LAUNCH.Enum(),
					Launch: &mesosproto.Offer_Operation_Launch{
						TaskInfos: taskInfos,
					},
				},
			},
			Filters: &mesosproto.Filters{RefuseSeconds: proto.Float64(1)},
		},
	}

	log.Debugf("Launching %d task(s) on agent %s", len(tasks), offers[0].GetHostname())

	// send call
	resp, err := s.Send(call)
	if err != nil {
		return fmt.Errorf("send launch call got error: %v", err)
	}

	if code := resp.StatusCode; code != http.StatusAccepted {
		return fmt.Errorf("launch call send but the status code not 202 got %d", code)
	}

	return nil
}

func (s *Scheduler) lock() {
	s.sem <- struct{}{}
}

func (s *Scheduler) unlock() {
	<-s.sem
}

func (s *Scheduler) Load() map[string]interface{} {
	s.RLock()
	defer s.RUnlock()

	tasks := []string{}
	for _, task := range s.tasks {
		tasks = append(tasks, task.ID())
	}

	return map[string]interface{}{
		"tasks": tasks,
	}
}
