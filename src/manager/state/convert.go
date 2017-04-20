package state

import (
	"reflect"
	"time"

	"github.com/Dataman-Cloud/swan/src/manager/store"
	"github.com/Dataman-Cloud/swan/src/types"
)

func AppToRaft(app *App) *store.Application {
	raftApp := &store.Application{
		ID:        app.ID,
		Name:      app.Name,
		CreatedAt: app.Created.UnixNano(),
		UpdatedAt: app.Updated.UnixNano(),
	}

	if app.CurrentVersion != nil {
		raftApp.Version = VersionToRaft(app.CurrentVersion, app.ID)
	}

	if app.ProposedVersion != nil {
		raftApp.ProposedVersion = VersionToRaft(app.ProposedVersion, app.ID)
	}

	if app.StateMachine != nil {
		raftApp.StateMachine = StateMachineToRaft(app.StateMachine)
	}

	return raftApp
}

func VersionToRaft(version *types.Version, appID string) *store.Version {
	raftVersion := &store.Version{
		ID:          version.ID,
		Command:     version.Command,
		Args:        version.Args,
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
		raftVersion.Container = ContainerToRaft(version.Container)
	}

	if version.KillPolicy != nil {
		raftVersion.KillPolicy = KillPolicyToRaft(version.KillPolicy)
	}

	if version.UpdatePolicy != nil {
		raftVersion.UpdatePolicy = UpdatePolicyToRaft(version.UpdatePolicy)
	}

	if version.HealthCheck != nil {
		raftVersion.HealthCheck = HealthCheckToRaft(version.HealthCheck)
	}

	return raftVersion
}

func VersionFromRaft(raftVersion *store.Version) *types.Version {
	version := &types.Version{
		ID:          raftVersion.ID,
		AppName:     raftVersion.AppName,
		Command:     raftVersion.Command,
		Args:        raftVersion.Args,
		CPUs:        raftVersion.Cpus,
		Mem:         raftVersion.Mem,
		Disk:        raftVersion.Disk,
		Instances:   raftVersion.Instances,
		RunAs:       raftVersion.RunAs,
		Priority:    raftVersion.Priority,
		Labels:      raftVersion.Labels,
		Env:         raftVersion.Env,
		Constraints: raftVersion.Constraints,
		URIs:        raftVersion.Uris,
		IP:          raftVersion.Ip,
		AppVersion:  raftVersion.AppVersion,
	}

	if raftVersion.Container != nil {
		version.Container = ContainerFromRaft(raftVersion.Container)
	}

	if raftVersion.KillPolicy != nil {
		version.KillPolicy = KillPolicyFromRaft(raftVersion.KillPolicy)
	}

	if raftVersion.UpdatePolicy != nil {
		version.UpdatePolicy = UpdatePolicyFromRaft(raftVersion.UpdatePolicy)
	}

	if raftVersion.HealthCheck != nil {
		version.HealthCheck = HealthCheckFromRaft(raftVersion.HealthCheck)
	}

	return version
}

func ContainerToRaft(container *types.Container) *store.Container {
	raftContainer := &store.Container{
		Type: container.Type,
	}

	if container.Docker != nil {
		raftContainer.Docker = DockerToRaft(container.Docker)
	}

	if container.Volumes != nil {
		var volumes []*store.Volume

		for _, volume := range container.Volumes {
			volumes = append(volumes, VolumeToRaft(volume))
		}

		raftContainer.Volumes = volumes
	}

	return raftContainer
}

func ContainerFromRaft(raftContainer *store.Container) *types.Container {
	container := &types.Container{
		Type: raftContainer.Type,
	}

	if raftContainer.Docker != nil {
		container.Docker = DockerFromRaft(raftContainer.Docker)
	}

	if raftContainer.Volumes != nil {
		var volumes []*types.Volume

		for _, volume := range raftContainer.Volumes {
			volumes = append(volumes, VolumeFromFaft(volume))
		}

		container.Volumes = volumes
	}

	return container
}

func DockerToRaft(docker *types.Docker) *store.Docker {
	raftDocker := &store.Docker{
		ForcePullImage: docker.ForcePullImage,
		Image:          docker.Image,
		Network:        docker.Network,
		Privileged:     docker.Privileged,
	}

	if docker.Parameters != nil {
		var parameters []*store.Parameter
		for _, parameter := range docker.Parameters {
			parameters = append(parameters, ParameterToRaft(parameter))
		}

		raftDocker.Parameters = parameters
	}

	if docker.PortMappings != nil {
		var portMappings []*store.PortMapping

		for _, portMapping := range docker.PortMappings {
			portMappings = append(portMappings, PortMappingToRaft(portMapping))
		}

		raftDocker.PortMappings = portMappings
	}

	return raftDocker
}

