package mesos

import (
	"github.com/Sirupsen/logrus"

	"github.com/Dataman-Cloud/swan/types"
)

// FullTaskEventsAndRecords generate all of current app tasks' db stats into
// sse events & proxy records & dns records
func (s *Scheduler) FullTaskEventsAndRecords() []*types.CombinedEvents {
	ret := make([]*types.CombinedEvents, 0, 0)

	apps, err := s.db.ListApps()
	if err != nil {
		logrus.Errorln("Shceduler.FullTaskEventsAndRecords() db ListApps error:", err)
		return ret
	}

	for _, app := range apps {
		tasks, err := s.db.ListTasks(app.ID)
		if err != nil {
			logrus.Errorln("Scheduler.FullTaskEventsAndRecords() db ListTasks error:", err)
			continue
		}

		for _, task := range tasks {
			task, err := s.db.GetTask(app.ID, task.ID)
			if err != nil {
				logrus.Errorln("Shceduler.TaskEvents() db GetTask error:", err)
				continue
			}
			ver, err := s.db.GetVersion(app.ID, task.Version)
			if err != nil {
				logrus.Errorln("Shceduler.TaskEvents() db GetVersion error:", err)
				continue
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
				AppID:  app.ID,
				TaskID: task.ID,
				IP:     task.IP,
				Weight: task.Weight,
			}

			if ver.Proxy != nil {
				proxyEnabled := ver.Proxy.Enabled
				for i, proxy := range ver.Proxy.Proxies {
					taskEv.GatewayEnabled = proxyEnabled
					taskEv.AppAlias = proxy.Alias
					taskEv.AppListen = proxy.Listen
					taskEv.AppSticky = proxy.Sticky
					if len(task.Ports) > 0 {
						taskEv.Port = task.Ports[i] // currently only support the first port within proxy & events
					}

					cmb := &types.CombinedEvents{
						Event: taskEv,
						DNS:   s.buildAgentDNSRecord(taskEv),
					}
					if proxyEnabled {
						cmb.Proxy = s.buildAgentProxyRecord(taskEv)
					}

					ret = append(ret, cmb)

				}
			}

		}
	}

	return ret
}
