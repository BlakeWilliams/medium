package mlog

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestJSONFormatter(t *testing.T) {
	formatter := JSONFormatter{}
	output := formatter.Format(LevelFatal, "i want to believe", Fields{"foo": "bar", "bar": 12})

	var log logLine
	err := json.Unmarshal(output, &log)
	require.NoError(t, err)

	require.Equal(t, "fatal", log.Level)
	require.Equal(t, "i want to believe", log.Msg)
	require.Equal(t, "bar", log.Foo)
	require.Equal(t, 12, log.Bar)
	require.WithinDuration(t, time.Now(), log.Time, time.Millisecond*10)
}

func TestJSONFormatter_IgnoreInvalidValues(t *testing.T) {
	formatter := JSONFormatter{}
	badValue := make(chan int)
	output := formatter.Format(LevelFatal, "i want to believe", Fields{"baz": badValue, "bar": 12, "quux": "\"hi\""})
	require.True(t, json.Valid(output), string(output))

	var log logLine
	err := json.Unmarshal(output, &log)
	require.NoError(t, err)

	require.Equal(t, "fatal", log.Level)
	require.Equal(t, "i want to believe", log.Msg)
	require.Equal(t, 12, log.Bar)
	require.WithinDuration(t, time.Now(), log.Time, time.Millisecond*10)
}
