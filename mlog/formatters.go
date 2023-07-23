package mlog

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
)

// Formatter is the interface used by Logger to take level, msg, and fields
// and turn them into output for the logger.
type Formatter interface {
	Format(level Level, msg string, fields Fields) []byte
}

// PrettyFormatter is a struct that implements the Formatter interface and
// prints out user readable logs.
type PrettyFormatter struct {
	// Determines if logs should be colored or not.
	Color bool
}

var _ Formatter = (*PrettyFormatter)(nil)

// Formats the provided arguments in a human-friendly format.
func (pf PrettyFormatter) Format(level Level, msg string, fields Fields) []byte {
	levelColor := colorForLevel(level)
	boldLevelColor := colorForLevel(level).Add(color.Bold)

	if !pf.Color {
		levelColor.DisableColor()
	}

	formattedMsg := msg
	if len(msg) < 28 {
		formattedMsg = formattedMsg + strings.Repeat(" ", 28-len(msg))
	}

	formattedLevel := fmt.Sprintf("[%s]", prettyLevelName(level))
	formatString := []string{
		boldLevelColor.Sprint(formattedLevel),
		formattedMsg,
	}

	for key, value := range fields {
		formattedField := fmt.Sprintf("%s=%v", levelColor.Sprint(key), value)
		formatString = append(formatString, formattedField)
	}

	return []byte(strings.Join(formatString, " "))
}

func prettyLevelName(level Level) string {
	switch level {
	case LevelDebug:
		return "DBG"
	case LevelInfo:
		return "INF"
	case LevelWarn:
		return "WRN"
	case LevelError:
		return "ERR"
	case LevelFatal:
		return "FAT"
	default:
		return "UNK"
	}
}

// JSONFormatter implements the Formatter interface and outputs logs in the
// JSON format.
type JSONFormatter struct{}

var _ Formatter = (*JSONFormatter)(nil)

// Formats the provided arguments into JSON. If a value provided by fields
// returns an error when calling json.Marhshal it will be silently omitted from
// the output.
func (jf JSONFormatter) Format(level Level, msg string, fields Fields) []byte {
	// Ensure the core values are encoded too
	fields["msg"] = msg
	fields["level"] = LevelName(level)
	fields["time"] = time.Now()

	validFields := make(map[string]json.RawMessage, len(fields))

	// TODO: support custom serialization for logging via interface?
	for key, value := range fields {
		if value, err := json.Marshal(value); err == nil {
			validFields[key] = value
		}
	}

	output, _ := json.Marshal(validFields)

	return output
}

func colorForLevel(level Level) *color.Color {
	switch level {
	case LevelInfo:
		return color.New(color.FgBlue)
	case LevelDebug:
		return color.New(color.FgGreen)
	case LevelWarn:
		return color.New(color.FgYellow)
	case LevelError, LevelFatal:
		return color.New(color.FgRed)
	default:
		return color.New()
	}
}
