package state

import (
	"fmt"
	"reflect"
	"time"

	"github.com/Dataman-Cloud/swan/src/manager/store"
	"github.com/Dataman-Cloud/swan/src/types"
)

func AppToDB(app *App) *store.Application {
	a := &store.Application{
		ID:        app.ID,
		Name:      app.Name,
		ClusterID: app.ClusterID,
		CreatedAt: app.Created.UnixNano(),
		UpdatedAt: app.Updated.UnixNano(),
		Mode:      string(app.Mode),
	}

	if app.CurrentVersion != nil {
		a.Version = VersionToDB(app.CurrentVersion, app.ID)
	}

	if app.ProposedVersion != nil {
		a.ProposedVersion = VersionToDB(app.ProposedVersion, app.ID)
	}

	if app.StateMachine != nil {
		a.StateMachine = StateMachineToDB(app.StateMachine)
	}

	return a
}

func VersionToDB(version *types.Version, appID string) *store.Version {
	ver := &store.Version{
		ID:          version.ID,
		Command:     version.Command,
		Cpus:        version.CPUs,
		Mem:         version.Mem,
		Disk:        version.Disk,
		Instances:   version.Instances,
		RunAs:       version.RunAs,
		Priority:    version.Priority,
		Labels:      version.Labels,
		Env:         version.Env,
		Constraints: version.Constraints,
		Uris:        version.URIs,
		Ip:          version.IP,
		AppName:     version.AppName,
		AppID:       appID,
		AppVersion:  version.AppVersion,
	}

	if version.Container != nil {
		ver.Container = ContainerToDB(version.Container)
	}

	if version.KillPolicy != nil {
		ver.KillPolicy = KillPolicyToDB(version.KillPolicy)
	}

	if version.UpdatePolicy != nil {
		ver.UpdatePolicy = UpdatePolicyToDB(version.UpdatePolicy)
	}

	if version.HealthCheck != nil {
		ver.HealthCheck = HealthCheckToDB(version.HealthCheck)
	}

	if version.Gateway != nil {
		ver.Gateway = GatewayToDB(version.Gateway)
	}

	return ver
}

func VersionFromDB(ver *store.Version) *types.Version {
	version := &types.Version{
		ID:          ver.ID,
		AppName:     ver.AppName,
		Command:     ver.Command,
		CPUs:        ver.Cpus,
		Mem:         ver.Mem,
		Disk:        ver.Disk,
		Instances:   ver.Instances,
		RunAs:       ver.RunAs,
		Priority:    ver.Priority,
		Labels:      ver.Labels,
		Env:         ver.Env,
		Constraints: ver.Constraints,
		URIs:        ver.Uris,
		IP:          ver.Ip,
		AppVersion:  ver.AppVersion,
	}

	if ver.Container != nil {
		version.Container = ContainerFromDB(ver.Container)
	}

	if ver.KillPolicy != nil {
		version.KillPolicy = KillPolicyFromDB(ver.KillPolicy)
	}

	if ver.UpdatePolicy != nil {
		version.UpdatePolicy = UpdatePolicyFromDB(ver.UpdatePolicy)
	}

	if ver.HealthCheck != nil {
		version.HealthCheck = HealthCheckFromDB(ver.HealthCheck)
	}

	if ver.Gateway != nil {
		version.Gateway = GatewayFromDB(ver.Gateway)
	}

	return version
}

func ContainerToDB(container *types.Container) *store.Container {
	c := &store.Container{
		Type: container.Type,
	}

	if container.Docker != nil {
		c.Docker = DockerToDB(container.Docker)
	}

	if container.Volumes != nil {
		var volumes []*store.Volume

		for _, volume := range container.Volumes {
			volumes = append(volumes, VolumeToDB(volume))
		}

		c.Volumes = volumes
	}

	return c
}

func ContainerFromDB(c *store.Container) *types.Container {
	container := &types.Container{
		Type: c.Type,
	}

	if c.Docker != nil {
		container.Docker = DockerFromDB(c.Docker)
	}

	if c.Volumes != nil {
		var volumes []*types.Volume

		for _, volume := range c.Volumes {
			volumes = append(volumes, VolumeFromFaft(volume))
		}

		container.Volumes = volumes
	}

	return container
}

