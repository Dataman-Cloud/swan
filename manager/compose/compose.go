package compose

import (
	"fmt"
	"time"

	"github.com/Dataman-Cloud/swan/manager/mesos"
	"github.com/Dataman-Cloud/swan/types"
	"github.com/Dataman-Cloud/swan/utils/labels"
	"github.com/Sirupsen/logrus"
)

func LaunchInstance(ins *types.Instance, sched *mesos.Scheduler) (err error) {
	defer func() {
		ins.UpdatedAt = time.Now()

		if err != nil {
			ins.Status = "failed"
			ins.ErrMsg = err.Error()
		} else {
			ins.Status = "ready"
			ins.ErrMsg = ""
		}

		memo(ins)
	}()

	svrOrders, err := ins.ServiceGroup.PrioritySort()
	if err != nil {
		return err
	}
	logrus.Printf("launch instance with order: %v", svrOrders)

	for _, svr := range svrOrders {
		sp := newSvrPack(ins, sched, svr)
		if err := sp.create(); err != nil {
			return fmt.Errorf("launch service %s error: %v", svr, err)
		}
	}

	return nil
}

func InstanceApps(sched *mesos.Scheduler, insName string) []*types.Application {
	selector, _ := labels.Parse("DM_INSTANCE_NAME=" + insName)
	return sched.ListApps(types.AppFilterOptions{
		LabelsSelector: selector,
	})
}

// svrPack is just a convinient hand for instance service(app) operation
type svrPack struct {
	ins   *types.Instance
	svr   *types.DockerService
	sched *mesos.Scheduler
}

func newSvrPack(ins *types.Instance, sched *mesos.Scheduler, svr string) *svrPack {
	return &svrPack{
		ins:   ins,
		svr:   ins.ServiceGroup[svr],
		sched: sched,
	}
}

func (sp *svrPack) insName() string {
	return sp.ins.Name
}

func (sp *svrPack) svrName() string {
	return sp.svr.Name
}

func (sp *svrPack) create() error {
	//ver, _ := SvrToVersion(sp.svr, sp.insName(), sp.sched.ClusterName())

	//_, err := sp.sched.CreateApp(ver, sp.insName())
	//if err != nil {
	//	return err
	//}

	// hanging wait for app creation until normal or timeout
	var (
		timeOut  = time.Minute * 10 // TODO user define
		interval = time.Millisecond * 300
		goesBy   float64
	)
	for {
		//app, err = sp.sched.InspectApp(app.ID)
		//if err != nil {
		//	return err
		//}
		//if app.StateMachine.ReadableState() == state.APP_STATE_NORMAL {
		//	break
		//}
		//if app.StateMachine.ReadableState() == state.APP_STATE_FAILED {
		//	return fmt.Errorf("app %s failed", sp.svrName())
		//}

		time.Sleep(interval)
		goesBy += interval.Seconds()
		if goesBy > timeOut.Seconds() {
			return fmt.Errorf("waitting for app %s timeout", sp.svrName())
		}
	}

	// wait delay
	time.Sleep(time.Second * time.Duration(sp.svr.Extra.WaitDelay))
	return nil
}

func memo(ins *types.Instance) error {
	//return store.DB().UpdateInstance(ins)
	return nil
}
