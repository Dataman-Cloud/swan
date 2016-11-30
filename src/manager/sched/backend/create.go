package backend

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Dataman-Cloud/swan/src/mesosproto/mesos"
	"github.com/Dataman-Cloud/swan/src/types"

	"github.com/Sirupsen/logrus"
)

func (b *Backend) LaunchApplication(version *types.Version) error {
	b.sched.TaskLaunched = 0

	// Set scheduler's status to busy for accepting resource.
	b.sched.Status = "busy"
	// Set scheduler's status back to idle after launch applicaiton.
	defer func() {
		b.sched.Status = "idle"
	}()

	resources := b.sched.BuildResources(version.Cpus, version.Mem, version.Disk)
	offers, err := b.sched.RequestOffers(resources)
	if err != nil {
		logrus.Errorf("Request offers failed: %s", err.Error())
		return err
	}

	pcph := false

	if version.Constraints != nil && len(version.Constraints) != 0 {
		constraints := filterConstraints(version.Constraints)
		for _, constraint := range constraints {
			cons := strings.Split(constraint, ":")
			if cons[1] == "UNIQUE" {
				pcph = true
			}

			if cons[1] == "LIKE" {
				offers = filterOffers(offers, cons[0], "LIKE", cons[2])
			}
		}
	}

	// per-container-per-host
	if pcph {
		return b.perContainerPerHost(offers, resources, version)
	}

	return b.binPack(offers, resources, version)
}

func filterConstraints(constraints []string) []string {
	filteredConstraints := make([]string, 0)
	for _, constraint := range constraints {
		cons := strings.Split(constraint, ":")
		if len(cons) > 3 || len(cons) < 2 {
			logrus.Errorf("Malformed Constraints")
			continue
		}

		if cons[1] != "UNIQUE" && cons[1] != "LIKE" {
			logrus.Errorf("Constraints operator %s not supported", cons[1])
			continue
		}

		if cons[1] == "UNIQUE" && strings.ToLower(cons[0]) != "hostname" {
			logrus.Errorf("Constraints operator UNIQUE only support 'hostname': %s", cons[0])
			continue
		}

		if cons[1] == "LIKE" && len(cons) < 3 {
			logrus.Errorf("Constraints operator LIKE required two operands")
			continue
		}

		filteredConstraints = append(filteredConstraints, constraint)
	}

	return filteredConstraints
}

func filterOffers(offers []*mesos.Offer, vars ...string) []*mesos.Offer {
	filteredOffers := make([]*mesos.Offer, 0)
	switch vars[1] {
	case "LIKE":
		for _, offer := range offers {
			for _, attr := range offer.Attributes {
				var value string
				name := attr.GetName()
				switch attr.GetType() {
				case mesos.Value_SCALAR:
					value = fmt.Sprintf("%d", *attr.GetScalar().Value)
				case mesos.Value_TEXT:
					value = fmt.Sprintf("%s", *attr.GetText().Value)
				default:
					logrus.Errorf("Unsupported attribute value: %s", attr.GetType())
				}

				if name == vars[0] && strings.Contains(value, vars[2]) {
					filteredOffers = append(filteredOffers, offer)
					break
				}

			}
		}
	}

	return filteredOffers
}

func (b *Backend) perContainerPerHost(offers []*mesos.Offer, resources []*mesos.Resource, version *types.Version) error {
	for _, offer := range offers {
		// TODO(pwzgorilla): Should also consider port resource
		cpus, mem, disk := b.sched.OfferedResources(offer)
		if cpus < version.Cpus || mem < version.Mem || disk < version.Disk {
			continue
		}

		task, err := b.sched.BuildTask(offer, version, "")
		if err != nil {
			return fmt.Errorf("Build task failed: %s", err.Error())
		}

		if err := b.store.SaveTask(task); err != nil {
			return fmt.Errorf("Save task failed: %s", err.Error())
		}

		var tasks []*mesos.TaskInfo
		taskInfo := b.sched.BuildTaskInfo(offer, resources, task)
		tasks = append(tasks, taskInfo)

		resp, err := b.sched.LaunchTasks(offer, tasks)
		if err != nil {
			return fmt.Errorf("Launchs task failed: %s", err.Error())
		}

		if resp != nil && resp.StatusCode != http.StatusAccepted {
			return fmt.Errorf("status code %d received", resp.StatusCode)
		}
	}

	return nil
}

func (b *Backend) binPack(offers []*mesos.Offer, resources []*mesos.Resource, version *types.Version) error {
	for _, offer := range offers {
		cpus, mem, disk := b.sched.OfferedResources(offer)
		var tasks []*mesos.TaskInfo
		for b.sched.TaskLaunched < int(version.Instances) &&
			cpus >= version.Cpus &&
			mem >= version.Mem &&
			disk >= version.Disk {
			task, err := b.sched.BuildTask(offer, version, "")
			if err != nil {
				return fmt.Errorf("Build task failed: %s", err.Error())
			}

			if err := b.store.SaveTask(task); err != nil {
				return fmt.Errorf("Save task failed: %s", err.Error())
			}

			taskInfo := b.sched.BuildTaskInfo(offer, resources, task)
			tasks = append(tasks, taskInfo)

			b.sched.TaskLaunched++
			cpus -= version.Cpus
			mem -= version.Mem
			disk -= version.Disk
		}

		if len(tasks) == 0 {
			return fmt.Errorf("Not enough resource")
		}

		resp, err := b.sched.LaunchTasks(offer, tasks)
		if err != nil {
			return fmt.Errorf("Launchs task failed: %s", err.Error())
		}

		if resp != nil && resp.StatusCode != http.StatusAccepted {
			return fmt.Errorf("status code %d received", resp.StatusCode)
		}
	}

	return nil
}
