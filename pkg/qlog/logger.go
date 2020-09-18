package qlog

import (
	"fmt"
	"strings"
)

type Logger interface {
	SetLevel(level Level)
	GetLevel() Level
	Log(level Level, args ...interface{})                 // 记录对应级别的日志
	Logf(level Level, format string, args ...interface{}) // 记录对应级别的日志
	Trace(args ...interface{})                            // 记录 TraceLevel 级别的日志
	Tracef(format string, args ...interface{})            // 格式化并记录 TraceLevel 级别的日志
	Debug(args ...interface{})                            // 记录 DebugLevel 级别的日志
	Debugf(format string, args ...interface{})            // 格式化并记录 DebugLevel 级别的日志
	Info(args ...interface{})                             // 记录 InfoLevel 级别的日志
	Infof(format string, args ...interface{})             // 格式化并记录 InfoLevel 级别的日志
	Warn(args ...interface{})                             // 记录 WarnLevel 级别的日志
	Warnf(format string, args ...interface{})             // 格式化并记录 WarnLevel 级别的日志
	Error(args ...interface{})                            // 记录 ErrorLevel 级别的日志
	Errorf(format string, args ...interface{})            // 格式化并记录 ErrorLevel 级别的日志
	Fatal(args ...interface{})                            // 记录 FatalLevel 级别的日志
	Fatalf(format string, args ...interface{})            // 格式化并记录 FatalLevel 级别的日志
	Panic(args ...interface{})                            // 记录 PanicLevel 级别的日志
	Panicf(format string, args ...interface{})            // 格式化并记录 PanicLevel 级别的日志
	WithField(key string, value interface{}) Logger       // 为日志添加一个上下文数据
	WithFields(fields Fields) Logger                      // 为日志添加多个上下文数据
	WithError(err error) Logger                           // 为日志添加标准错误上下文数据
}

// Fields type, used to pass to `WithFields`.
type Fields map[string]interface{}

// Level type
type Level uint32

const (
	// PanicLevel level, highest level of severity. Logs and then calls panic with the
	// message passed to Debug, Info, ...
	PanicLevel Level = iota
	// FatalLevel level. Logs and then calls `logger.Exit(1)`. It will exit even if the
	// logging level is set to Panic.
	FatalLevel
	// ErrorLevel level. Logs. Used for errors that should definitely be noted.
	// Commonly used for hooks to send errors to an error tracking service.
	ErrorLevel
	// WarnLevel level. Non-critical entries that deserve eyes.
	WarnLevel
	// InfoLevel level. General operational entries about what's going on inside the
	// application.
	InfoLevel
	// DebugLevel level. Usually only enabled when debugging. Very verbose logging.
	DebugLevel
	// TraceLevel level. Designates finer-grained informational events than the Debug.
	TraceLevel
)

func (l Level) String() string {
	switch l {
	case TraceLevel:
		return "trace"
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warning"
	case ErrorLevel:
		return "error"
	case FatalLevel:
		return "fatal"
	case PanicLevel:
		return "panic"
	default:
		return "unknown"
	}
}

// ParseLevel takes a string level and returns the Logrus log level constant.
func ParseLevel(lvl string) (Level, error) {
	switch strings.ToLower(lvl) {
	case "panic":
		return PanicLevel, nil
	case "fatal":
		return FatalLevel, nil
	case "error":
		return ErrorLevel, nil
	case "warn", "warning":
		return WarnLevel, nil
	case "info":
		return InfoLevel, nil
	case "debug":
		return DebugLevel, nil
	case "trace":
		return TraceLevel, nil
	}

	var l Level
	return l, fmt.Errorf("not a valid log Level: %q", lvl)
}