func DockerToDB(docker *types.Docker) *store.Docker {
	d := &store.Docker{
		ForcePullImage: docker.ForcePullImage,
		Image:          docker.Image,
		Network:        docker.Network,
		Privileged:     docker.Privileged,
	}

	if docker.Parameters != nil {
		var parameters []*store.Parameter
		for _, parameter := range docker.Parameters {
			parameters = append(parameters, ParameterToDB(parameter))
		}

		d.Parameters = parameters
	}

	if docker.PortMappings != nil {
		var portMappings []*store.PortMapping

		for _, portMapping := range docker.PortMappings {
			portMappings = append(portMappings, PortMappingToDB(portMapping))
		}

		d.PortMappings = portMappings
	}

	return d
}

func DockerFromDB(d *store.Docker) *types.Docker {
	docker := &types.Docker{
		ForcePullImage: d.ForcePullImage,
		Image:          d.Image,
		Network:        d.Network,
		Privileged:     d.Privileged,
	}

	if d.Parameters != nil {
		var parameters []*types.Parameter
		for _, parameter := range d.Parameters {
			parameters = append(parameters, ParameterFromDB(parameter))
		}

		docker.Parameters = parameters
	}

	if d.PortMappings != nil {
		var portMappings []*types.PortMapping
		for _, portMapping := range d.PortMappings {
			portMappings = append(portMappings, PortMappingFromDB(portMapping))
		}

		docker.PortMappings = portMappings
	}

	return docker
}

func ParameterToDB(parameter *types.Parameter) *store.Parameter {
	return &store.Parameter{
		Key:   parameter.Key,
		Value: parameter.Value,
	}
}

func ParameterFromDB(p *store.Parameter) *types.Parameter {
	return &types.Parameter{
		Key:   p.Key,
		Value: p.Value,
	}
}

func PortMappingToDB(portMapping *types.PortMapping) *store.PortMapping {
	return &store.PortMapping{
		ContainerPort: portMapping.ContainerPort,
		HostPort:      portMapping.HostPort,
		Name:          portMapping.Name,
		Protocol:      portMapping.Protocol,
	}
}

func PortMappingFromDB(pm *store.PortMapping) *types.PortMapping {
	return &types.PortMapping{
		ContainerPort: pm.ContainerPort,
		HostPort:      pm.HostPort,
		Name:          pm.Name,
		Protocol:      pm.Protocol,
	}
}

func VolumeToDB(volume *types.Volume) *store.Volume {
	return &store.Volume{
		ContainerPath: volume.ContainerPath,
		HostPath:      volume.HostPath,
		Mode:          volume.Mode,
	}
}

func VolumeFromFaft(v *store.Volume) *types.Volume {
	return &types.Volume{
		ContainerPath: v.ContainerPath,
		HostPath:      v.HostPath,
		Mode:          v.Mode,
	}
}

func KillPolicyToDB(killPolicy *types.KillPolicy) *store.KillPolicy {
	return &store.KillPolicy{
		Duration: killPolicy.Duration,
	}
}

func KillPolicyFromDB(k *store.KillPolicy) *types.KillPolicy {
	return &types.KillPolicy{
		Duration: k.Duration,
	}
}

func UpdatePolicyToDB(updatePolicy *types.UpdatePolicy) *store.UpdatePolicy {
	return &store.UpdatePolicy{
		UpdateDelay:  updatePolicy.UpdateDelay,
		MaxRetries:   updatePolicy.MaxRetries,
		MaxFailovers: updatePolicy.MaxFailovers,
		Action:       updatePolicy.Action,
	}
}

func UpdatePolicyFromDB(p *store.UpdatePolicy) *types.UpdatePolicy {
	return &types.UpdatePolicy{
		UpdateDelay:  p.UpdateDelay,
		MaxRetries:   p.MaxRetries,
		MaxFailovers: p.MaxFailovers,
		Action:       p.Action,
	}
}

func HealthCheckToDB(healthCheck *types.HealthCheck) *store.HealthCheck {
	c := &store.HealthCheck{
		ID:                  healthCheck.ID,
		Address:             healthCheck.Address,
		Protocol:            healthCheck.Protocol,
		PortName:            healthCheck.PortName,
		Path:                healthCheck.Path,
		Value:               healthCheck.Value,
		ConsecutiveFailures: healthCheck.ConsecutiveFailures,
		GracePeriodSeconds:  healthCheck.GracePeriodSeconds,
		IntervalSeconds:     healthCheck.IntervalSeconds,
		TimeoutSeconds:      healthCheck.TimeoutSeconds,
		DelaySeconds:        healthCheck.DelaySeconds,
	}

	return c
}

