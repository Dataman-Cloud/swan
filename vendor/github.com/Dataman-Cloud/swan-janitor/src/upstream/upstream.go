package upstream

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/Dataman-Cloud/swan-janitor/src/loadbalance"

	log "github.com/Sirupsen/logrus"
)

type Upstream struct {
	State     *UpstreamState // new|listening|outdated
	StaleMark bool           // mark if the current upstream not inuse anymore

	ServiceName   string `json:"ServiceName"`
	FrontendPort  string // port listen
	FrontendIp    string // ip listen
	FrontendProto string // http|https|tcp

	Targets     []*Target `json:"Target"`
	LoadBalance *loadbalance.RoundRobinLoadBalancer
}

type UpstreamKey struct {
	Proto string
	Ip    string
	Port  string
}

func (uk UpstreamKey) ToString() string {
	return fmt.Sprintf("%s://%s:%s", uk.Proto, uk.Ip, uk.Port)
}

type UpstreamStateEnum string

const (
	STATE_NEW       UpstreamStateEnum = "new"
	STATE_LISTENING UpstreamStateEnum = "listening"
	STATE_OUTDATED  UpstreamStateEnum = "outdated"
	STATE_CHANGED   UpstreamStateEnum = "changed"
)

type UpstreamState struct {
	state    UpstreamStateEnum
	upstream *Upstream
}

func (us UpstreamState) State() UpstreamStateEnum {
	return us.state
}

func NewUpstreamState(u *Upstream, newState UpstreamStateEnum) *UpstreamState {
	return &UpstreamState{
		upstream: u,
		state:    newState,
	}
}

func (us *UpstreamState) Update(newState UpstreamStateEnum) {
	if us.state != newState {
		log.Debugf("change state of upstream <%s> to [%s]", us.upstream.Key(), newState)
		us.state = newState
	}
}

func (u *Upstream) ToString() string {
	targets := []string{}
	for _, t := range u.Targets {
		targets = append(targets, t.ToString())
	}

	return fmt.Sprintf("%s-%s-%s-%s Targets: \n %s", u.ServiceName, u.FrontendProto, u.FrontendIp, u.FrontendPort, strings.Join(targets, "\n  "))
}

func (u *Upstream) Equal(u1 *Upstream) bool {
	fieldsEqual := u.ServiceName == u1.ServiceName &&
		u.FrontendPort == u1.FrontendPort &&
		u.FrontendIp == u1.FrontendIp &&
		u.FrontendProto == u1.FrontendProto

	targetsSizeEqual := len(u.Targets) == len(u1.Targets)

	targetsEqual := true
	uTargets := make([]string, 0)
	for _, t := range u.Targets {
		uTargets = append(uTargets, t.ToString())
	}
	u1Targets := make([]string, 0)
	for _, t := range u1.Targets {
		u1Targets = append(u1Targets, t.ToString())
	}
	for index, targetStr := range uTargets {
		if targetStr != uTargets[index] {
			targetsEqual = false
		}
	}

	return fieldsEqual && targetsSizeEqual && targetsEqual
}

func (u *Upstream) GetTarget(serviceID string) *Target {
	for _, t := range u.Targets {
		if t.ServiceID == serviceID {
			return t
			break
		}
	}
	return nil
}

func (u *Upstream) FieldsEqualButTargetsDiffer(u1 *Upstream) bool {
	fieldsEqual := u.ServiceName == u1.ServiceName &&
		u.FrontendPort == u1.FrontendPort &&
		u.FrontendIp == u1.FrontendIp &&
		u.FrontendProto == u1.FrontendProto

	if !fieldsEqual {
		return false
	}

	targetsSizeEqual := len(u.Targets) == len(u1.Targets)

	if fieldsEqual && !targetsSizeEqual {
		return true
	}

	targetsEqual := true
	uTargets := make([]string, 0)
	u1Targets := make([]string, 0)

	for _, t := range u.Targets {
		uTargets = append(uTargets, t.ToString())
	}
	for _, t := range u1.Targets {
		u1Targets = append(u1Targets, t.ToString())
	}

	sort.Strings(uTargets)
	sort.Strings(u1Targets)

	for index, targetStr := range uTargets {
		if targetStr != u1Targets[index] {
			targetsEqual = false
		}
	}

	return (fieldsEqual && !targetsSizeEqual) || (fieldsEqual && !targetsEqual)
}

func (u *Upstream) FieldsEqual(u1 *Upstream) bool {
	fieldsEqual := u.ServiceName == u1.ServiceName &&
		u.FrontendPort == u1.FrontendPort &&
		u.FrontendIp == u1.FrontendIp &&
		u.FrontendProto == u1.FrontendProto

	return fieldsEqual
}

func (u *Upstream) EntryPointEqual(u1 *Upstream) bool {
	fieldsEqual := u.FrontendPort == u1.FrontendPort &&
		u.FrontendIp == u1.FrontendIp &&
		u.FrontendProto == u1.FrontendProto

	return fieldsEqual
}

func (u *Upstream) SetState(newState UpstreamStateEnum) {
	if u.State == nil {
		u.State = NewUpstreamState(u, newState)
	} else {
		u.State.Update(newState)
	}
}

func (u *Upstream) StateIs(expectState UpstreamStateEnum) bool {
	return u.State.state == expectState
}

func (u *Upstream) Key() UpstreamKey {
	return UpstreamKey{Proto: u.FrontendProto, Ip: u.FrontendIp, Port: u.FrontendPort}
}

func (u *Upstream) Remove(target *Target) {
	index := -1
	for k, v := range u.Targets {
		if v.Equal(target) {
			index = k
			break
		}
	}
	if index >= 0 {
		u.Targets = append(u.Targets[:index], u.Targets[index+1:]...)
	}
}

func (u *Upstream) NextTargetEntry() *url.URL {
	rr := u.LoadBalance
	current := u.Targets[rr.NextIndex]
	rr.NextIndex = (rr.NextIndex + 1) % len(u.Targets)
	return current.Entry()
}
