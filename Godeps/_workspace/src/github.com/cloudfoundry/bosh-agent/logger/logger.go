package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime/debug"
	"strings"
)

type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelNone LogLevel = 99
)

var levels = map[string]LogLevel{
	"DEBUG": LevelDebug,
	"INFO":  LevelInfo,
	"WARN":  LevelWarn,
	"ERROR": LevelError,
	"NONE":  LevelNone,
}
var levelKeys = []string{"DEBUG", "INFO", "WARN", "ERROR", "NONE"}

func Levelify(levelString string) (LogLevel, error) {
	upperLevelString := strings.ToUpper(levelString)
	level, ok := levels[upperLevelString]
	if !ok {
		expected := strings.Join(levelKeys, ", ")
		return level, fmt.Errorf("Unknown LogLevel string '%s', expected one of [%s]", levelString, expected)
	}
	return level, nil
}

type Logger struct {
	level LogLevel
	out   *log.Logger
	err   *log.Logger
}

func New(level LogLevel, out, err *log.Logger) Logger {
	return Logger{
		level: level,
		out:   out,
		err:   err,
	}
}

func NewLogger(level LogLevel) Logger {
	return NewWriterLogger(level, os.Stdout, os.Stderr)
}

func NewWriterLogger(level LogLevel, out, err io.Writer) Logger {
	return Logger{
		level: level,
		out:   log.New(out, "", log.LstdFlags),
		err:   log.New(err, "", log.LstdFlags),
	}
}

func (l Logger) Debug(tag, msg string, args ...interface{}) {
	if l.level > LevelDebug {
		return
	}

	msg = fmt.Sprintf("DEBUG - %s", msg)
	l.getOutLogger(tag).Printf(msg, args...)
}

// DebugWithDetails will automatically change the format of the message
// to insert a block of text after the log
func (l Logger) DebugWithDetails(tag, msg string, args ...interface{}) {
	msg = msg + "\n********************\n%s\n********************"
	l.Debug(tag, msg, args...)
}

func (l Logger) Info(tag, msg string, args ...interface{}) {
	if l.level > LevelInfo {
		return
	}

	msg = fmt.Sprintf("INFO - %s", msg)
	l.getOutLogger(tag).Printf(msg, args...)
}

func (l Logger) Warn(tag, msg string, args ...interface{}) {
	if l.level > LevelWarn {
		return
	}

	msg = fmt.Sprintf("WARN - %s", msg)
	l.getErrLogger(tag).Printf(msg, args...)
}

func (l Logger) Error(tag, msg string, args ...interface{}) {
	if l.level > LevelError {
		return
	}

	msg = fmt.Sprintf("ERROR - %s", msg)
	l.getErrLogger(tag).Printf(msg, args...)
}

// ErrorWithDetails will automatically change the format of the message
// to insert a block of text after the log
func (l Logger) ErrorWithDetails(tag, msg string, args ...interface{}) {
	msg = msg + "\n********************\n%s\n********************"
	l.Error(tag, msg, args...)
}

func (l Logger) HandlePanic(tag string) {
	panic := recover()

	if panic != nil {
		var msg string

		switch obj := panic.(type) {
		case string:
			msg = obj
		case fmt.Stringer:
			msg = obj.String()
		case error:
			msg = obj.Error()
		default:
			msg = fmt.Sprintf("%#v", obj)
		}

		l.ErrorWithDetails(tag, "Panic: %s", msg, debug.Stack())
		os.Exit(2)
	}
}

func (l Logger) getOutLogger(tag string) (logger *log.Logger) {
	return l.updateLogger(l.out, tag)
}

func (l Logger) getErrLogger(tag string) (logger *log.Logger) {
	return l.updateLogger(l.err, tag)
}

func (l Logger) updateLogger(logger *log.Logger, tag string) *log.Logger {
	prefix := fmt.Sprintf("[%s] ", tag)
	logger.SetPrefix(prefix)
	return logger
}
