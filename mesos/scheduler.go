package mesos

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"

	magent "github.com/Dataman-Cloud/swan/mesos/agent"
	"github.com/Dataman-Cloud/swan/mesos/filter"
	"github.com/Dataman-Cloud/swan/mesos/strategy"
	"github.com/Dataman-Cloud/swan/mesosproto"
	"github.com/Dataman-Cloud/swan/mole"
	"github.com/Dataman-Cloud/swan/store"
	"github.com/Dataman-Cloud/swan/types"
	"github.com/Dataman-Cloud/swan/utils"
)

type SchedulerConfig struct {
	ZKHost []string
	ZKPath string

	Strategy string

	ReconciliationInterval  float64
	ReconciliationStep      int64
	ReconciliationStepDelay float64

	HeartbeatTimeout        float64
	MaxTasksPerOffer        int
	EnableCapabilityKilling bool
	EnableCheckPoint        bool
}

// Scheduler represents a client interacting with mesos master via x-protobuf
type Scheduler struct {
	http      *httpClient // mesos scheduler http client
	cfg       *SchedulerConfig
	framework *mesosproto.FrameworkInfo

	quit chan struct{} // TODO on followers

	leader  string
	cluster string // name of mesos cluster

	db store.Store

	handlers map[mesosproto.Event_Type]eventHandler

	sync.RWMutex                          // protect followings two
	agents       map[string]*magent.Agent // holding offers (agents)
	pendingTasks map[string]*Task

	reconcileTimer *time.Ticker

	strategy strategy.Strategy
	filters  []filter.Filter

	eventmgr *eventManager

	clusterMaster *mole.Master

	sem chan struct{} // to order the mesos offer acquirement by multi app launching
}

