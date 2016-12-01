package framework

import (
	"sync"

	"github.com/Dataman-Cloud/swan/src/manager/framework/api"
	"github.com/Dataman-Cloud/swan/src/manager/framework/engine"

	"github.com/Dataman-Cloud/swan/src/util"

	"golang.org/x/net/context"
)

type Framework struct {
	Engine  *engine.Engine
	HttpApi *api.Api

	StopC chan struct{}
}

func New(config util.SwanConfig) (*Framework, error) {
	f := &Framework{
		StopC: make(chan struct{}),
	}
	f.Engine = engine.NewEngine(config)
	f.HttpApi = api.NewApi()

	return f, nil
}

func (f *Framework) Start(ctx context.Context) error {
	apiStopC := make(chan struct{})
	engineStopC := make(chan struct{})

	go func() {
		<-ctx.Done()
		apiStopC <- struct{}{}
		engineStopC <- struct{}{}
		f.StopC <- struct{}{}
	}()

	var err error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		wg.Done()
		err = f.HttpApi.Start()
	}()

	go func() {
		wg.Done()
		err = f.Engine.Start()
	}()

	wg.Wait()
	if err != nil {
		return err
	}

	return nil
}
