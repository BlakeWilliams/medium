package tell

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPubSub(t *testing.T) {
	notifier := New()

	calls := 0
	notifier.Subscribe("omg", func(e Event) {
		calls += 1
	})

	event := notifier.Start("omg", nil)
	event.Finish()

	event = notifier.Start("omg", nil)
	event.Finish()

	require.Equal(t, 2, calls)
}

func TestPubSub_NilSubscription(t *testing.T) {
	notifier := New()

	event := notifier.Start("omg", nil)
	event.Finish()
}

func TestEvent_Wait(t *testing.T) {
	notifier := New()

	var start time.Time
	var finish time.Time

	done := make(chan struct{})
	notifier.Subscribe("omg", func(e Event) {
		defer close(done)
		start = e.StartedAt
		finish = e.FinishedAt
	})

	event := notifier.Start("omg", nil)
	time.Sleep(time.Millisecond * 50)
	event.Finish()

	<-done

	require.NotZero(t, start)
	require.NotZero(t, finish)

	require.InDelta(t, 50, finish.Sub(start).Milliseconds(), 20)
	require.InDelta(t, 5, time.Since(finish).Milliseconds(), 5)
}

func TestInFlightEvent_FinishMultiple(t *testing.T) {
	notifier := New()

	calls := 0
	notifier.Subscribe("omg", func(e Event) {
		calls++
	})

	event := notifier.Start("omg", nil)
	event.Finish()
	event.Finish()

	require.Equal(t, calls, 1)
}
