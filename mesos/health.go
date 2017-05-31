package mesos

import (
	"github.com/Dataman-Cloud/swan/types"
)

func (s *Scheduler) TaskEvents() []*types.TaskEvent {
	var events []*types.TaskEvent

	//apps, err := s.db.ListApps()
	//if err != nil {
	//	return events
	//}

	//for _, app := range apps {
	//for _, task := range app.Tasks {
	//	typ := types.EventTypeTaskUnhealthy
	//	if task.Healthy {
	//		typ = types.EventTypeTaskHealthy
	//	}

	//	events = append(events, &types.TaskEvent{
	//		Type:           typ,
	//		AppID:          app.ID,
	//		AppAlias:       app.Alias,
	//		ID:             task.ID,
	//		IP:             task.IP,
	//		Port:           task.Port,
	//		Weight:         task.Weight,
	//		GatewayEnabled: task.ProxyEnabled,
	//	})

	//}
	//}

	return events
}
