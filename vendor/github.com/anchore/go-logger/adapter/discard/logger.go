package discard

import (
	iface "github.com/anchore/go-logger"
	"io"
)

var _ iface.Logger = (*logger)(nil)
var _ iface.Controller = (*logger)(nil)

type logger struct {
}

func New() iface.Logger {
	return &logger{}
}

func (l *logger) Tracef(format string, args ...interface{}) {
}

func (l *logger) Debugf(format string, args ...interface{}) {}

func (l *logger) Infof(format string, args ...interface{}) {}

func (l *logger) Warnf(format string, args ...interface{}) {}

func (l *logger) Errorf(format string, args ...interface{}) {}

func (l *logger) Trace(args ...interface{}) {}

func (l *logger) Debug(args ...interface{}) {}

func (l *logger) Info(args ...interface{}) {}

func (l *logger) Warn(args ...interface{}) {}

func (l *logger) Error(args ...interface{}) {}

func (l *logger) WithFields(fields ...interface{}) iface.MessageLogger {
	return l
}

func (l *logger) Nested(fields ...interface{}) iface.Logger { return l }

func (l *logger) SetOutput(writer io.Writer) {}

func (l *logger) GetOutput() io.Writer { return nil }
