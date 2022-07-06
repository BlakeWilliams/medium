package mlog

type null struct{}

func (n null) Debug(msg string, fields Fields)             {}
func (n null) Warn(msg string, fields Fields)              {}
func (n null) Error(msg string, fields Fields)             {}
func (n null) Info(msg string, fields Fields)              {}
func (n null) Log(level string, msg string, fields Fields) {}
func (n null) WithDefaults(fields Fields) Logger           { return n }

var _ Logger = (*null)(nil)
