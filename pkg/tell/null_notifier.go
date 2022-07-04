package tell

type nullNotifier struct{}

var NullNotifier Notifier = &nullNotifier{}

func (n *nullNotifier) Subscribe(event string, handler func(Event)) {}
func (n *nullNotifier) Start(eventName string, payload Payload) *Event {
	return &Event{published: true, onFinish: func() {}, Payload: payload}
}
