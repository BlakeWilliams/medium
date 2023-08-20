package mlog

import (
	"context"
	"errors"
	"io"
)

type Level int8

const (
	// Debug level logs, useful for development but too noisy for production.
	LevelDebug Level = iota
	// Info level, the default log level. Useful for standard log messages.
	LevelInfo
	// Warning level, useful for warnings such as deprecation notices.
	LevelWarn
	// Error level, useful for emitting information about errors that occur,
	// such as rescued panics.
	LevelError
	// Fatal level, useful for application breaking errors. This level does not
	// call os.Exit like other packages.
	LevelFatal
	// Null level, useful for testing or disabling logging.
	LevelNull
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
		Fatal(msg string, fields Fields)
		// Returns the current log level
		Level() Level
		WithDefaults(Fields) Logger
	}

	baseLogger struct {
		writer        *writer
		defaultFields Fields
		formatter     Formatter
		level         Level
	}
)

type ctxKey struct{}

// Returns a new Logger that formats logs using the given formatter, writing to
// the provided io.Writer
//
// The level argument allows you to set what types of logs this logger will
// emit. Any log with a level less than the provided level will be ignored.
// See the Level* constants for more details.
func New(w io.Writer, level Level, formatter Formatter) Logger {
	writer := &writer{out: w}
	return baseLogger{writer: writer, level: level, formatter: formatter}
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
	return context.WithValue(ctx, ctxKey{}, logger)
}

// WithDefaults returns a new context with a logger that includes the provided
// fields for each log call (e.g. Debug, Error, etc)
func WithDefaults(ctx context.Context, fields Fields) (context.Context, error) {
	ctxValue := ctx.Value(ctxKey{})

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

// Prints a fatal level message to the logger stored in ctx, if a logger is present. This does not call `os.Exit()`
func Fatal(ctx context.Context, msg string, fields Fields) {
	if logger, ok := LoggerFrom(ctx); ok {
		logger.Error(msg, fields)
	}
}

// Extracts a logger from context. If no logger is present, a null logger is
// returned and the second return value will be false.
func LoggerFrom(ctx context.Context) (Logger, bool) {
	ctxValue := ctx.Value(ctxKey{})

	if logger, ok := ctxValue.(Logger); ok {
		return logger, true
	}

	return Null(), false
}

// Prints a debug level log message. The message includes the fields directly
// passed to it and any default field values defined on the logger.
func (bl baseLogger) Debug(msg string, fields Fields) { bl.log(LevelDebug, msg, fields) }

// Prints a warning level log message. The message includes the fields directly
// passed to it and any default field values defined on the logger.
func (bl baseLogger) Warn(msg string, fields Fields) { bl.log(LevelWarn, msg, fields) }

// Prints a error level log message. The message includes the fields directly
// passed to it and any default field values defined on the logger.
func (bl baseLogger) Error(msg string, fields Fields) { bl.log(LevelError, msg, fields) }

// Prints a info level log message. The message includes the fields directly
// passed to it and any default field values defined on the logger.
func (bl baseLogger) Info(msg string, fields Fields) { bl.log(LevelInfo, msg, fields) }

// Prints a fatal level log message. The message includes the fields directly
// passed to it and any default field values defined on the logger. This does
// not call os.Exit.
func (bl baseLogger) Fatal(msg string, fields Fields) { bl.log(LevelFatal, msg, fields) }

// Prints a log message using the provided level.The message includes the fields directly
// passed to it and any default field values defined on the logger.
func (bl baseLogger) log(level Level, msg string, fields Fields) {
	if level < bl.level {
		return
	}

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

// Returns the current log level
func (bl baseLogger) Level() Level { return bl.level }

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

// Returns the human readable name for the given level
func LevelName(level Level) string {
	switch level {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	case LevelFatal:
		return "fatal"
	case LevelNull:
		return "null"
	default:
		return "unknown"
	}
}
