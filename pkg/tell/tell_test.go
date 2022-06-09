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
		calls++
		e.Wait()
	})

	event := notifier.Start("omg", nil)
	event.Finish()

	event = notifier.Start("omg", nil)
	event.Finish()
}

func TestPubSub_NilSubscription(t *testing.T) {
	notifier := New()

	event := notifier.Start("omg", nil)
	event.Finish()
}

func TestEvent_Wait(t *testing.T) {
	notifier := New()

	beforeWait := time.Now().Add(-24 * time.Hour)
	afterWait := time.Now().Add(-24 * time.Hour)

	done := make(chan struct{})
	notifier.Subscribe("omg", func(e Event) {
		defer close(done)
		beforeWait = time.Now()
		e.Wait()
		afterWait = time.Now()
	})

	event := notifier.Start("omg", nil)
	time.Sleep(time.Millisecond * 50)
	event.Finish()

	<-done

	require.InDelta(t, beforeWait.Unix(), time.Now().Unix(), 1)
	require.InDelta(t, afterWait.Unix(), time.Now().Unix(), 1)

	require.InDelta(t, afterWait.Sub(beforeWait).Milliseconds(), 50, 20)
}

func TestEvent_WaitMultiple(t *testing.T) {
	notifier := New()

	done := make(chan struct{})
	notifier.Subscribe("omg", func(e Event) {
		defer close(done)
		e.Wait()
		e.Wait()
		e.Wait()
		e.Wait()
		e.Wait()
	})

	event := notifier.Start("omg", nil)
	time.Sleep(time.Millisecond * 50)
	event.Finish()

	<-done
}

func TestEvent_WaitTimeout(t *testing.T) {
	notifier := New()
	notifier.MaxWait = time.Millisecond * 10

	timer := time.NewTimer(20 * time.Millisecond)
	defer timer.Stop()

	done := make(chan struct{})
	notifier.Subscribe("omg", func(e Event) {
		defer close(done)
		e.Wait()
	})

	notifier.Start("omg", nil)

	select {
	case <-done:
	case <-timer.C:
		require.Fail(t, "event.Wait() did not timeout")
	}
}

func TestInFlightEvent_FinishMultiple(t *testing.T) {
	event := &InFlightEvent{
		published: false,
		finished:  make(chan struct{}),
	}

	event.Finish()
	event.Finish()
}
