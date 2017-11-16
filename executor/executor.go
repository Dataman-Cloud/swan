package main

import (
	"errors"
	"os"

	"github.com/Dataman-Cloud/swan/executor/driver"
	"github.com/Dataman-Cloud/swan/executor/example"
	"github.com/Dataman-Cloud/swan/executor/kvm"
	"github.com/Dataman-Cloud/swan/mesosproto"
	log "github.com/Sirupsen/logrus"
)

var (
	ErrNotImplemented       = errors.New("not implemented yet")
	ErrNotSupportedExecutor = errors.New("not supported executor")
)

func init() {
	log.SetOutput(os.Stdout)
}

func main() {
	// init executor
	if len(os.Args) < 2 {
		log.Fatal("require at least one argument, eg: [example|kvm|pod]")
	}

	var (
		executor Executor
		err      error
	)

	switch os.Args[1] {
	case "example":
		executor = example.New()
	case "kvm":
		executor, err = kvm.New()
	case "pod":
		log.Fatal(ErrNotImplemented)
	default:
		log.Fatal(ErrNotSupportedExecutor)
	}

	if err != nil {
		log.Fatal(err)
	}

	// init driver
	driv, err := driver.NewSwanDriver()
	if err != nil {
		log.Fatal(err)
	}

	// install necessary event handlers
	driv.RegisterEventHandlers(mesosproto.ExecEvent_UNKNOWN, []driver.EventHandler{executor.HandleUnknown})
	driv.RegisterEventHandlers(mesosproto.ExecEvent_SUBSCRIBED, []driver.EventHandler{executor.HandleSubscribed})
	driv.RegisterEventHandlers(mesosproto.ExecEvent_LAUNCH, []driver.EventHandler{executor.HandlePreLaunch, executor.HandleLaunch})
	driv.RegisterEventHandlers(mesosproto.ExecEvent_LAUNCH_GROUP, []driver.EventHandler{executor.HandleLaunchGroup})
	driv.RegisterEventHandlers(mesosproto.ExecEvent_KILL, []driver.EventHandler{executor.HandlePreKill, executor.HandleKill})
	driv.RegisterEventHandlers(mesosproto.ExecEvent_ACKNOWLEDGED, []driver.EventHandler{executor.HandleAcknowledged})
	driv.RegisterEventHandlers(mesosproto.ExecEvent_MESSAGE, []driver.EventHandler{executor.HandleMessage})
	driv.RegisterEventHandlers(mesosproto.ExecEvent_ERROR, []driver.EventHandler{executor.HandleError})
	driv.RegisterEventHandlers(mesosproto.ExecEvent_SHUTDOWN, []driver.EventHandler{executor.HandleShutdown})

	if err := driv.Start(); err != nil {
		log.Fatal(err)
	}
}

type Executor interface {
	HandleUnknown(driv driver.Driver, ev *mesosproto.ExecEvent) error

	// Invoked once the executor driver has been able to successfully connect with Mesos Slave
	HandleSubscribed(driv driver.Driver, ev *mesosproto.ExecEvent) error

	HandlePreLaunch(driv driver.Driver, ev *mesosproto.ExecEvent) error
	HandleLaunch(driv driver.Driver, ev *mesosproto.ExecEvent) error

	HandleLaunchGroup(driv driver.Driver, ev *mesosproto.ExecEvent) error

	HandlePreKill(driv driver.Driver, ev *mesosproto.ExecEvent) error
	HandleKill(driv driver.Driver, ev *mesosproto.ExecEvent) error

	HandleAcknowledged(driv driver.Driver, ev *mesosproto.ExecEvent) error

	HandleMessage(driv driver.Driver, ev *mesosproto.ExecEvent) error

	HandleError(driv driver.Driver, ev *mesosproto.ExecEvent) error

	// Invoked when the executor should terminate all of its currently running tasks.
	HandleShutdown(driv driver.Driver, ev *mesosproto.ExecEvent) error

	// Invoked when the executor re-registers with a restarted slave.
	//Reregistered(driver.Driver)

	// Invoked when the executor becomes "disconnected" from the slave
	//Disconnected(driver.Driver)
}
