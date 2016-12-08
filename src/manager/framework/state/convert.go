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

func ParameterToRaft(parameter *types.Parameter) *rafttypes.Parameter {
	return &rafttypes.Parameter{
		Key:   parameter.Key,
		Value: parameter.Value,
	}
}

func PortMappingToRaft(portMapping *types.PortMapping) *rafttypes.PortMapping {
	return &rafttypes.PortMapping{
		ContainerPort: portMapping.ContainerPort,
		Name:          portMapping.Name,
		Protocol:      portMapping.Protocol,
	}
}

func VolumeToRaft(volume *types.Volume) *rafttypes.Volume {
	return &rafttypes.Volume{
		ContainerPath: volume.ContainerPath,
		HostPath:      volume.HostPath,
		Mode:          volume.Mode,
	}
}

func KillPolicyToRaft(killPolicy *types.KillPolicy) *rafttypes.KillPolicy {
	return &rafttypes.KillPolicy{
		Duration: killPolicy.Duration,
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

func CommandToRaft(command *types.Command) *rafttypes.Command {
	return &rafttypes.Command{command.Value}
}