func DockerFromRaft(raftDocker *store.Docker) *types.Docker {
	docker := &types.Docker{
		ForcePullImage: raftDocker.ForcePullImage,
		Image:          raftDocker.Image,
		Network:        raftDocker.Network,
		Privileged:     raftDocker.Privileged,
	}

	if raftDocker.Parameters != nil {
		var parameters []*types.Parameter
		for _, parameter := range raftDocker.Parameters {
			parameters = append(parameters, ParameterFromRaft(parameter))
		}

		docker.Parameters = parameters
	}

	if raftDocker.PortMappings != nil {
		var portMappings []*types.PortMapping
		for _, portMapping := range raftDocker.PortMappings {
			portMappings = append(portMappings, PortMappingFromRaft(portMapping))
		}

		docker.PortMappings = portMappings
	}

	return docker
}

func ParameterToRaft(parameter *types.Parameter) *store.Parameter {
	return &store.Parameter{
		Key:   parameter.Key,
		Value: parameter.Value,
	}
}

func ParameterFromRaft(raftParameter *store.Parameter) *types.Parameter {
	return &types.Parameter{
		Key:   raftParameter.Key,
		Value: raftParameter.Value,
	}
}

func PortMappingToRaft(portMapping *types.PortMapping) *store.PortMapping {
	return &store.PortMapping{
		ContainerPort: portMapping.ContainerPort,
		HostPort:      portMapping.HostPort,
		Name:          portMapping.Name,
		Protocol:      portMapping.Protocol,
	}
}

func PortMappingFromRaft(raftPortMapping *store.PortMapping) *types.PortMapping {
	return &types.PortMapping{
		ContainerPort: raftPortMapping.ContainerPort,
		HostPort:      raftPortMapping.HostPort,
		Name:          raftPortMapping.Name,
		Protocol:      raftPortMapping.Protocol,
	}
}

func VolumeToRaft(volume *types.Volume) *store.Volume {
	return &store.Volume{
		ContainerPath: volume.ContainerPath,
		HostPath:      volume.HostPath,
		Mode:          volume.Mode,
	}
}

func VolumeFromFaft(raftVolume *store.Volume) *types.Volume {
	return &types.Volume{
		ContainerPath: raftVolume.ContainerPath,
		HostPath:      raftVolume.HostPath,
		Mode:          raftVolume.Mode,
	}
}

func KillPolicyToRaft(killPolicy *types.KillPolicy) *store.KillPolicy {
	return &store.KillPolicy{
		Duration: killPolicy.Duration,
	}
}

func KillPolicyFromRaft(raftKillPolicy *store.KillPolicy) *types.KillPolicy {
	return &types.KillPolicy{
		Duration: raftKillPolicy.Duration,
	}
}

func UpdatePolicyToRaft(updatePolicy *types.UpdatePolicy) *store.UpdatePolicy {
	return &store.UpdatePolicy{
		UpdateDelay:  updatePolicy.UpdateDelay,
		MaxRetries:   updatePolicy.MaxRetries,
		MaxFailovers: updatePolicy.MaxFailovers,
		Action:       updatePolicy.Action,
	}
}

func UpdatePolicyFromRaft(raftUpdatePolicy *store.UpdatePolicy) *types.UpdatePolicy {
	return &types.UpdatePolicy{
		UpdateDelay:  raftUpdatePolicy.UpdateDelay,
		MaxRetries:   raftUpdatePolicy.MaxRetries,
		MaxFailovers: raftUpdatePolicy.MaxFailovers,
		Action:       raftUpdatePolicy.Action,
	}
}

