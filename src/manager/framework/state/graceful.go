package state

import (
	"time"

	"github.com/Sirupsen/logrus"
)

type TestAndRestartFunc func(slot *Slot) bool

type RestartPolicy struct {
	BackoffSeconds        time.Duration
	BackoffFactor         float64
	MaxLaunchDelaySeconds time.Duration

	currentInterval time.Duration
	checkTimer      *time.Timer
	restartFunc     TestAndRestartFunc

	slot *Slot
}

func NewRestartPolicy(slot *Slot, BackoffSeconds time.Duration, BackoffFactor float64,
	MaxLaunchDelaySeconds time.Duration, restartFunc TestAndRestartFunc) *RestartPolicy {

	p := &RestartPolicy{
		BackoffSeconds:        BackoffSeconds,
		BackoffFactor:         BackoffFactor,
		MaxLaunchDelaySeconds: MaxLaunchDelaySeconds,
		currentInterval:       BackoffSeconds,
		slot:                  slot,
		restartFunc:           restartFunc,
	}

	p.checkTimer = time.AfterFunc(p.currentInterval, func() {
		p.restartAndSetNextTimer()
	})

	return p
}

func (rs *RestartPolicy) restartAndSetNextTimer() {
	logrus.Debug("call restartFunc now")

	stopCheck := rs.restartFunc(rs.slot)
	rs.currentInterval = time.Duration(rs.currentInterval.Seconds()*rs.BackoffFactor) * time.Second
	if !stopCheck && rs.currentInterval < rs.MaxLaunchDelaySeconds {
		rs.checkTimer = time.AfterFunc(rs.currentInterval, rs.restartAndSetNextTimer)
	}
}

func (rs *RestartPolicy) Stop() {
	if rs.checkTimer != nil {
		logrus.Infof("stop RestartPolicy timer for slot %s", rs.slot.ID)
		rs.checkTimer.Stop()
	}
}
