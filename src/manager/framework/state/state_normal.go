package state

import (
	"github.com/Sirupsen/logrus"
)

type StateNormal struct {
	name string
	app  *App
}

func NewStateNormal(app *App) *StateNormal {
	return &StateNormal{
		app:  app,
		name: APP_STATE_NORMAL,
	}
}

func (normal *StateNormal) OnEnter() {
	logrus.Debug("state normal OnEnter")

	normal.app.EmitAppEvent(normal.name)
}

func (normal *StateNormal) OnExit() {
	logrus.Debug("state normal OnExit")
}

func (normal *StateNormal) Step() {
	logrus.Debug("state normal step")
}

func (normal *StateNormal) Name() string {
	return normal.name
}

// state machine can transit to any state if current state is normal
func (normal *StateNormal) CanTransitTo(targetState string) bool {
	logrus.Debugf("state normal CanTransitTo %s", targetState)

	return true
}