func HealthCheckToRaft(healthCheck *types.HealthCheck) *store.HealthCheck {
	raftHealthCheck := &store.HealthCheck{
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

	return raftHealthCheck
}

func HealthCheckFromRaft(raftHealthCheck *store.HealthCheck) *types.HealthCheck {
	healthCheck := &types.HealthCheck{
		ID:                  raftHealthCheck.ID,
		Address:             raftHealthCheck.Address,
		Protocol:            raftHealthCheck.Protocol,
		PortName:            raftHealthCheck.PortName,
		Path:                raftHealthCheck.Path,
		Value:               raftHealthCheck.Value,
		ConsecutiveFailures: raftHealthCheck.ConsecutiveFailures,
		GracePeriodSeconds:  raftHealthCheck.GracePeriodSeconds,
		IntervalSeconds:     raftHealthCheck.IntervalSeconds,
		TimeoutSeconds:      raftHealthCheck.TimeoutSeconds,
		DelaySeconds:        raftHealthCheck.DelaySeconds,
	}

	return healthCheck
}

func SlotToRaft(slot *Slot) *store.Slot {
	raftSlot := &store.Slot{
		Index:     int32(slot.Index),
		ID:        slot.ID,
		AppID:     slot.App.ID,
		VersionID: slot.Version.ID,
		Healthy:   slot.Healthy(),
		State:     slot.State,
		Weight:    slot.GetWeight(),
	}

	if slot.CurrentTask != nil {
		raftSlot.CurrentTask = TaskToRaft(slot.CurrentTask)
	}

	if len(slot.TaskHistory) > 0 {
		raftSlot.TaskHistory = make([]*store.Task, 0)
		for _, t := range slot.TaskHistory {
			raftSlot.TaskHistory = append(raftSlot.TaskHistory, TaskToRaft(t))
		}
	}

	return raftSlot
}

func SlotFromRaft(raftSlot *store.Slot, app *App) *Slot {
	slot := &Slot{
		Index:         int(raftSlot.Index),
		ID:            raftSlot.ID,
		State:         raftSlot.State,
		OfferID:       raftSlot.CurrentTask.OfferID,
		AgentID:       raftSlot.CurrentTask.AgentID,
		Ip:            raftSlot.CurrentTask.Ip,
		AgentHostName: raftSlot.CurrentTask.AgentHostName,
		healthy:       raftSlot.Healthy,
		weight:        raftSlot.Weight,
		TaskHistory:   make([]*Task, 0),
	}

	if raftSlot.CurrentTask != nil {
		slot.CurrentTask = TaskFromRaft(raftSlot.CurrentTask, app)
		slot.CurrentTask.Slot = slot
	}

	if len(raftSlot.TaskHistory) > 0 {
		for _, t := range raftSlot.TaskHistory {
			slot.TaskHistory = append(slot.TaskHistory, TaskFromRaft(t, app))
		}
	}

	for _, version := range app.Versions {
		if raftSlot.VersionID == version.ID {
			slot.Version = version
		}
	}

	return slot
}

func TaskToRaft(task *Task) *store.Task {
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

func TaskFromRaft(raftTask *store.Task, app *App) *Task {
	task := &Task{
		ID:            raftTask.ID,
		State:         raftTask.State,
		Stdout:        raftTask.Stdout,
		Stderr:        raftTask.Stderr,
		HostPorts:     raftTask.HostPorts,
		OfferID:       raftTask.OfferID,
		AgentID:       raftTask.AgentID,
		Ip:            raftTask.Ip,
		AgentHostName: raftTask.AgentHostName,
		Reason:        raftTask.Reason,
		Message:       raftTask.Message,
		Created:       time.Unix(0, raftTask.CreatedAt),
		ContainerId:   raftTask.ContainerId,
		ContainerName: raftTask.ContainerName,
	}

	for _, version := range app.Versions {
		if raftTask.VersionID == version.ID {
			task.Version = version
		}
	}

	return task
}

func OfferAllocatorItemToRaft(slotID string, offerInfo *OfferInfo) *store.OfferAllocatorItem {
	item := &store.OfferAllocatorItem{
		OfferID:  offerInfo.OfferID,
		SlotID:   slotID,
		Hostname: offerInfo.Hostname,
		AgentID:  offerInfo.AgentID,
	}

	return item
}

func OfferAllocatorItemFromRaft(item *store.OfferAllocatorItem) (slotID string, offerInfo *OfferInfo) {
	return item.SlotID, &OfferInfo{
		OfferID:  item.OfferID,
		Hostname: item.Hostname,
		AgentID:  item.AgentID,
	}
}

func StateMachineFromRaft(app *App, machine *store.StateMachine) *StateMachine {
	return &StateMachine{
		state: StateFromRaft(app, machine.State),
	}
}

func StateFromRaft(app *App, state *store.State) State {
	slot, _ := app.GetSlot(int(state.CurrentSlotIndex))
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

func StateMachineToRaft(machine *StateMachine) *store.StateMachine {
	return &store.StateMachine{
		State: StateToRaft(machine.CurrentState()),
	}
}

func StateToRaft(state State) *store.State {
	stateObj := reflect.ValueOf(state).Elem()
	typeOfT := stateObj.Type()
	raftState := &store.State{}

	for i := 0; i < stateObj.NumField(); i++ {
		f := stateObj.Field(i)
		switch typeOfT.Field(i).Name {
		case "Name":
			raftState.Name = f.Interface().(string)
		case "CurrentSlotIndex":
			raftState.CurrentSlotIndex = int64(f.Interface().(int))
		case "TargetSlotIndex":
			raftState.TargetSlotIndex = int64(f.Interface().(int))
		case "SlotCountNeedUpdate":
			raftState.SlotCountNeedUpdate = int64(f.Interface().(int))
		default:
		}
	}

	return raftState
}
