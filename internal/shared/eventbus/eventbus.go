package eventbus

type EventBus struct{}

func New() *EventBus {
	return &EventBus{}
}
