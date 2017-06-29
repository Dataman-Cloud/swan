package mesos

import (
	"github.com/Sirupsen/logrus"

	"github.com/Dataman-Cloud/swan/types"
)

// TaskEvents generate all of current app tasks' db stats into sse events
func (s *Scheduler) TaskEvents() []*types.TaskEvent {
	ret := make([]*types.TaskEvent, 0, 0)

	apps, err := s.db.ListApps()
	if err != nil {
		logrus.Errorln("Shceduler.TaskEvents() db ListApps error:", err)
		return ret
	}

	for _, app := range apps {
		for _, task := range app.Tasks {

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

			var (
				alias        string
				proxyEnabled bool
			)
			if ver.Proxy != nil {
				alias = ver.Proxy.Alias
				proxyEnabled = ver.Proxy.Enabled
			}

			ret = append(ret, &types.TaskEvent{
				Type:           evType,
				AppID:          app.ID,
				AppAlias:       alias,
				TaskID:         task.ID,
				IP:             task.IP,
				Port:           task.Port,
				Weight:         task.Weight,
				GatewayEnabled: proxyEnabled,
			})
		}
	}

	return ret
}
