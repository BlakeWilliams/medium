package tell

import (
	"sync"
	"time"
)

type (
	// Notifier is an interface that allows applications to subscribe to, and
	// publish events.
	Notifier interface {
		Subscribe(event string, handler func(Event))
		Start(eventName string, payload Payload) *Event
	}

	// SimpleNotifier is a simple event notifier that can be used to hook into
	// framework events. This allows consumers of frameworks and libraries to
	// implement logging, tracing, and other functionality based on events that
	// occur in the system.
	SimpleNotifier struct {
		subscriptions map[string][]func(Event)
		mu            sync.RWMutex
	}

	// Payload is a map of key/value pairs that can be used to pass data along
	// to subscribers. This type is provided for convenience.
	Payload = map[string]any

	// Event is passed to subscribers when an event is published.
	Event struct {
		Name       string
		Payload    map[string]any
		StartedAt  time.Time
		FinishedAt time.Time
		published  bool
		onFinish   func()
	}
)

// New returns a new notifier instance. Applications should typically have a
// single notifier that is passed to each component as needed.
func New() *SimpleNotifier {
	return &SimpleNotifier{subscriptions: make(map[string][]func(Event))}
}

// Subscribe adds a handler to the list of handlers for the given event.
func (n *SimpleNotifier) Subscribe(event string, handler func(Event)) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.subscriptions[event] = append(n.subscriptions[event], handler)
}

// Start publishes an event for the given eventName and passes in payload.
// The event will not complete until Finish is called.
func (n *SimpleNotifier) Start(eventName string, payload Payload) *Event {
	// Handle empty payloads and allow them to be accessed instead of panicing
	if payload == nil {
		payload = make(Payload)
	}

	event := &Event{
		Name:      eventName,
		StartedAt: time.Now(),
		Payload:   payload,
	}
	event.onFinish = func() { n.publishEvent(event) }

	return event
}

func (n *SimpleNotifier) publishEvent(e *Event) {
	e.FinishedAt = time.Now()

	n.mu.RLock()
	defer n.mu.RUnlock()

	if _, ok := n.subscriptions[e.Name]; !ok {
		return
	}

	for _, sub := range n.subscriptions[e.Name] {
		sub(*e)
	}
}

// Finish marks the event as finished and emits the event to subscribers.
// Finish will attempt to emit the event only once in scenarios where it's called multiple times, but it's not gauranteed.
func (e *Event) Finish() {
	if e.published {
		return
	}

	e.published = true

	e.onFinish()
}
