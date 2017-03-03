package event

type EventSubscriber interface {
	Write(e *Event) error
	InterestIn(e *Event) bool
	GetKey() string
}
