package state

import (
	rafttypes "github.com/Dataman-Cloud/swan/src/manager/raft/types"
	"github.com/Dataman-Cloud/swan/src/types"
)

func AppToRaft(app *App) *rafttypes.Application {
	raftApp := &rafttypes.Application{
		ID:        app.AppId,
		CreatedAt: app.Created.UnixNano(),
		UpdatedAt: app.Updated.UnixNano(),
		State:     app.State,
	}

	if app.CurrentVersion != nil {
		raftApp.Version = VersionToRaft(app.CurrentVersion)
	}

	return raftApp
}

func VersionToRaft(version *types.Version) *rafttypes.Version {
	raftVersion := &rafttypes.Version{
		ID:          version.ID,
		Command:     version.Command,
		Cpus:        version.Cpus,
		Mem:         version.Mem,
		Disk:        version.Disk,
		Instances:   version.Instances,
		RunAs:       version.RunAs,
		Labels:      version.Labels,
		Env:         version.Env,
		Constraints: version.Constraints,
		Uris:        version.Uris,
		Ip:          version.Ip,
		Mode:        version.Mode,
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

	if version.HealthChecks != nil {
		var healthChecks []*rafttypes.HealthCheck
		for _, healthCheck := range version.HealthChecks {
			healthChecks = append(healthChecks, HealthCheckToRaft(healthCheck))
		}
		raftVersion.HealthChecks = healthChecks
	}

	return raftVersion
}

func VersionFromRaft() {
	version := &types.Version{
		ID:          raftVersion.ID,
		Command:     raftVersion.Command,
		Cpus:        raftVersion.Cpus,
		Mem:         raftVersion.Mem,
		Disk:        raftVersion.Disk,
		Instances:   raftVersion.Instances,
		RunAs:       raftVersion.RunAs,
		Labels:      raftVersion.Labels,
		Env:         raftVersion.Env,
		Constraints: raftVersion.Constraints,
		Uris:        raftVersion.Uris,
		Ip:          raftVersion.Ip,
		Mode:        raftVersion.Mode,
	}

	if raftVersion.Container != nil {
		version.Container = ContainerFromContainer(raftVersion.Container)
	}

	if raftVersion.KillPolicy != nil {
		version.KillPolicy = KillPolicyFromRaft(raftVersion.Container)
	}

	if raftVersion.UpdatePolicy != nil {
		version.UpdatePolicy = UpdatePolicyFromRaft(raftVersion.UpdatePolicy)
	}

	if raftVersion.HealthChecks != nil {
		var healthChecks []*types.HealthCheck
		for _, healthCheck := range raftVersion.HealthChecks {
			healthChecks = append(healthChecks, HealthCheckFromRaft(healthCheck))
		}

		version.HealthChecks = healthChecks
	}

	return version
}

func ContainerToRaft(container *types.Container) *rafttypes.Container {
	raftContainer := &rafttypes.Container{
		Type: container.Type,
	}

	if container.Docker != nil {
		raftContainer.Docker = DockerToRaft(container.Docker)
	}

	if container.Volumes != nil {
		var volumes []*rafttypes.Volume

		for _, volume := range container.Volumes {
			volumes = append(volumes, VolumeToRaft(volume))
		}

		raftContainer.Volumes = volumes
	}

	return raftContainer
}

func ContainerFromContainer(raftContainer *rafttypes.Container) types.Container {
	container := &types.Container{
		Type: raftContainer.Type,
	}

	if raftContainer.Docker != nil {
		contianer.Docker = DockerFromRaft(raftContainer.Docker)
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

func DockerToRaft(docker *types.Docker) *rafttypes.Docker {
	raftDocker := &rafttypes.Docker{
		ForcePullImage: docker.ForcePullImage,
		Image:          docker.Image,
		Network:        docker.Network,
		Privileged:     docker.Privileged,
	}

	if docker.Parameters != nil {
		var parameters []*rafttypes.Parameter
		for _, parameter := range docker.Parameters {
			parameters = append(parameters, ParameterToRaft(parameter))
		}

		raftDocker.Parameters = parameters
	}

	if docker.PortMappings != nil {
		var portMappings []*rafttypes.PortMapping

		for _, portMapping := range docker.PortMappings {
			portMappings = append(portMappings, PortMappingToRaft(portMapping))
		}

		raftDocker.PortMappings = portMappings
	}

	return raftDocker
}

func DockerFromRaft(raftDocker *rafttypes.Docker) *types.Docker {
	docker := &types.Docker{
		ForcePullImage: raftDocker.ForcePullImage,
		Image:          raftDocker.ForcePullImage,
		Network:        raftDockerk.Network,
		Privileged:     raftDocer.Privileged,
	}

	if raftDocker.Parameters != nil {
		var parameters []*types.Parameter
		for _, parameter := range raftDockerk.Parameters {
			parameters = append(parameters, ParameterFromRaft(parameter))
		}

		docker.Parameters = parameters
	}

	if docker.PortMappings != nil {
		var portMappings []*types.PortMapping
		for _, portMapping := range raftDocer.PortMappings {
			portMappings = append(portMappings, PortMapping)
		}

		docker.PortMappings = portMappings
	}

	return docker
}

func ParameterToRaft(parameter *types.Parameter) *rafttypes.Parameter {
	return &rafttypes.Parameter{
		Key:   parameter.Key,
		Value: parameter.Value,
	}
}

func ParameterFromRaft(raftParameter *rafttypes.Parameter) *types.Parameter {
	return &types.Parameter{
		Key:   raftParameter.Key,
		Value: raftParameter.Value,
	}
}

func PortMappingToRaft(portMapping *types.PortMapping) *rafttypes.PortMapping {
	return &rafttypes.PortMapping{
		ContainerPort: portMapping.ContainerPort,
		Name:          portMapping.Name,
		Protocol:      portMapping.Protocol,
	}
}

func PortMappingFromRaft(raftPortMapping *rafttypes.PortMapping) *types.PortMapping {
	return &types.PortMapping{
		ContainerPort: raftPortMapping.ContainerPort,
		Name:          raftPortMapping.Name,
		Protocol:      raftPortMapping.Protocol,
	}
}

func VolumeToRaft(volume *types.Volume) *rafttypes.Volume {
	return &rafttypes.Volume{
		ContainerPath: volume.ContainerPath,
		HostPath:      volume.HostPath,
		Mode:          volume.Mode,
	}
}

func VolumeFromFaft(raftVolume *rafttypes.Volume) *types.Volume {
	return &types.Volume{
		ContainerPath: raftVolume.ContainerPath,
		HostPath:      raftVolume.HostPath,
		Mode:          raftVolume.Mode,
	}
}

func KillPolicyToRaft(killPolicy *types.KillPolicy) *rafttypes.KillPolicy {
	return &rafttypes.KillPolicy{
		Duration: killPolicy.Duration,
	}
}

func KillPolicyFromRaft(raftKillPolicy *rafttypes.KillPolicy) *types.KillPolicy {
	return &types.KillPolicy{
		Duration: raftKillPolicy.Duration,
	}
}

func UpdatePolicyToRaft(updatePolicy *types.UpdatePolicy) *rafttypes.UpdatePolicy {
	return &rafttypes.UpdatePolicy{
		UpdateDelay:  updatePolicy.UpdateDelay,
		MaxRetries:   updatePolicy.MaxRetries,
		MaxFailovers: updatePolicy.MaxFailovers,
		Action:       updatePolicy.Action,
	}
}

func UpdatePolicyFromRaft(raftUpdatePolicy *rafttypes.UpdatePolicy) *types.UpdatePolicy {
	return &types.UpdatePolicy{
		UpdateDelay:  raftUpdatePolicy.UpdateDelay,
		MaxRetries:   raftUpdatePolicy.MaxRetries,
		MaxFailovers: raftUpdatePolicy.MaxFailovers,
		Action:       raftUpdatePolicy.Action,
	}
}

func HealthCheckToRaft(healthCheck *types.HealthCheck) *rafttypes.HealthCheck {
	raftHealthCheck := &rafttypes.HealthCheck{
		ID:        healthCheck.ID,
		Address:   healthCheck.Address,
		Protocol:  healthCheck.Protocol,
		Port:      healthCheck.Port,
		PortIndex: healthCheck.PortIndex,
		PortName:  healthCheck.PortName,
		Path:      healthCheck.Path,
		MaxConsecutiveFailures: healthCheck.MaxConsecutiveFailures,
		GracePeriodSeconds:     healthCheck.GracePeriodSeconds,
		IntervalSeconds:        healthCheck.IntervalSeconds,
		TimeoutSeconds:         healthCheck.TimeoutSeconds,
	}

	if healthCheck.Command != nil {
		raftHealthCheck.Command = CommandToRaft(healthCheck.Command)
	}

	return raftHealthCheck
}

func HeadlthCheckFromRaft(raftHealthCheck *rafttypes.HealthCheck) *types.HealthCheck {
	healthCheck := &types.HealthCheck{
		ID:        raftHealthCheck.ID,
		Address:   raftHealthCheck.Address,
		Protocol:  raftHealthCheck.Protocol,
		Port:      raftHealthCheck.Port,
		PortIndex: raftHealthCheck.PortIndex,
		PortName:  raftHealthCheck.PortName,
		Path:      raftHealthCheck.Path,
		MaxConsecutiveFailures: raftHealthCheck.MaxConsecutiveFailures,
		GracePeriodSeconds:     raftHealthCheck.GracePeriodSeconds,
		IntervalSeconds:        raftHealthCheck.IntervalSeconds,
		TimeoutSeconds:         raftHealthCheck.TimeoutSeconds,
	}

	if raftHealthCheck.Command != nil {
		healthCheck.Command = CommandFromRaft(raftHealthCheck.Command)
	}

	return healthCheck
}

func CommandToRaft(command *types.Command) *rafttypes.Command {
	return &rafttypes.Command{command.Value}
}

func CommandFromRaft(raftCommand *rafttypes.Command) *types.Command {
	return &types.Command{
		Value: raftCommand.Value,
	}
}
