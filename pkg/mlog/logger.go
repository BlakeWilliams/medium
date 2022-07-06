package mlog

import (
	"context"
	"errors"
	"io"
)

type (
	// Represents fields to be logged. Some fields may not be compatible with
	// all formatters. e.g. the JSONFormatter will omit fields that can't be
	// marshaled.
	Fields = map[string]any

	// Represents an object that will log to an io.Writer
	Logger interface {
		Debug(msg string, fields Fields)
		Info(msg string, fields Fields)
		Warn(msg string, fields Fields)
		Error(msg string, fields Fields)
		Log(level string, msg string, fields Fields)
		WithDefaults(Fields) Logger
	}

	baseLogger struct {
		writer        *writer
		defaultFields Fields
		formatter     Formatter
	}
)

var ctxKey = struct{}{}

// Returns a new Logger that formats logs using the given formatter, writing to
// the provided io.Writer
func New(w io.Writer, formatter Formatter) Logger {
	writer := &writer{out: w}
	return baseLogger{writer: writer, formatter: formatter}
}

// Returns a logger that implements the Logger interface but does not write or
// format the provided fields.
func Null() Logger {
	return null{}
}

// Inject returns a new context that can be used with the top-level logger
// functions, such as Debug, Info, Warn, etc.
//
// This is useful for passing a logger throughout an application without having
// to explicitly pass it or write boilerplate context.Value fetching/casting.
func Inject(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, ctxKey, logger)
}

// WithDefaults returns a new context with a logger that includes the provided
// fields for each log call (e.g. Debug, Error, etc)
func WithDefaults(ctx context.Context, fields Fields) (context.Context, error) {
	ctxValue := ctx.Value(ctxKey)

	if logger, ok := ctxValue.(Logger); ok {
		return Inject(ctx, logger.WithDefaults(fields)), nil
	}

	return ctx, errors.New("no logger exists in context")
}

// Prints a debug level message to the logger stored in ctx, if a logger is present.
func Debug(ctx context.Context, msg string, fields Fields) {
	if logger, ok := LoggerFrom(ctx); ok {
		logger.Debug(msg, fields)
	}
}

// Prints a info level message to the logger stored in ctx, if a logger is present.
func Info(ctx context.Context, msg string, fields Fields) {
	if logger, ok := LoggerFrom(ctx); ok {
		logger.Info(msg, fields)
	}
}

// Prints a warn level message to the logger stored in ctx, if a logger is present.
func Warn(ctx context.Context, msg string, fields Fields) {
	if logger, ok := LoggerFrom(ctx); ok {
		logger.Warn(msg, fields)
	}
}

// Prints a error level message to the logger stored in ctx, if a logger is present.
func Error(ctx context.Context, msg string, fields Fields) {
	if logger, ok := LoggerFrom(ctx); ok {
		logger.Error(msg, fields)
	}
}

// Extracts a logger from context. If no logger is present, a null logger is
// returned and the second return value will be false.
func LoggerFrom(ctx context.Context) (Logger, bool) {
	ctxValue := ctx.Value(ctxKey)

	if logger, ok := ctxValue.(Logger); ok {
		return logger, true
	}

	return Null(), false
}

func logFields(ctx context.Context, level string, msg string, fields Fields) {
	ctxValue := ctx.Value(ctxKey)

	if logger, ok := ctxValue.(Logger); ok {
		logger.Log(level, msg, fields)
	}
}

// Prints a debug level log message. The message includes the fields directly
// passed to it and any default field values defined on the logger.
func (bl baseLogger) Debug(msg string, fields Fields) { bl.Log("debug", msg, fields) }

// Prints a warning level log message. The message includes the fields directly
// passed to it and any default field values defined on the logger.
func (bl baseLogger) Warn(msg string, fields Fields) { bl.Log("warn", msg, fields) }

// Prints a error level log message. The message includes the fields directly
// passed to it and any default field values defined on the logger.
func (bl baseLogger) Error(msg string, fields Fields) { bl.Log("error", msg, fields) }

// Prints a info level log message. The message includes the fields directly
// passed to it and any default field values defined on the logger.
func (bl baseLogger) Info(msg string, fields Fields) { bl.Log("info", msg, fields) }

// Prints a log message using the provided level.The message includes the fields directly
// passed to it and any default field values defined on the logger.
func (bl baseLogger) Log(level string, msg string, fields Fields) {
	allFields := make(Fields, len(fields)+len(bl.defaultFields))
	for k, v := range bl.defaultFields {
		allFields[k] = v
	}
	for k, v := range fields {
		allFields[k] = v
	}

	output := bl.formatter.Format(level, msg, allFields)
	bl.writer.Print(output)
}

// Returns a new Logger that will always log the provided fields in subsequent
// log calls.
//
// If called on a logger that already contains default fields, the
// new logger will include the parent logger's default fields in addition to the passed in fields.
func (bl baseLogger) WithDefaults(fields Fields) Logger {
	defaultFields := make(Fields, len(fields)+len(bl.defaultFields))
	for k, v := range bl.defaultFields {
		defaultFields[k] = v
	}
	for k, v := range fields {
		defaultFields[k] = v
	}

	return baseLogger{writer: bl.writer, formatter: bl.formatter, defaultFields: defaultFields}
}
