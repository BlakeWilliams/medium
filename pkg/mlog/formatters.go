package mlog

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
)

type (
	Formatter interface {
		Format(level string, msg string, fields Fields) []byte
	}

	PrettyFormatter struct {
		Color bool
	}

	jsonFormatter struct {
	}
)

var _ Formatter = (*PrettyFormatter)(nil)

func (pf PrettyFormatter) Format(level string, msg string, fields Fields) []byte {
	levelColor := colorForLevel(level)

	if !pf.Color {
		levelColor.DisableColor()
	}

	formatString := []string{
		levelColor.Add(color.Bold).Sprintf("[%s]", strings.ToUpper(level)),
		msg,
	}

	for key, value := range fields {
		formattedField := fmt.Sprintf("%s=%v", levelColor.Sprint(key), value)
		formatString = append(formatString, formattedField)
	}

	return []byte(strings.Join(formatString, " "))
}

var _ Formatter = (*jsonFormatter)(nil)

func (jf jsonFormatter) Format(level string, msg string, fields Fields) []byte {
	// Ensure the core values are encoded too
	fields["msg"] = msg
	fields["level"] = level
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

func JSONFormatter() Formatter {
	return jsonFormatter{}
}

func colorForLevel(level string) *color.Color {
	switch level {
	case "info":
		return color.New(color.FgCyan)
	case "debug":
		return color.New(color.FgGreen)
	case "warn":
		return color.New(color.FgYellow)
	case "error":
		return color.New(color.FgRed)
	default:
		return color.New()
	}
}
