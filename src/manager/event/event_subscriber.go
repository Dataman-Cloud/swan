package event

type EventSubscriber interface {
	Write(e *Event) error
	Subscribe(bus *EventBus) error
	Unsubscribe(bus *EventBus) error
	InterestIn(e *Event) bool
}
