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
	App   *App
	state State
	lock  sync.Mutex
}

// default state for a new statemachine is creating
func NewStateMachine(app *App) *StateMachine {
	machine := &StateMachine{
		App: app,
	}

	return machine
}

func (machine *StateMachine) Start() {
	machine.state = NewStateCreating(machine)
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
func (machine *StateMachine) TransitTo(targetStateString string, args ...interface{}) error {
	if machine.state.CanTransitTo(targetStateString) {
		defer machine.lock.Unlock()
		machine.lock.Lock()

		machine.state.OnExit()
		machine.state = machine.StateFactory(targetStateString, args...)
		machine.state.OnEnter()

		return nil
	} else {
		return errors.New(fmt.Sprintf("cann't transit from state: %s to state: %s", machine.state.Name(), targetStateString))
	}
}

// move state machine step foward
func (machine *StateMachine) Step() {
	machine.state.Step()
}

func (machine *StateMachine) StateFactory(stateName string, args ...interface{}) State {
	switch stateName {
	case APP_STATE_NORMAL:
		return NewStateNormal(machine)
	case APP_STATE_CREATING:
		return NewStateCreating(machine)
	case APP_STATE_DELETING:
		return NewStateDeleting(machine)
	case APP_STATE_SCALE_UP:
		return NewStateScaleUp(machine)
	case APP_STATE_SCALE_DOWN:
		return NewStateScaleDown(machine)
	case APP_STATE_UPDATING:
		slotCountNeedUpdate, ok := args[0].(int)
		if !ok {
			slotCountNeedUpdate = 1
		}
		fmt.Println(slotCountNeedUpdate)
		return NewStateUpdating(machine, slotCountNeedUpdate)

	case APP_STATE_CANCEL_UPDATE:
		return NewStateCancelUpdate(machine)
	default:
		panic(errors.New("unrecognized state"))
	}
}

type State interface {
	OnEnter()
	OnExit()

	Name() string
	Step()
	CanTransitTo(targetState string) bool
}
