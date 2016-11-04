package backend

import (
	"net/http"

	"github.com/Dataman-Cloud/swan/mesosproto/mesos"
	"github.com/Dataman-Cloud/swan/types"
	"github.com/Sirupsen/logrus"
)

func (b *Backend) LaunchApplication(version *types.Version) error {
	b.sched.TaskLaunched = 0

	// Set scheduler's status to busy for accepting resource.
	b.sched.Status = "busy"

	go func() {
		resources := b.sched.BuildResources(version.Cpus, version.Mem, version.Disk)
		offers, err := b.sched.RequestOffers(resources)
		if err != nil {
			logrus.Errorf("Request offers failed: %s", err.Error())
		}

		for _, offer := range offers {
			cpus, mem, disk := b.sched.OfferedResources(offer)
			var tasks []*mesos.TaskInfo
			for b.sched.TaskLaunched < version.Instances &&
				cpus >= version.Cpus &&
				mem >= version.Mem &&
				disk >= version.Disk {
				task, err := b.sched.BuildTask(offer, version, "")
				if err != nil {
					logrus.Errorf("Build task failed: %s", err.Error())
					return
				}

				if err := b.store.PutTask(version.AppID, task); err != nil {
					return
				}

				taskInfo := b.sched.BuildTaskInfo(offer, resources, task)
				tasks = append(tasks, taskInfo)

				if len(task.HealthChecks) != 0 {
					if err := b.store.PutHealthcheck(task,
						*taskInfo.Container.Docker.PortMappings[0].HostPort,
						version.ID); err != nil {
					}
					for _, healthCheck := range task.HealthChecks {
						check := types.Check{
							ID:       task.Name,
							Address:  task.AgentHostname,
							Port:     int64(*taskInfo.Container.Docker.PortMappings[0].HostPort),
							TaskID:   task.Name,
							AppID:    version.ID,
							Protocol: healthCheck.Protocol,
							Interval: healthCheck.IntervalSeconds,
							Timeout:  healthCheck.TimeoutSeconds,
						}
						if healthCheck.Command != nil {
							check.Command = healthCheck.Command
						}

						check.Path = healthCheck.Path
						check.MaxFailures = healthCheck.MaxConsecutiveFailures

						b.sched.HealthCheckManager.Add(&check)
					}
				}

				b.sched.TaskLaunched++
				cpus -= version.Cpus
				mem -= version.Mem
				disk -= version.Disk
			}

			resp, err := b.sched.LaunchTasks(offer, tasks)
			if err != nil {
				logrus.Errorf("Launchs task failed: %s", err.Error())
			}

			if resp != nil && resp.StatusCode != http.StatusAccepted {
				logrus.Errorf("status code %d received", resp.StatusCode)
			}
		}

		// Set scheduler's status back to idle after launch applicaiton.
		b.sched.Status = "idle"
	}()

	return nil
}
