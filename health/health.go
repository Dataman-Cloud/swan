package health

import (
	"fmt"

	"github.com/Dataman-Cloud/swan/registry/consul"
	"github.com/Dataman-Cloud/swan/types"
	"github.com/Sirupsen/logrus"
	fifo "github.com/foize/go.fifo"
)

type HealthChecker struct {
	consul    *consul.Consul
	checkers  map[string]Checker
	msgQueue  chan types.ReschedulerMsg
	taskQueue *fifo.Queue
}

func NewHealthChecker(consul *consul.Consul, queue chan types.ReschedulerMsg) *HealthChecker {
	return &HealthChecker{
		consul:    consul,
		msgQueue:  queue,
		taskQueue: fifo.NewQueue(),
		checkers:  make(map[string]Checker),
	}
}

func (m *HealthChecker) Init() {
	logrus.Info("Initial health checkers...")
	checks, err := m.consul.ListChecks()
	if err != nil {
		logrus.Errorf("Initial health checker failed: %s", err)
		return
	}

	if len(checks) == 0 {
		logrus.Info("No checks defined.Skip")
		return
	}

	for _, check := range checks {
		m.Add(check)
	}
}

func (m *HealthChecker) Start() {
	go func() {
		for {
			if checker := m.Next(); checker != nil {
				checker.(Checker).Start()
			}
		}
	}()
}

func (m *HealthChecker) Add(check *types.Check) {
	switch check.Protocol {
	case "http", "HTTP":
		logrus.WithFields(logrus.Fields{"protocol": "http",
			"url":         fmt.Sprintf("http://%s:%d", check.Address, check.Port),
			"interval":    check.Interval,
			"timeout":     check.Timeout,
			"maxFailures": check.MaxFailures,
		}).Info("Register health checker")
		checker := &HTTPChecker{
			ID:          check.TaskID,
			Url:         fmt.Sprintf("http://%s:%d", check.Address, check.Port),
			Interval:    check.Interval,
			Timeout:     check.Timeout,
			MaxFailures: check.MaxFailures,
			FailedHandler: func(appId, taskId string) error {
				return m.HealthCheckFailedHandler(appId, taskId)
			},
			AppID:  check.AppID,
			TaskID: check.TaskID,
			quit:   make(chan struct{}),
		}
		m.taskQueue.Add(checker)
		m.checkers[check.TaskID] = checker
	case "tcp", "TCP":
		logrus.WithFields(logrus.Fields{"protocol": "tcp",
			"address":     check.Address,
			"port":        check.Port,
			"interval":    check.Interval,
			"timeout":     check.Timeout,
			"maxFailures": check.MaxFailures,
		}).Info("Register health checker")
		checker := &TCPChecker{
			ID:          check.TaskID,
			Addr:        fmt.Sprintf("%s:%d", check.Address, check.Port),
			Interval:    check.Interval,
			Timeout:     check.Timeout,
			MaxFailures: check.MaxFailures,
			FailedHandler: func(appId, taskId string) error {
				return m.HealthCheckFailedHandler(appId, taskId)
			},
			AppID:  check.AppID,
			TaskID: check.TaskID,
			quit:   make(chan struct{}),
		}
		m.taskQueue.Add(checker)
		m.checkers[check.TaskID] = checker
	}

}

func (m *HealthChecker) Next() (item interface{}) {
	return m.taskQueue.Next()
}

func (m *HealthChecker) StopCheck(id string) {
	if checker, exist := m.checkers[id]; exist {
		logrus.Infof("Remove health check for task %s", id)
		checker.Stop()
	}
}
