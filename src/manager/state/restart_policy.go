package state

import (
	"time"

	"github.com/Sirupsen/logrus"
)

type TestAndRestartFunc func(slot *Slot) bool

type RestartPolicy struct {
	interval    time.Duration
	maxRetry    int
	retried     int
	restartFunc TestAndRestartFunc
	slot        *Slot
	quitCh      chan struct{}
}

func NewRestartPolicy(slot *Slot, interval time.Duration, maxRetry int, restartFunc TestAndRestartFunc) *RestartPolicy {

	p := &RestartPolicy{
		interval:    interval,
		maxRetry:    maxRetry,
		retried:     0,
		restartFunc: restartFunc,
		slot:        slot,
		quitCh:      make(chan struct{}),
	}

	go func() {
		ticker := time.NewTicker(p.interval)
		logrus.Printf("start slot watcher on %s", slot.ID)

		defer func() {
			ticker.Stop()
			logrus.Printf("stop slot watcher on %s", slot.ID)
		}()

		for {
			select {
			case <-ticker.C:
				restarted := p.restartFunc(p.slot)
				if restarted {
					p.retried++
				} else {
					p.retried = 0
				}
				if p.retried >= p.maxRetry {
					logrus.Println("reached max restart retries, quit")
					if p.slot.App.StateIs(APP_STATE_CREATING) {
						p.slot.App.TransitTo(APP_STATE_FAILED)
					}
					return
				}

			case <-p.quitCh:
				return
			}
		}
	}()

	return p
}

func (p *RestartPolicy) Stop() {
	close(p.quitCh)
}
