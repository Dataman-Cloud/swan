package health

import (
	"fmt"
	"github.com/Dataman-Cloud/swan/registry/consul"
	"github.com/Dataman-Cloud/swan/types"
	"github.com/Sirupsen/logrus"
)

type HealthChecker struct {
	consul      *consul.Consul
	checkers    []Checker
	failedQueue chan types.ReschedulerMsg
}

func NewHealthChecker(consul *consul.Consul, queue chan types.ReschedulerMsg) *HealthChecker {
	return &HealthChecker{
		consul:      consul,
		failedQueue: queue,
	}
}

func (hc *HealthChecker) Init() {
	logrus.Info("Initial health checkers...")
	checks, err := hc.consul.ListChecks()
	if err != nil {
		logrus.Errorf("Initial health checker failed: %s", err)
		return
	}

	if len(checks) == 0 {
		logrus.Info("No checks defined.Skip")
		return
	}

	for _, check := range checks {
		switch check.Protocol {
		case "http", "HTTP":
			logrus.WithFields(logrus.Fields{"protocol": "http",
				"url":         fmt.Sprintf("http://%s:%d", check.Address, check.Port),
				"interval":    check.Interval,
				"timeout":     check.Timeout,
				"maxFailures": check.MaxFailures,
			}).Info("Register health checker")
			hc.checkers = append(hc.checkers, &HTTPChecker{
				Url:         fmt.Sprintf("http://%s:%d", check.Address, check.Port),
				Interval:    check.Interval,
				Timeout:     check.Timeout,
				MaxFailures: check.MaxFailures,
				FailedHandler: func() error {
					return hc.HealthCheckFailedHandler(check, hc.failedQueue)
				},
			})
		case "tcp", "TCP":
			logrus.WithFields(logrus.Fields{"protocol": "tcp",
				"address":     check.Address,
				"port":        check.Port,
				"interval":    check.Interval,
				"timeout":     check.Timeout,
				"maxFailures": check.MaxFailures,
			}).Info("Register health checker")
			hc.checkers = append(hc.checkers, &TCPChecker{
				Addr:        check.Address,
				Port:        check.Port,
				Interval:    check.Interval,
				Timeout:     check.Timeout,
				MaxFailures: check.MaxFailures,
				FailedHandler: func() error {
					return hc.HealthCheckFailedHandler(check, hc.failedQueue)
				},
			})
		}
	}
}

func (hc *HealthChecker) Start() {
	logrus.Info("Begin to start health check")
	for _, checker := range hc.checkers {
		checker.Start()
	}
}
