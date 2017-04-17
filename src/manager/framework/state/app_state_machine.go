package state

import (
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
	sync.Mutex
}

// default state for a new statemachine is creating
func NewStateMachine() *StateMachine {
	return &StateMachine{}
}

func (machine *StateMachine) Start(startState State) {
	machine.state = startState
	machine.state.OnEnter()
}

func (machine *StateMachine) CurrentState() State {
	return machine.state
}

// test if targetState is changable
func (machine *StateMachine) CanTransitTo(targetStateString string) bool {
	return machine.state.CanTransitTo(targetStateString)
}

func (machine *StateMachine) ReadableState() string {
	if machine != nil && machine.state != nil {
		return machine.state.StateName()
	}
	return ""
}

// test machine.state is stateExpected
func (machine *StateMachine) Is(stateExpected string) bool {
	return machine.state.StateName() == stateExpected
}

// transition from one state to another,  return error if not a valid
// transtion condition
func (machine *StateMachine) TransitTo(targetState State) error {
	if machine.state.CanTransitTo(targetState.StateName()) {
		defer machine.Unlock()
		machine.Lock()

		machine.state.OnExit()
		machine.state = targetState
		machine.state.OnEnter()

		return nil
	}

	return fmt.Errorf("can't transit from state: %s to state: %s", machine.state.StateName(), targetState.StateName())
}

// move state machine step foward
func (machine *StateMachine) Step() {
	machine.state.Step()
}

type State interface {
	OnEnter()
	OnExit()

	StateName() string
	Step()
	CanTransitTo(targetState string) bool
}