func HealthCheckFromDB(c *store.HealthCheck) *types.HealthCheck {
	healthCheck := &types.HealthCheck{
		ID:                  c.ID,
		Address:             c.Address,
		Protocol:            c.Protocol,
		PortName:            c.PortName,
		Path:                c.Path,
		Value:               c.Value,
		ConsecutiveFailures: c.ConsecutiveFailures,
		GracePeriodSeconds:  c.GracePeriodSeconds,
		IntervalSeconds:     c.IntervalSeconds,
		TimeoutSeconds:      c.TimeoutSeconds,
		DelaySeconds:        c.DelaySeconds,
	}

	return healthCheck
}

func GatewayToDB(gateway *types.Gateway) *store.Gateway {
	g := &store.Gateway{
		Weight:  gateway.Weight,
		Enabled: gateway.Enabled,
	}

	return g
}

func GatewayFromDB(g *store.Gateway) *types.Gateway {
	gateway := &types.Gateway{
		Weight:  g.Weight,
		Enabled: g.Enabled,
	}

	return gateway
}

func SlotToDB(slot *Slot) *store.Slot {
	fmt.Println("=====slot touch====", slot.Healthy())
	s := &store.Slot{
		Index:     int32(slot.Index),
		ID:        slot.ID,
		AppID:     slot.App.ID,
		VersionID: slot.Version.ID,
		Healthy:   slot.Healthy(),
		State:     slot.State,
		Weight:    slot.GetWeight(),
	}

	if slot.CurrentTask != nil {
		s.CurrentTask = TaskToDB(slot.CurrentTask)
	}

	if len(slot.TaskHistory) > 0 {
		s.TaskHistory = make([]*store.Task, 0)
		for _, t := range slot.TaskHistory {
			if t.Slot != nil {
				s.TaskHistory = append(s.TaskHistory, TaskToDB(t))
			}
		}
	}

	return s
}

func SlotFromDB(s *store.Slot, app *App) *Slot {
	slot := &Slot{
		Index:         int(s.Index),
		ID:            s.ID,
		State:         s.State,
		OfferID:       s.CurrentTask.OfferID,
		AgentID:       s.CurrentTask.AgentID,
		Ip:            s.CurrentTask.Ip,
		AgentHostName: s.CurrentTask.AgentHostName,
		healthy:       s.Healthy,
		weight:        s.Weight,
		TaskHistory:   make([]*Task, 0),
	}

	if s.CurrentTask != nil {
		slot.CurrentTask = TaskFromDB(s.CurrentTask, app)
		slot.CurrentTask.Slot = slot
	}

	if len(s.TaskHistory) > 0 {
		for _, t := range s.TaskHistory {
			slot.TaskHistory = append(slot.TaskHistory, TaskFromDB(t, app))
		}
	}

	for _, version := range app.Versions {
		if s.VersionID == version.ID {
			slot.Version = version
		}
	}

	return slot
}

func TaskToDB(task *Task) *store.Task {
	return &store.Task{
		ID:            task.ID,
		AppID:         task.Slot.App.ID,
		VersionID:     task.Version.ID,
		SlotID:        task.Slot.ID,
		State:         task.State,
		Stdout:        task.Stdout,
		Stderr:        task.Stderr,
		HostPorts:     task.HostPorts,
		OfferID:       task.OfferID,
		AgentID:       task.AgentID,
		Ip:            task.Ip,
		AgentHostName: task.AgentHostName,
		Reason:        task.Reason,
		Message:       task.Message,
		CreatedAt:     task.Created.UnixNano(),
		ArchivedAt:    task.ArchivedAt.UnixNano(),
		ContainerId:   task.ContainerId,
		ContainerName: task.ContainerName,
	}
}

func TaskFromDB(t *store.Task, app *App) *Task {
	task := &Task{
		ID:            t.ID,
		State:         t.State,
		Stdout:        t.Stdout,
		Stderr:        t.Stderr,
		HostPorts:     t.HostPorts,
		OfferID:       t.OfferID,
		AgentID:       t.AgentID,
		Ip:            t.Ip,
		AgentHostName: t.AgentHostName,
		Reason:        t.Reason,
		Message:       t.Message,
		Created:       time.Unix(0, t.CreatedAt),
		ContainerId:   t.ContainerId,
		ContainerName: t.ContainerName,
	}

	for _, version := range app.Versions {
		if t.VersionID == version.ID {
			task.Version = version
		}
	}

	return task
}

