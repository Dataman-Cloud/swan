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
	MaxRestryTimes        int

	currentInterval time.Duration
	checkTimer      *time.Timer
	restartFunc     TestAndRestartFunc

	slot *Slot

	retriedTimes int
}

func NewRestartPolicy(slot *Slot, BackoffSeconds time.Duration, BackoffFactor float64,
	MaxLaunchDelaySeconds time.Duration, retryTimes int, restartFunc TestAndRestartFunc) *RestartPolicy {

	p := &RestartPolicy{
		BackoffSeconds:        BackoffSeconds,
		BackoffFactor:         BackoffFactor,
		MaxLaunchDelaySeconds: MaxLaunchDelaySeconds,
		MaxRestryTimes:        retryTimes,
		currentInterval:       BackoffSeconds,
		slot:                  slot,
		restartFunc:           restartFunc,
		retriedTimes:          0,
	}

	p.checkTimer = time.AfterFunc(p.currentInterval, func() {
		p.restartAndSetNextTimer()
	})

	return p
}

func (rs *RestartPolicy) restartAndSetNextTimer() {
	if rs.retriedTimes >= rs.MaxRestryTimes {
		rs.Stop()
		return
	}

	logrus.Debugf("call restartFunc now for %s", rs.slot.ID)

	stopCheck := rs.restartFunc(rs.slot)
	rs.currentInterval = time.Duration(rs.currentInterval.Seconds()*rs.BackoffFactor) * time.Second
	if !stopCheck && rs.currentInterval < rs.MaxLaunchDelaySeconds {
		rs.checkTimer = time.AfterFunc(rs.currentInterval, rs.restartAndSetNextTimer)
	}

	rs.retriedTimes++
}

func (rs *RestartPolicy) Stop() {
	if rs.checkTimer != nil {
		logrus.Infof("stop RestartPolicy timer for slot %s", rs.slot.ID)
		rs.checkTimer.Stop()
	}
}
