package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelNone         LogLevel = 99
	legacyTimeFormat           = "2006/01/02 15:04:05"
	rfc3339TimeFormat          = "2006-01-02T15:04:05.000000000Z"
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

func AsString(level LogLevel) string {
	for k, v := range levels {
		if level == v {
			return k
		}
	}

	return "DEBUG"
}

// to update cd logger && go run github.com/maxbrunsfeld/counterfeiter -generate
// counterfeiter:generate . Logger
type Logger interface {
	Debug(tag, msg string, args ...interface{})
	DebugWithDetails(tag, msg string, args ...interface{})
	Info(tag, msg string, args ...interface{})
	Warn(tag, msg string, args ...interface{})
	Error(tag, msg string, args ...interface{})
	ErrorWithDetails(tag, msg string, args ...interface{})
	HandlePanic(tag string)
	ToggleForcedDebug()
	UseRFC3339Timestamps()
	UseTags(tags []LogTag)
	Flush() error
	FlushTimeout(time.Duration) error
}

type logger struct {
	level           LogLevel
	logger          *log.Logger
	forcedDebug     bool
	loggerMu        sync.Mutex
	timestampFormat string
	tags            []LogTag
}

type LogTag struct {
	Name     string   `json:"name"`
	LogLevel LogLevel `json:"log_level"`
}

func New(level LogLevel, out *log.Logger) Logger {
	out.SetFlags(0)
	return &logger{
		level:           level,
		logger:          out,
		timestampFormat: legacyTimeFormat,
	}
}

func NewLogger(level LogLevel) Logger {
	return NewWriterLogger(level, os.Stderr)
}

func NewWriterLogger(level LogLevel, writer io.Writer) Logger {
	return New(
		level,
		log.New(writer, "", log.LstdFlags),
	)
}

func (l *logger) UseRFC3339Timestamps() {
	l.timestampFormat = rfc3339TimeFormat
}

func (l *logger) UseTags(tags []LogTag) {
	l.tags = tags
}

func (l *logger) Flush() error                       { return nil }
func (l *logger) FlushTimeout(_ time.Duration) error { return nil }

func (l *logger) Debug(tag, msg string, args ...interface{}) {
	if l.getLogLevel(tag) > LevelDebug && !l.forcedDebug {
		return
	}

	msg = "DEBUG - " + msg
	l.printf(tag, msg, args...)
}

// DebugWithDetails will automatically change the format of the message
// to insert a block of text after the log
func (l *logger) DebugWithDetails(tag, msg string, args ...interface{}) {
	msg = msg + "\n********************\n%s\n********************"
	l.Debug(tag, msg, args...)
}

func (l *logger) Info(tag, msg string, args ...interface{}) {
	if l.getLogLevel(tag) > LevelInfo && !l.forcedDebug {
		return
	}

	msg = "INFO - " + msg
	l.printf(tag, msg, args...)
}

func (l *logger) Warn(tag, msg string, args ...interface{}) {
	if l.getLogLevel(tag) > LevelWarn && !l.forcedDebug {
		return
	}

	msg = "WARN - " + msg
	l.printf(tag, msg, args...)
}

func (l *logger) Error(tag, msg string, args ...interface{}) {
	if l.getLogLevel(tag) > LevelError && !l.forcedDebug {
		return
	}

	msg = "ERROR - " + msg
	l.printf(tag, msg, args...)
}

// ErrorWithDetails will automatically change the format of the message
// to insert a block of text after the log
func (l *logger) ErrorWithDetails(tag, msg string, args ...interface{}) {
	msg = msg + "\n********************\n%s\n********************"
	l.Error(tag, msg, args...)
}

func (l *logger) recoverPanic(tag string) (didPanic bool) {
	if e := recover(); e != nil {
		var msg string
		switch obj := e.(type) {
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
		return true
	}
	return false
}

func (l *logger) HandlePanic(tag string) {
	if l.recoverPanic(tag) {
		os.Exit(2)
	}
}

func (l *logger) ToggleForcedDebug() {
	l.forcedDebug = !l.forcedDebug
}

func (l *logger) printf(tag, msg string, args ...interface{}) {
	s := fmt.Sprintf(msg, args...)
	l.loggerMu.Lock()
	timestamp := time.Now().Format(l.timestampFormat)
	l.logger.SetPrefix("[" + tag + "] " + timestamp + " ")
	l.logger.Output(2, s) //nolint:errcheck
	l.loggerMu.Unlock()
}

func (l *logger) getLogLevel(tag string) LogLevel {
	for _, logTag := range l.tags {
		if logTag.Name == tag {
			return logTag.LogLevel
		}
	}
	return l.level
}
