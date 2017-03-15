package state

import (
	"errors"
	"fmt"
	"sync"
)

const (
	APP_STATE_NORMAL        = "normal"
	APP_STATE_CREATING      = "creating"
	APP_STATE_DELETING      = "deleting"
	APP_STATE_UPDATING      = "updating"
	APP_STATE_CANCEL_UPDATE = "cancel_update"
	APP_STATE_SCALE_UP      = "scale_up"
	APP_STATE_SCALE_DOWN    = "scale_down"
)

type StateMachine struct {
	state State
	lock  sync.Mutex
}

// default state for a new statemachine is creating
func NewStateMachine() *StateMachine {
	machine := &StateMachine{}

	return machine
}

func (machine *StateMachine) Start(startState State) {
	machine.state = startState
	machine.state.OnEnter()
}

// return the current state of machine in readable format
func (machine *StateMachine) ReadableState() string {
	return machine.state.Name()
}

// test if targetState is changable
func (machine *StateMachine) CanTransitTo(targetStateString string) bool {
	return machine.state.CanTransitTo(targetStateString)
}

// test machine.state is stateExpected
func (machine *StateMachine) Is(stateExpected string) bool {
	return machine.state.Name() == stateExpected
}

// transition from one state to another,  return error if not a valid
// transtion condition
func (machine *StateMachine) TransitTo(targetState State) error {
	if machine.state.CanTransitTo(targetState.Name()) {
		defer machine.lock.Unlock()
		machine.lock.Lock()

		machine.state.OnExit()
		machine.state = targetState
		machine.state.OnEnter()

		return nil
	} else {
		return errors.New(fmt.Sprintf("cann't transit from state: %s to state: %s", machine.state.Name(), targetState.Name()))
	}
}

// move state machine step foward
func (machine *StateMachine) Step() {
	machine.state.Step()
}

type State interface {
	OnEnter()
	OnExit()

	Name() string
	Step()
	CanTransitTo(targetState string) bool
}
