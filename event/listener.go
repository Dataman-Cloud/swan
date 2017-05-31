package event

type EventListener interface {
	Write(e *Event) error
	InterestIn(e *Event) bool
	Key() string
	Wait()
}
