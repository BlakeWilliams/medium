package hooks

import (
	"time"
)

type HookHandler[T any] func(Event[T])

type Event[T any] struct {
	Duration time.Duration
	Payload T
}

type Emitter[T any] struct {
	subscribers []HookHandler[T]
}

func NewEmitter[T any]() *Emitter[T] {
	return &Emitter[T]{
		subscribers: make([]HookHandler[T], 0),
	}
}

func (e *Emitter[T]) Subscribe(subscriber HookHandler[T]) {
	e.subscribers = append(e.subscribers, subscriber)
}

func (e *Emitter[T]) Emit(payload T, fn func()) {
	start := time.Now()

	fn()

	event := Event[T]{Duration: time.Since(start), Payload: payload}
	for _, subscriber := range e.subscribers {
		subscriber(event)
	}
}
