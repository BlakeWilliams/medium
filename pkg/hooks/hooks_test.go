package hooks

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type TestEvent struct {
	Value string
}

func TestHook(t *testing.T) {
	emitter := NewEmitter[TestEvent]()

	values := make([]string, 0)

	emitter.Subscribe(func(e Event[TestEvent]) {
		values = append(values, "omg:"+e.Payload.Value)
		require.NotNil(t, e.Duration)
	})

	emitter.Subscribe(func(e Event[TestEvent]) {
		values = append(values, "wow:"+e.Payload.Value)
		require.NotNil(t, e.Duration)
	})

	emitter.Emit(TestEvent{Value: "ok"}, func ()  {})
	require.Equal(t, []string{"omg:ok", "wow:ok"}, values)
}
