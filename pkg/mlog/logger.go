package mlog

import (
	"context"
	"errors"
	"io"
)

type (
	// Represents fields to be logged
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

	Writer interface {
		WriteLog(level string, msg string, fields Fields)
	}
)

var ctxKey = struct{}{}

func New(w io.Writer, formatter Formatter) Logger {
	writer := &writer{out: w}
	return baseLogger{writer: writer, formatter: formatter}
}

func Null() Logger {
	return null{}
}

func Inject(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, ctxKey, logger)
}

func WithDefaults(ctx context.Context, fields Fields) (context.Context, error) {
	ctxValue := ctx.Value(ctxKey)

	if logger, ok := ctxValue.(Logger); ok {
		return Inject(ctx, logger.WithDefaults(fields)), nil
	}

	return ctx, errors.New("no logger exists in context")
}

func Debug(ctx context.Context, msg string, fields Fields) {
	if logger, ok := LoggerFrom(ctx); ok {
		logger.Debug(msg, fields)
	}
}
func Info(ctx context.Context, msg string, fields Fields) {
	if logger, ok := LoggerFrom(ctx); ok {
		logger.Info(msg, fields)
	}
}
func Warn(ctx context.Context, msg string, fields Fields) {
	if logger, ok := LoggerFrom(ctx); ok {
		logger.Warn(msg, fields)
	}
}
func Error(ctx context.Context, msg string, fields Fields) {
	if logger, ok := LoggerFrom(ctx); ok {
		logger.Error(msg, fields)
	}
}

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

func (bl baseLogger) Debug(msg string, fields Fields) { bl.Log("debug", msg, fields) }
func (bl baseLogger) Warn(msg string, fields Fields)  { bl.Log("warn", msg, fields) }
func (bl baseLogger) Error(msg string, fields Fields) { bl.Log("error", msg, fields) }
func (bl baseLogger) Info(msg string, fields Fields)  { bl.Log("info", msg, fields) }
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
