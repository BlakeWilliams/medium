package mlog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type logLine struct {
	Bar   int       `json:"bar"`
	Foo   string    `json:"foo"`
	Level string    `json:"level"`
	Msg   string    `json:"msg"`
	Time  time.Time `json:"time"`
	Name  string    `json:"name"`
}

func TestContext(t *testing.T) {
	testcases := map[string]struct {
		Func  func(context.Context, string, Fields)
		Level string
	}{
		"Info":  {Func: Info, Level: "info"},
		"Warn":  {Func: Warn, Level: "warn"},
		"Error": {Func: Error, Level: "error"},
		"Debug": {Func: Debug, Level: "debug"},
	}
	for desc, tc := range testcases {
		t.Run(desc, func(t *testing.T) {
			var buf bytes.Buffer
			logger := New(&buf, JSONFormatter{})
			ctx := context.Background()
			ctx = Inject(ctx, logger)

			// Perform log call
			tc.Func(ctx, "the truth is out there", Fields{"foo": "bar", "bar": 12})

			line := bytes.Split(bytes.TrimRight(buf.Bytes(), "\n"), []byte("\n"))[0]
			require.NotEmpty(t, line)

			var log logLine
			err := json.Unmarshal(line, &log)
			require.NoError(t, err)
			fmt.Println(line)

			require.Equal(t, tc.Level, log.Level)
			require.Equal(t, "the truth is out there", log.Msg)
			require.Equal(t, "bar", log.Foo)
			require.Equal(t, 12, log.Bar)
			require.WithinDuration(t, time.Now(), log.Time, time.Millisecond*10)
		})
	}
}

func TestWithDefaults(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, JSONFormatter{})
	logger = logger.WithDefaults(Fields{"foo": "override me", "name": "Fox Mulder"})

	logger.Info("the truth is out there", Fields{"foo": "bar", "bar": 12})

	var log logLine
	err := json.Unmarshal(buf.Bytes(), &log)
	require.NoError(t, err)

	require.Equal(t, "info", log.Level)
	require.Equal(t, "the truth is out there", log.Msg)
	require.Equal(t, "bar", log.Foo)
	require.Equal(t, "Fox Mulder", log.Name)
	require.Equal(t, 12, log.Bar)
	require.WithinDuration(t, time.Now(), log.Time, time.Millisecond*10)

	buf.Reset()

	// Test WithDefaults does not mutate original logger
	logger.WithDefaults(Fields{"name": "Dana Scully"})

	logger.Info("the truth is out there", Fields{"foo": "bar", "bar": 12})

	err = json.Unmarshal(buf.Bytes(), &log)
	require.NoError(t, err)

	require.Equal(t, "info", log.Level)
	require.Equal(t, "the truth is out there", log.Msg)
	require.Equal(t, "bar", log.Foo)
	require.Equal(t, "Fox Mulder", log.Name)
	require.Equal(t, 12, log.Bar)
	require.WithinDuration(t, time.Now(), log.Time, time.Millisecond*10)
}

func TestWithDefaults_Nested(t *testing.T) {
	ctx := context.Background()
	var buf bytes.Buffer
	logger := New(&buf, JSONFormatter{})
	ctx = Inject(ctx, logger)
	ctx, err := WithDefaults(ctx, Fields{"foo": "First default"})
	require.NoError(t, err)
	ctx, err = WithDefaults(ctx, Fields{"name": "Fox Mulder"})
	require.NoError(t, err)

	Info(ctx, "the truth is out there", Fields{"bar": 12})

	var log logLine
	err = json.Unmarshal(buf.Bytes(), &log)
	require.NoError(t, err)

	require.Equal(t, "info", log.Level)
	require.Equal(t, "the truth is out there", log.Msg)
	require.Equal(t, "First default", log.Foo)
	require.Equal(t, "Fox Mulder", log.Name)
	require.Equal(t, 12, log.Bar)
	require.WithinDuration(t, time.Now(), log.Time, time.Millisecond*10)

	buf.Reset()
}
