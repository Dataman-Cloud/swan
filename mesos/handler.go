package mesos

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/Dataman-Cloud/swan/mesosproto"
	"github.com/Dataman-Cloud/swan/types"

	log "github.com/Sirupsen/logrus"
)

const (
	defaultOfferTimeout  = 30 * time.Second
	defaultRefuseTimeout = 5 * time.Second
)

type eventHandler func(*mesosproto.Event)

func (s *Scheduler) subscribedHandler(event *mesosproto.Event) {
	var (
		ev = event.GetSubscribed()
		id = ev.FrameworkId
	)

	log.Printf("Subscription successful with frameworkId %s", id.GetValue())

	s.framework.Id = id

	if err := s.db.UpdateFrameworkId(id.GetValue()); err != nil {
		log.Errorf("update frameworkid got error:%s", err)
	}

	s.startReconcile()
}

func (s *Scheduler) offersHandler(event *mesosproto.Event) {
	var (
		offers = event.Offers.Offers
	)

	log.Debugf("Receiving %d offer(s) from mesos", len(offers))

	for _, offer := range offers {
		agentId := offer.AgentId.GetValue()
		attrs := offer.GetAttributes()
		hostname := offer.GetHostname()

		a := s.getAgent(agentId)
		if a == nil {
			a = newAgent(agentId, hostname, attrs)
			s.addAgent(a)
		}

		s.addOffer(offer)
	}
}

func (s *Scheduler) rescindedHandler(event *mesosproto.Event) {
	var (
		offerId = event.GetRescind().OfferId.GetValue()
	)

	log.Debugln("Receiving rescind msg for offer ", offerId)

	for _, agent := range s.getAgents() {
		if offer := agent.getOffer(offerId); offer != nil {
			s.removeOffer(offer)
			break
		}
	}
}

func (s *Scheduler) updateHandler(event *mesosproto.Event) {
	var (
		status  = event.GetUpdate().GetStatus()
		state   = status.GetState()
		taskId  = status.TaskId.GetValue()
		healthy = status.GetHealthy()
		data    = status.GetData()
	)

	log.Debugf("Received status update %s for task %s %s %s", status.GetState(), taskId, status.GetReason().String(), status.GetMessage())

	var appId string
	// get appId
	parts := strings.SplitN(taskId, ".", 3)
	if len(parts) >= 3 {
		appId = parts[2]
	}

	// get container id & name
	var cinfos []struct {
		ID   string `json:"Id"`
		Name string `json:"Name"`
	}
	json.Unmarshal(data, &cinfos)

	// only broadcast unhealthy event and return
	if state == mesosproto.TaskState_TASK_FINISHED ||
		state == mesosproto.TaskState_TASK_UNKNOWN ||
		state == mesosproto.TaskState_TASK_KILLING {
		taskEv := &types.TaskEvent{
			Type:   types.EventTypeTaskUnhealthy,
			AppID:  appId,
			TaskID: taskId,
		}
		if err := s.eventmgr.broadcast(taskEv); err != nil {
			log.Errorln("broadcast task event got error:", err)
		}
		if err := s.broadcastEventRecords(taskEv); err != nil {
			log.Errorln("broadcast to sync proxy & dns records error:", err)
		}
		return
	}

	// obtain db task & update
	task, err := s.db.GetTask(appId, taskId)
	if err != nil {
		return
	}

	task.Status = state.String()
	if len(cinfos) > 0 {
		if cid := cinfos[0].ID; cid != "" {
			task.ContainerID = cid
		}
		if cname := cinfos[0].Name; cname != "" {
			task.ContainerName = cname
		}
	}

	ver, err := s.db.GetVersion(appId, task.Version) // task corresponding version
	if err != nil {
		log.Errorf("find task version got error: %v. task %s, version %s", err, task.ID, task.Version)
		return
	}

	var (
		previousHealthy = task.Healthy // save previous
	)

	if task.Healthy != types.TaskHealthyUnset {
		task.Healthy = types.TaskUnHealthy
		if healthy {
			task.Healthy = types.TaskHealthy
		}
	}

	if state != mesosproto.TaskState_TASK_RUNNING {
		task.ErrMsg = status.GetReason().String() + ":" + status.GetMessage()
	}

	if err := s.db.UpdateTask(appId, task); err != nil {
		log.Errorf("update task status error: %v, %s", err, state.String())
		return
	}

	// broadcasting task events
	log.Debugf("task %s healthy: %s --> %s (%s)", taskId, previousHealthy, task.Healthy, task.Status)
	if previousHealthy != task.Healthy { // skip on no-change
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

		var (
			alias        string
			proxyEnabled bool
			listen       string
			sticky       bool
		)
		if ver.Proxy != nil {
			proxyEnabled = ver.Proxy.Enabled
			alias = ver.Proxy.Alias
			listen = ver.Proxy.Listen
			sticky = ver.Proxy.Sticky
		}

		taskEv := &types.TaskEvent{
			Type:           evType,
			AppID:          appId,
			AppAlias:       alias,
			AppListen:      listen,
			AppSticky:      sticky,
			TaskID:         taskId,
			IP:             task.IP,
			Port:           task.Port,
			Weight:         task.Weight,
			GatewayEnabled: proxyEnabled,
		}

		if err := s.eventmgr.broadcast(taskEv); err != nil {
			log.Errorln("broadcast task event got error:", err)
		}

		if err := s.broadcastEventRecords(taskEv); err != nil {
			log.Errorln("broadcast to sync proxy & dns records error:", err)
			// TODO: memo db task errmsg
		}
	}

	// reschedule failed tasks
	if state != mesosproto.TaskState_TASK_RUNNING {
		if task.Retries >= task.MaxRetries {
			// no more retry
			log.Debugln("task", taskId, "maxRetries:", task.MaxRetries, "retries:", task.Retries)
			return
		}

		if t := s.getPendingTask(taskId); t != nil {
			// task already being rescheduling.
			log.Debugln("task already in rescheduling", taskId)
			return
		}

		app, err := s.db.GetApp(appId)
		if err != nil {
			if strings.Contains(err.Error(), "not exists") {
				return
			}

			log.Errorf("updateHandler:rescheduleTask s.db.GetApp() failed %s", err)
			return
		}

		if app.OpStatus == types.OpStatusDeleting {
			// deleting task no need to reschedule.
			return
		}

		go s.rescheduleTask(appId, task)
	}

	return
}

func (s *Scheduler) heartbeatHandler(event *mesosproto.Event) {
	log.Debugln("Receive heartbeat msg from mesos")
}

func (s *Scheduler) errHandler(event *mesosproto.Event) {
	ev := event.GetError()

	log.Debugf("Receive error msg %s", ev.GetMessage())
	s.reconnect()
}

func (s *Scheduler) failureHandler(event *mesosproto.Event) {
	var (
		ev      = event.GetFailure()
		agentId = ev.GetAgentId()
		execId  = ev.GetExecutorId()
		status  = ev.GetStatus()
	)

	if execId != nil {
		log.Debugf("Receive failure msg for executor %s terminated with status %d", execId.GetValue(), status)
		return
	}

	log.Debugf("Receive msg for agent %s removed.", agentId.GetValue())

}

func (s *Scheduler) messageHandler(event *mesosproto.Event) {
}
