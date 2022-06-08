package tell

import (
	"time"
)

type (
	// Notifier is an interface that allows applications to subscribe to, and
	// publish events.
	Notifier interface {
		Subscribe(event string, handler func(Event))
		Start(eventName string, payload Payload) *InFlightEvent
	}

	// SimpleNotifier is a simple event notifier that can be used to hook into
	// framework events. This allows consumers of frameworks and libraries to
	// implement logging, tracing, and other functionality based on events that
	// occur in the system.
	SimpleNotifier struct {
		subscriptions map[string][]func(Event)
		// MaxWait is the maximum amount of time to wait for a subscriber to
		// finish processing an event. The default is 5 seconds.
		MaxWait time.Duration
	}

	// Payload is a map of key/value pairs that can be used to pass data along
	// to subscribers. This type is provided for convenience.
	Payload = map[string]any

	// Event is passed to subscribers when an event is published.
	Event struct {
		Name     string
		Payload  map[string]any
		Start    time.Time
		maxWait  time.Duration
		wait     chan struct{}
		inFlight *InFlightEvent
	}

	// InFlightEvent is an event that has been started but has not yet
	// completed.
	InFlightEvent struct {
		published bool
		finished  chan struct{}
	}
)

// New returns a new notifier instance. Applications should typically have a
// single notifier that is passed to each component as needed.
func New() *SimpleNotifier {
	return &SimpleNotifier{subscriptions: make(map[string][]func(Event)), MaxWait: time.Second * 5}
}

// Subscribe adds a handler to the list of handlers for the given event.
func (n *SimpleNotifier) Subscribe(event string, handler func(Event)) {
	n.subscriptions[event] = append(n.subscriptions[event], handler)
}

// Start publishes an event for the given eventName and passes in payload.
// The event will not complete until Finish is called.
func (n *SimpleNotifier) Start(eventName string, payload Payload) *InFlightEvent {
	inFlight := &InFlightEvent{
		published: false,
		finished:  make(chan struct{}),
	}

	// Handle empty payloads and allow them to be accessed instead of panicing
	if payload == nil {
		payload = make(Payload)
	}

	event := Event{
		Name:     eventName,
		Start:    time.Now(),
		Payload:  payload,
		maxWait:  n.MaxWait,
		inFlight: inFlight,
		wait:     make(chan struct{}),
	}

	n.publishEvent(event)

	return inFlight
}

func (n *SimpleNotifier) publishEvent(e Event) {
	if _, ok := n.subscriptions[e.Name]; !ok {
		return
	}

	for _, sub := range n.subscriptions[e.Name] {
		// Call each subscriber, waiting for the to complete or call e.Wait()
		go func() {
			// Close e.wait in case there is no e.Wait() call
			defer close(e.wait)
			sub(e)
		}()
		<-e.wait
	}
}

// Finish marks the event as finished and allows subscribers to continue by
// unblocking Event.Wait
func (e *InFlightEvent) Finish() {
	if e.published {
		return
	}

	e.published = true
	close(e.finished)
}

// Allows subscribers to wait for the event to finish. This is useful for
// use-cases like measuring execution time and tracing.
func (e *Event) Wait() {
	if e.inFlight.published {
		return
	}

	// Signal that the event is now waiting for the subscriber to finish so
	// other subscribers can be called.
	select {
	case e.wait <- struct{}{}:
	default:
	}

	timer := time.NewTimer(e.maxWait)

	select {
	case <-e.inFlight.finished:
		timer.Stop()
	case <-timer.C:
		return
	}
}
