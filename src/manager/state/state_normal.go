package state

import (
	"github.com/Sirupsen/logrus"
)

type StateNormal struct {
	Name string
	App  *App
}

func NewStateNormal(app *App) *StateNormal {
	return &StateNormal{
		App:  app,
		Name: APP_STATE_NORMAL,
	}
}

func (normal *StateNormal) OnEnter() {
	logrus.Debug("state normal OnEnter")

	normal.App.EmitAppEvent(normal.Name)
}

func (normal *StateNormal) OnExit() {
	logrus.Debug("state normal OnExit")
}

func (normal *StateNormal) Step() {
	logrus.Debug("state normal step")
}

func (normal *StateNormal) StateName() string {
	return normal.Name
}

// state machine can transit to any state if current state is normal
func (normal *StateNormal) CanTransitTo(targetState string) bool {
	logrus.Debugf("state normal CanTransitTo %s", targetState)

	return true
}
