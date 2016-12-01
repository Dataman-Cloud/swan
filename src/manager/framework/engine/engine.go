package engine

import (
	"errors"
	"time"

	"github.com/Dataman-Cloud/swan/src/manager/framework/event"
	"github.com/Dataman-Cloud/swan/src/manager/framework/scheduler"
	"github.com/Dataman-Cloud/swan/src/manager/framework/state"
	"github.com/Dataman-Cloud/swan/src/manager/swancontext"
	"github.com/Dataman-Cloud/swan/src/mesosproto/sched"
	"github.com/Dataman-Cloud/swan/src/types"
	"github.com/Dataman-Cloud/swan/src/util"

	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)

type Engine struct {
	scontext         *swancontext.SwanContext
	scheduler        *scheduler.Scheduler
	heartbeater      *time.Ticker
	userEventChan    chan *event.UserEvent
	mesosCallChan    chan *sched.Call
	mesosFailureChan chan error

	stopC chan struct{}

	apps map[string]*state.App
}

func NewEngine(config util.SwanConfig) *Engine {
	engine := &Engine{
		scheduler:     scheduler.NewScheduler(config.Scheduler),
		heartbeater:   time.NewTicker(1 * time.Second),
		userEventChan: make(chan *event.UserEvent, 1024), // TODO
		mesosCallChan: make(chan *sched.Call, 1024),      // 1024 TODO
	}
	return engine
}

// shutdown main engine and related
func (engine *Engine) Stop() error {
	engine.stopC <- struct{}{}
	return nil
}

// revive from crash or rotate from leader change
func (engine *Engine) Start() error {
	engine.Run(context.Background()) // context as a placeholder
	return nil
}

// main loop
func (engine *Engine) Run(ctx context.Context) error {
	if err := engine.scheduler.ConnectToMesosAndAcceptEvent(); err != nil {
		return err
	}

	for {
		select {
		case e := <-engine.scheduler.MesosEventChan:
			logrus.WithFields(logrus.Fields{"mesos": "yes"}).Debugf("event: %s", e)
		case e := <-engine.userEventChan:
			logrus.WithFields(logrus.Fields{"userevent": "yes"}).Debugf("event: %s", e)
		case call := <-engine.mesosCallChan:
			logrus.WithFields(logrus.Fields{"call": "yes"}).Debugf("xx")
			// TODO wrapp this later
			resp, err := engine.scheduler.Send(call)
			if err != nil {
				return err
			}
			if resp.StatusCode != 200 {
				logrus.Errorf("send response not 200")
			}
		case e := <-engine.mesosFailureChan:
			logrus.WithFields(logrus.Fields{"failure": "yes"}).Debugf("%s", e)
		case <-engine.heartbeater.C: // heartbeat timeout for now
			logrus.WithFields(logrus.Fields{"period checker": "yes"}).Debugf("beat")
		case <-engine.stopC:
			logrus.Infof("stopping main engine")
			return nil
		}
	}
}

func (engine *Engine) handlerMesosEvent(event *event.MesosEvent) {
}

func (engine *Engine) CreateApp(version *types.Version) error {
	_, appExists := engine.apps[version.AppId]
	if appExists {
		return errors.New("app already exists")
	}

	app, err := state.NewApp(version)
	if err != nil {
		return err
	}

	engine.apps[version.AppId] = app
	return nil
}

func (engine *Engine) UpdateApp(version *types.Version) error {
	app, appExists := engine.apps[version.AppId]
	if !appExists {
		return errors.New("app not exists")
	}

	if err := app.Update(version); err != nil {
		return err
	}

	return nil
}

func (engine *Engine) InspectApp(appId string) (*state.App, error) {
	app, appExists := engine.apps[appId]
	if !appExists {
		return nil, errors.New("app not exists")
	}
	return app, nil
}

func (engine *Engine) DeleteApp(appId string) error {
	app, appExists := engine.apps[appId]
	if !appExists {
		return errors.New("app not exists")
	}
	return app.Delete()
}