func OfferAllocatorItemToDB(slotID string, offerInfo *OfferInfo) *store.OfferAllocatorItem {
	item := &store.OfferAllocatorItem{
		OfferID:  offerInfo.OfferID,
		SlotID:   slotID,
		Hostname: offerInfo.Hostname,
		AgentID:  offerInfo.AgentID,
	}

	return item
}

func OfferAllocatorItemFromDB(item *store.OfferAllocatorItem) (slotID string, offerInfo *OfferInfo) {
	return item.SlotID, &OfferInfo{
		OfferID:  item.OfferID,
		Hostname: item.Hostname,
		AgentID:  item.AgentID,
	}
}

func StateMachineFromDB(app *App, machine *store.StateMachine) *StateMachine {
	state := StateFromDB(app, machine.State)

	if state == nil {
		return &StateMachine{state: NewStateNormal(app)}
	}

	return &StateMachine{state: state}
}

func StateFromDB(app *App, state *store.State) State {
	slot, ok := app.GetSlot(int(state.CurrentSlotIndex))
	if !ok {
		return nil
	}

	switch state.Name {
	case APP_STATE_NORMAL:
		return &StateNormal{
			App:  app,
			Name: APP_STATE_NORMAL,
		}
	case APP_STATE_CREATING:
		return &StateCreating{
			App:              app,
			Name:             APP_STATE_CREATING,
			CurrentSlotIndex: int(state.CurrentSlotIndex),
			CurrentSlot:      slot,
			TargetSlotIndex:  int(state.TargetSlotIndex),
		}
	case APP_STATE_SCALE_UP:
		return &StateScaleUp{
			App:              app,
			Name:             APP_STATE_SCALE_UP,
			CurrentSlotIndex: int(state.CurrentSlotIndex),
			CurrentSlot:      slot,
			TargetSlotIndex:  int(state.TargetSlotIndex),
		}
	case APP_STATE_SCALE_DOWN:
		return &StateScaleDown{
			App:              app,
			Name:             APP_STATE_SCALE_DOWN,
			CurrentSlotIndex: int(state.CurrentSlotIndex),
			CurrentSlot:      slot,
			TargetSlotIndex:  int(state.TargetSlotIndex),
		}
	case APP_STATE_CANCEL_UPDATE:
		return &StateCancelUpdate{
			App:              app,
			Name:             APP_STATE_CANCEL_UPDATE,
			CurrentSlotIndex: int(state.CurrentSlotIndex),
			CurrentSlot:      slot,
			TargetSlotIndex:  int(state.TargetSlotIndex),
		}
	case APP_STATE_UPDATING:
		return &StateUpdating{
			App:                 app,
			Name:                APP_STATE_UPDATING,
			CurrentSlotIndex:    int(state.CurrentSlotIndex),
			CurrentSlot:         slot,
			TargetSlotIndex:     int(state.TargetSlotIndex),
			SlotCountNeedUpdate: int(state.SlotCountNeedUpdate),
		}

	case APP_STATE_DELETING:
		return &StateDeleting{
			App:              app,
			Name:             APP_STATE_DELETING,
			CurrentSlotIndex: int(state.CurrentSlotIndex),
			CurrentSlot:      slot,
			TargetSlotIndex:  int(state.TargetSlotIndex),
		}
	default:
		return nil
	}
}

func StateMachineToDB(machine *StateMachine) *store.StateMachine {
	return &store.StateMachine{
		State: StateToDB(machine.CurrentState()),
	}
}

func StateToDB(state State) *store.State {
	stateObj := reflect.ValueOf(state).Elem()
	typeOfT := stateObj.Type()
	s := &store.State{}

	for i := 0; i < stateObj.NumField(); i++ {
		f := stateObj.Field(i)
		switch typeOfT.Field(i).Name {
		case "Name":
			s.Name = f.Interface().(string)
		case "CurrentSlotIndex":
			s.CurrentSlotIndex = int64(f.Interface().(int))
		case "TargetSlotIndex":
			s.TargetSlotIndex = int64(f.Interface().(int))
		case "SlotCountNeedUpdate":
			s.SlotCountNeedUpdate = int64(f.Interface().(int))
		default:
		}
	}

	return s
}