// NewScheduler...
func NewScheduler(cfg *SchedulerConfig, db store.Store, clusterMaster *mole.Master) (*Scheduler, error) {
	s := &Scheduler{
		cfg:           cfg,
		quit:          make(chan struct{}),
		agents:        make(map[string]*magent.Agent),
		pendingTasks:  make(map[string]*Task),
		db:            db,
		strategy:      strategy.NewBinPackStrategy(), // default strategy
		filters:       []filter.Filter{filter.NewConstraintsFilter(), filter.NewResourceFilter()},
		eventmgr:      NewEventManager(),
		clusterMaster: clusterMaster,
		sem:           make(chan struct{}, 1), // allow only one offer acquirement at one time
	}

	switch cfg.Strategy {
	case "random":
		s.strategy = strategy.NewRandomStrategy()
	case "binpack", "binpacking":
		s.strategy = strategy.NewBinPackStrategy()
	case "spread":
		s.strategy = strategy.NewSpreadStrategy()
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

	s.framework = s.buildFramework()

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

// Cluster return current mesos cluster's name
func (s *Scheduler) ClusterName() string {
	return s.cluster
}

func (s *Scheduler) FrameworkId() *mesosproto.FrameworkID {
	return s.framework.Id
}

// SendCall send mesos request against the mesos master's scheduler api endpoint.
// note it's the caller's responsibility to deal with the SendCall() error
func (s *Scheduler) SendCall(call *mesosproto.Call, expectCode int) (*http.Response, error) {
	payload, err := proto.Marshal(call)
	if err != nil {
		return nil, err
	}

	resp, err := s.http.send(payload)
	if err != nil {
		return nil, err
	}

	if code := resp.StatusCode; code != expectCode {
		bs, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("sendCall() with unexpected response [%d] - [%s]", code, string(bs))
	}

	return resp, nil
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
	resp, err := s.SendCall(call, http.StatusOK)
	if err != nil {
		log.Errorln("connect().SendCall() error:", err)
		return nil, fmt.Errorf("subscribe to mesos leader [%s] error [%v]", l, err)
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

		// emit event status to pending task
		log.Debugf("Finding task %s to send status %s", taskId, state.String())
		s.sendTaskStatus(taskId, status)

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

	f := magent.NewOffer(offer)

	log.Debugf("Received offer %s with resource cpus:[%.2f] mem:[%.2fG] disk:[%.2fG] ports:%v from agent %s",
		f.GetId(), f.GetCpus(), f.GetMem()/1024, f.GetDisk()/1024, f.GetPortRange(), f.GetHostname())

	a.AddOffer(f)
	time.AfterFunc(time.Second*5, func() { // release the offer later
		if s.removeOffer(f) {
			s.declineOffers([]*magent.Offer{f})
		}
	})
}

func (s *Scheduler) removeOffer(offer *magent.Offer) bool {
	log.Debugln("Removing offer ", offer.GetId())

	a := s.getAgent(offer.GetAgentId())
	if a == nil {
		return false
	}

	found := a.RemoveOffer(offer.GetId())
	if a.Empty() {
		s.removeAgent(offer.GetAgentId())
	}

	return found
}

func (s *Scheduler) declineOffers(offers []*magent.Offer) error {
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

	if _, err := s.SendCall(call, http.StatusAccepted); err != nil {
		log.Errorln("declineOffers().SendCall() error:", err)
		return err
	}

	return nil
}

func (s *Scheduler) addAgent(agent *magent.Agent) {
	s.Lock()
	defer s.Unlock()

	s.agents[agent.ID()] = agent
}

func (s *Scheduler) getAgent(agentId string) *magent.Agent {
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

func (s *Scheduler) getAgents() []*magent.Agent {
	s.RLock()
	defer s.RUnlock()

	agents := make([]*magent.Agent, 0)
	for _, agent := range s.agents {
		agents = append(agents, agent)
	}

	return agents
}

func (s *Scheduler) addPendingTask(t *Task) {
	log.Debugf("Add pending task %s", t.TaskId.GetValue())

	s.Lock()
	defer s.Unlock()

	s.pendingTasks[t.TaskId.GetValue()] = t
}

func (s *Scheduler) getPendingTask(taskId string) *Task {
	s.RLock()
	defer s.RUnlock()

	task, ok := s.pendingTasks[taskId]
	if !ok {
		return nil
	}

	return task
}

func (s *Scheduler) removePendingTask(taskId string) {
	log.Debugf("Remove pending task %s", taskId)

	s.Lock()
	defer s.Unlock()

	t, ok := s.pendingTasks[taskId]
	if !ok {
		return
	}

	close(t.updates)

	delete(s.pendingTasks, taskId)
}

// send update event to pending task's buffered channel
func (s *Scheduler) sendTaskStatus(taskID string, status *mesosproto.TaskStatus) {
	s.RLock()
	defer s.RUnlock()

	t, ok := s.pendingTasks[taskID]
	if !ok {
		return
	}

	t.updates <- status
}

func (s *Scheduler) KillTask(taskId, agentId string, gracePeriod int64) error {
	log.Printf("Killing task %s with agentId %s", taskId, agentId)

	if agentId == "" {
		log.Warnf("agentId of task %s is empty, ignore", taskId)
		return nil
	}

	t := NewTask(nil, taskId, "")

	s.addPendingTask(t)
	defer s.removePendingTask(taskId) // prevent leak

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

	if gracePeriod > 0 {
		call.Kill.KillPolicy = &mesosproto.KillPolicy{
			GracePeriod: &mesosproto.DurationInfo{
				Nanoseconds: proto.Int64(gracePeriod * 1000 * 1000),
			},
		}
	}

	// send call
	if _, err := s.SendCall(call, http.StatusAccepted); err != nil {
		log.Errorln("KillTask().SendCall() error:", err)
		return err
	}

	log.Debugf("Waiting for task %s to be killed by mesos", taskId)
	for status := range t.GetStatus() {
		log.Debugf("Receiving status %s for task %s", status.GetState().String(), taskId)
		if IsTaskDone(status) {
			log.Printf("Task %s killed", taskId)
			break
		}
	}

	// ensure dns & proxy records could be cleaned up
	parts := strings.SplitN(taskId, ".", 3)
	if len(parts) >= 3 {
		appId := parts[2]
		s.broadCastCleanupEvents(appId, taskId)
	}

	return nil
}

func (s *Scheduler) reconcileTasks(tasks map[*mesosproto.TaskID]*mesosproto.AgentID) error {
	log.Printf("reconcile %d tasks", len(tasks))
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

	if _, err := s.SendCall(call, http.StatusAccepted); err != nil {
		log.Errorln("reconcileTasks().SendCall() error:", err)
		return err
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
		if _, err := s.SendCall(call, http.StatusAccepted); err != nil {
			log.Errorln("AckUpdateEvent().SendCall() error:", err)
			return err
		}
	}

	return nil
}

func (s *Scheduler) SubscribeEvent(w io.Writer, remote string) error {
	if s.eventmgr.Full() {
		return fmt.Errorf("%s", "too many event clients")
	}

	s.eventmgr.subscribe(remote, w)
	s.eventmgr.wait(remote)

	return nil
}

func (s *Scheduler) runReconcile() {
	var (
		step  = int(s.cfg.ReconciliationStep)
		delay = time.Duration(s.cfg.ReconciliationStepDelay) * time.Second
	)

	log.Println("Start task reconciliation with the Mesos master")

	apps, err := s.db.ListApps()
	if err != nil {
		log.Errorf("List app got error for task reconcile. %v", err)
		return
	}

	// filter only the OpStatusNoop apps
	var idx int
	for _, app := range apps {
		if app.OpStatus == types.OpStatusNoop {
			apps[idx] = app
			idx++
		}
	}
	apps = apps[:idx]

	tasks := make([]*types.Task, 0)

	for _, app := range apps {
		ts, err := s.db.ListTasks(app.ID)
		if err != nil {
			log.Errorf("List tasks got error: %v", err)
			continue
		}

		for _, t := range ts {
			// skip pending tasks
			if t.Status == "pending" {
				continue
			}
			tasks = append(tasks, t)
		}
	}

	var (
		total = len(tasks)
		send  = 0
	)
	log.Printf("%d tasks to be reconciled", total)

	m := make(map[*mesosproto.TaskID]*mesosproto.AgentID)
	for _, task := range tasks {
		taskID := &mesosproto.TaskID{Value: proto.String(task.ID)}
		agentID := &mesosproto.AgentID{Value: proto.String(task.AgentId)}

		m[taskID] = agentID

		if len(m) >= step || (len(m)+send) >= total {
			if err := s.reconcileTasks(m); err != nil {
				log.Errorf("reconcile %d tasks got error: %v", len(m), err)
			}

			send += len(m)
			m = make(map[*mesosproto.TaskID]*mesosproto.AgentID)
			time.Sleep(delay)
		}
	}
}

func (s *Scheduler) startReconcileLoop() {
	var (
		interval = time.Duration(s.cfg.ReconciliationInterval) * time.Second
	)

	s.reconcileTimer = time.NewTicker(interval)

	go func() {
		s.runReconcile() // run reconcile once immediately on start up

		for range s.reconcileTimer.C {
			s.runReconcile()
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

func (s *Scheduler) Offers() interface{} {
	s.RLock()
	defer s.RUnlock()

	offers := make([]*magent.Offer, 0)
	for _, a := range s.agents {
		for _, f := range a.GetOffers() {
			offers = append(offers, f)
		}
	}

	return offers
}

// wait proper offers according by grouped-task's constraints & resources requirments
func (s *Scheduler) waitOffers(filterOpts *filter.FilterOptions) ([]*magent.Offer, error) {
	log.Debugln("Finding suitable agent to run tasks")

	var (
		offers         = make([]*magent.Offer, 0, 0)
		maxWait        = time.Second * 5
		waitTimeout    = time.After(maxWait)
		err            error // global final error
		filteredAgents []*magent.Agent
	)

	for {
		// make the offer exclusively
		s.lockOffer()

		select {
		case <-waitTimeout:
			s.unlockOffer() // make other launchers avaliable to mesos offers
			if err != nil {
				return nil, fmt.Errorf("without proper agents: %s", err.Error())
			} else {
				return nil, fmt.Errorf("wait offer timeout in %s", maxWait)
			}

		default:
			// ensure we have at least one agent avaliable
			agents := s.getAgents()
			if len(agents) == 0 {
				log.Warnf("without proper offers, no agents avaliable, retrying ...")
				goto RETRY
			}

			// filter agents by resources & constraints
			filteredAgents, err = filter.ApplyFilters(s.filters, filterOpts, agents)
			if err != nil {
				log.Warnf("without proper offers: [%v], retrying ...", err)
				goto RETRY
			}

			offers = filteredAgents[0].GetOffers()
			log.Debugf("Found %d agents with %d offers avaliable", len(filteredAgents), len(offers))

			// now we got proper offers
			if len(offers) > 0 {
				for _, offer := range offers {
					s.removeOffer(offer)
				}
				s.unlockOffer()
				return offers, nil
			}

		RETRY:
			// no proper offers, delay for a while and try again ...
			s.unlockOffer()
			time.Sleep(time.Millisecond * 500)
		}
	}
}

func (s *Scheduler) LaunchTasks(tasks []*Task) error {

	var (
		wg     sync.WaitGroup
		groups = [][]*Task{}
		count  = len(tasks)
		step   = s.cfg.MaxTasksPerOffer
		cfg    = tasks[0].cfg
	)

	var errs struct {
		m []error
		sync.Mutex
	}
	errs.m = make([]error, 0, 0)

	// cut all tasks into sub pieces
	for i := 0; i < count; i = i + step {
		m := i + step
		if m > count {
			m = count
		}
		groups = append(groups, tasks[i:m])
	}

	// launch each sub-pieces of tasks
	for _, group := range groups {
		// save db tasks
		for _, task := range group {
			var (
				restart     = task.cfg.RestartPolicy
				retries     = 3
				healthCheck = task.cfg.HealthCheck
				healthy     = types.TaskHealthyUnset
				versionId   = task.cfg.Version
				taskId      = task.TaskId.GetValue()
				taskName    = task.GetName()
			)

			if restart != nil && restart.Retries >= 0 {
				retries = restart.Retries
			}

			if healthCheck != nil {
				healthy = types.TaskUnHealthy
			}

			dbtask := &types.Task{
				ID:         taskId,
				Name:       taskName,
				Weight:     100,
				Status:     "pending",
				Healthy:    healthy,
				Version:    versionId,
				MaxRetries: retries,
				Created:    time.Now(),
				Updated:    time.Now(),
			}

			parts := strings.SplitN(taskId, ".", 3)
			if len(parts) < 3 {
				return fmt.Errorf("malformed taskId: %s", taskId)
			}

			appId := parts[2]

			log.Debugf("Create task %s in db", task.ID)
			if err := s.db.CreateTask(appId, dbtask); err != nil {
				return fmt.Errorf("create db task failed: %v", err)
			}
		}

		// try to use filter options to obtain proper offers
		filterOpts := &filter.FilterOptions{
			ResRequired: cfg.ResourcesRequired(),
			Replicas:    len(group),
			Constraints: cfg.Constraints,
		}

		// try obtain proper offers
		offers, err := s.waitOffers(filterOpts)
		if err != nil {
			return err
		}

		// add penging tasks before actually launching the mesos tasks
		for _, task := range group {
			s.addPendingTask(task)
		}

		// launch group tasks with specified offers
		if err := s.launchGroupTasksWithOffers(offers, group); err != nil {
			// NOTE: required to prevent memory leaks if tasks not emited to mesos master.
			for _, task := range group {
				s.removePendingTask(task.ID())
			}
			return err
		}

		// wait grouped-tasks status in background
		for _, task := range group {
			wg.Add(1)
			go func(task *Task) {
				defer wg.Done()
				log.Debugf("Waiting for task %s to be running", task.ID())

				for status := range task.GetStatus() {
					log.Debugf("Receiving status %s for task %s", status.GetState().String(), task.ID())

					if !IsTaskDone(status) {
						continue
					}

					if err := DetectTaskError(status); err != nil {
						log.Errorf("Launch task %s failed: %v", task.ID(), err)
						errs.Lock()
						errs.m = append(errs.m, err)
						errs.Unlock()
					} else {
						log.Printf("Launch task %s succeed", task.ID())
					}
					s.removePendingTask(task.ID())
					return
				}
			}(task)
		}
	}

	wg.Wait()

	if len(errs.m) == 0 {
		return nil
	}
	return fmt.Errorf("%d tasks launch failed", len(errs.m))
}

// launch grouped runtime tasks with specified mesos offers
func (s *Scheduler) launchGroupTasksWithOffers(offers []*magent.Offer, tasks []*Task) error {
	ports := make([]uint64, 0)
	for _, offer := range offers {
		ports = append(ports, offer.GetPorts()...)
	}

	var idx int
	for _, task := range tasks {
		// port allocated only on bridge network.
		if task.cfg.Network == "bridge" {
			num := len(task.cfg.PortMappings)
			if idx+num > len(ports) {
				return errors.New("no enough ports avaliable")
			}

			task.cfg.Ports = ports[idx : idx+num]
			idx += num
		}

		task.AgentId = &mesosproto.AgentID{
			Value: proto.String(offers[0].GetAgentId()),
		}

		// Set IP for build. Host or Bridge mode use mesos agent IP.
		if task.cfg.Network == "host" || task.cfg.Network == "bridge" {
			task.cfg.IP = offers[0].GetHostname()
		}

		task.Build()
	}

	appId := strings.SplitN(tasks[0].GetName(), ".", 2)[1]

	// memo update each db tasks' AgentID, IP, Port ...
	for _, t := range tasks {
		dbtask, err := s.db.GetTask(appId, t.GetTaskId().GetValue())
		if err != nil {
			log.Errorln("get task got error: %v", err)
			continue
		}

		dbtask.AgentId = t.AgentId.GetValue()
		dbtask.IP = t.cfg.IP
		if t.cfg.Network == "host" || t.cfg.Network == "bridge" {
			dbtask.IP = offers[0].GetHostname()
		}

		dbtask.Ports = t.cfg.Ports

		if err := s.db.UpdateTask(appId, dbtask); err != nil {
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

	log.Printf("Launching %d task(s) on agent %s", len(tasks), offers[0].GetHostname())

	// send call
	if _, err := s.SendCall(call, http.StatusAccepted); err != nil {
		log.Errorln("launch().SendCall() error:", err)
		return fmt.Errorf("send launch call got error: %v", err)
	}

	return nil
}

func (s *Scheduler) lockOffer() {
	s.sem <- struct{}{}
}

func (s *Scheduler) unlockOffer() {
	<-s.sem
}

func (s *Scheduler) Load() map[string]interface{} {
	s.RLock()
	defer s.RUnlock()

	tasks := []string{}
	for _, task := range s.pendingTasks {
		tasks = append(tasks, task.ID())
	}

	return map[string]interface{}{
		"tasks": tasks,
	}
}

func (s *Scheduler) rescheduleTask(appId string, task *types.Task) {
	log.Println("Rescheduling task:", task.Name)

	var (
		taskName = task.Name
		verId    = task.Version
	)

	ver, err := s.db.GetVersion(appId, verId)
	if err != nil {
		log.Errorf("rescheduleTask(): get version failed: %s", err)
		return
	}

	seq := strings.SplitN(task.Name, ".", 2)[0]
	idx, _ := strconv.Atoi(seq)
	cfg := types.NewTaskConfig(ver, idx)

	var (
		name       = taskName
		id         = fmt.Sprintf("%s.%s", utils.RandomString(12), name)
		restart    = ver.RestartPolicy
		retries    = task.Retries
		maxRetries = 3
	)

	if restart != nil && restart.Retries > maxRetries {
		maxRetries = restart.Retries
	}

	dbtask := &types.Task{
		ID:         id,
		Name:       name,
		Weight:     100,
		Status:     "retrying",
		Version:    verId,
		Healthy:    types.TaskHealthyUnset,
		Retries:    retries + 1,
		MaxRetries: maxRetries,
		Created:    time.Now(),
		Updated:    time.Now(),
	}

	if ver.IsHealthSet() {
		dbtask.Healthy = types.TaskUnHealthy
	}

	histories := task.Histories

	if len(histories) >= maxRetries {
		histories = histories[1:]
	}

	for _, history := range histories {
		dbtask.Histories = append(dbtask.Histories, history)
	}

	task.Histories = []*types.Task{}

	dbtask.Histories = append(dbtask.Histories, task)

	if err := s.db.CreateTask(appId, dbtask); err != nil {
		log.Errorf("rescheduleTask(): create dbtask %s error: %v", dbtask.ID, err)
		return
	}

	m := NewTask(cfg, dbtask.ID, dbtask.Name)

	if err := s.LaunchTasks([]*Task{m}); err != nil {
		log.Errorf("rescheduleTask(): launch task %s error: %v", dbtask.ID, err)

		dbtask.Status = "Failed"
		dbtask.ErrMsg = fmt.Sprintf("launch task failed: %v", err)

		if err = s.db.UpdateTask(appId, dbtask); err != nil {
			log.Errorf("rescheduleTask(): update dbtask %s error: %v", dbtask.ID, err)
		}

		return
	}

	log.Printf("Reschedule task %s succeed", task.Name)
}

func (s *Scheduler) SendEvent(appId string, task *types.Task) error {
	ver, err := s.db.GetVersion(appId, task.Version)
	if err != nil {
		return fmt.Errorf("Shceduler.SendEvent() db GetVersion error: %v", err)
	}

	evType := types.EventTypeTaskUnhealthy
	switch task.Healthy {
	case types.TaskHealthy:
		evType = types.EventTypeTaskHealthy
	case types.TaskHealthyUnset:
		if task.Status == "TASK_RUNNING" {
			evType = types.EventTypeTaskHealthy
		}
	case types.TaskUnHealthy:
	}

	taskEv := &types.TaskEvent{
		Type:   evType,
		AppID:  appId,
		TaskID: task.ID,
		IP:     task.IP,
		Weight: task.Weight,
	}

	if ver.Proxy != nil {
		taskEv.GatewayEnabled = ver.Proxy.Enabled
		taskEv.AppAlias = ver.Proxy.Alias
		taskEv.AppListen = ver.Proxy.Listen
		taskEv.AppSticky = ver.Proxy.Sticky
	}

	if len(task.Ports) > 0 {
		taskEv.Port = task.Ports[0] // currently only support the first port within proxy & events
	}

	if err := s.eventmgr.broadcast(taskEv); err != nil {
		return fmt.Errorf("Shceduler.SendEvent(): broadcast task event got error: %v", err)
	}

	if err := s.broadcastEventRecords(taskEv); err != nil {
		return fmt.Errorf("Shceduler.SendEvent(): broadcast to sync proxy & dns records error: %v", err)
		// TODO: memo db task errmsg
	}

	return nil
}
