package state

type StateFailed struct {
	Name string
	App  *App
}

func NewStateFailed(app *App) *StateFailed {
	return &StateFailed{
		App:  app,
		Name: APP_STATE_FAILED,
	}
}

func (s *StateFailed) OnEnter() {
	s.App.EmitAppEvent(s.Name)
}

func (s *StateFailed) OnExit()                         {}
func (s *StateFailed) Step()                           {}
func (s *StateFailed) StateName() string               { return s.Name }
func (s *StateFailed) CanTransitTo(target string) bool { return true }
