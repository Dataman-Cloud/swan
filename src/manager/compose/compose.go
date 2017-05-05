package compose

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/Dataman-Cloud/swan/src/manager/scheduler"
	"github.com/Dataman-Cloud/swan/src/manager/state"
	"github.com/Dataman-Cloud/swan/src/manager/store"
	"github.com/Sirupsen/logrus"
)

type InsErr struct {
	Operation string `json:"operation"`
	Service   string `json:"service"`
	ErrMsg    string `json:"errmsg"`
}

func (ie InsErr) Error() string {
	bs, _ := json.Marshal(ie)
	return string(bs)
}

func LaunchInstance(ins *store.Instance, sched *scheduler.Scheduler, jid string) (err error) {
	var (
		op = "launch"
	)

	defer func() {
		if err != nil {
			logrus.Errorln(err)
		}
	}()

	svrOrders := ins.ServiceGroup.PrioritySort()
	logrus.Printf("deploy instance with order: %v", svrOrders)

	for _, svr := range svrOrders {
		sp := newSvrPack(ins, sched, svr, jid)
		if err := sp.create(); err != nil {
			return InsErr{op, svr, err.Error()}
		}
	}

	return nil
}

//
// svrPack is just a convinient hand for instance service(app) operation
//
type svrPack struct {
	ins   *store.Instance
	svr   *store.DockerService
	sched *scheduler.Scheduler
	jid   string
}

func newSvrPack(ins *store.Instance, sched *scheduler.Scheduler, svr, jid string) *svrPack {
	return &svrPack{
		ins:   ins,
		svr:   ins.ServiceGroup[svr],
		sched: sched,
		jid:   jid,
	}
}

func (sp *svrPack) insName() string {
	return sp.ins.Name
}

func (sp *svrPack) svrName() string {
	return sp.svr.Name
}

func (sp *svrPack) create() error {
	ver, _ := SvrToVersion(sp.svr, sp.insName())

	logrus.Println("----------------> create app:", sp.insName(), sp.svrName())
	json.NewEncoder(os.Stdout).Encode(ver)
	logrus.Println("<----------------")

	app, err := sp.sched.CreateApp(ver, sp.insName())
	if err != nil {
		return err
	}

	// hanging wait for app creation until normal or timeout
	var (
		timeOut  = time.Minute * 10 // TODO user define
		interval = time.Millisecond * 300
		goesBy   float64
	)
	for {
		app, err = sp.sched.InspectApp(app.ID)
		if err != nil {
			return err
		}
		if app.StateMachine.ReadableState() == state.APP_STATE_NORMAL {
			break
		}

		time.Sleep(interval)
		goesBy += interval.Seconds()
		if goesBy > timeOut.Seconds() {
			return errors.New("time out")
		}
	}

	// wait delay
	time.Sleep(time.Second * time.Duration(sp.svr.Extra.WaitDelay))
	return nil
}
