package qlog

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	JsonFormatter = "json"
	TextFormatter = "text"

	errorKey = "error"
)

var (
	errUnknowLevel = fmt.Errorf("got unknown logger level")

	// 所有受支持的日志级别集合（到底层驱动日志级别的映射关系）
	levels = map[Level]logrus.Level{
		PanicLevel: logrus.PanicLevel,
		FatalLevel: logrus.FatalLevel,
		ErrorLevel: logrus.ErrorLevel,
		WarnLevel:  logrus.WarnLevel,
		InfoLevel:  logrus.InfoLevel,
		DebugLevel: logrus.DebugLevel,
		TraceLevel: logrus.TraceLevel,
	}
)

// ----------------------------------------
//  基础日志记录器的日志
// ----------------------------------------
type log struct {
	entry        *logrus.Entry
	depth        int
	reportCaller bool
}

type Option struct {
	Output       io.Writer
	Level        string
	Formatter    string
	ReportCaller bool
}

var defaultOption = &Option{
	Output:       os.Stdout,
	Level:        "INFO",
	Formatter:    JsonFormatter,
	ReportCaller: true,
}

// 创建日志记录器
func New() Logger {
	return NewWithOption(defaultOption)
}

// 通过Option，创建日志记录器
func NewWithOption(option *Option) Logger {
	if option == nil {
		option = defaultOption
	}

	logger := logrus.New()

	// set level
	if option.Level != "" {
		level, err := logrus.ParseLevel(option.Level)
		if err == nil {
			logger.SetLevel(level)
		}
	}

	// set formatter
	if option.Formatter == JsonFormatter {
		logger.SetFormatter(&logrus.JSONFormatter{})
	}

	// set output
	if option.Output != nil {
		logger.SetOutput(option.Output)
	} else {
		logger.SetOutput(os.Stdout)
	}

	// set no lock
	logger.SetNoLock()

	return &log{
		entry:        logrus.NewEntry(logger),
		depth:        1,
		reportCaller: option.ReportCaller,
	}
}

// 通往logrus的最终入口
func (l *log) log(level logrus.Level, args ...interface{}) {
	entry := l.entry
	// report caller
	if l.reportCaller {
		entry = l.entry.WithField("file", caller(l.depth+3))
	}
	entry.Log(level, args...)

	//输出error/fatal数目到metrics
	if metricsFunc != nil {
		levelStr := ""
		if level == logrus.ErrorLevel {
			levelStr = "ERROR"
		} else if level == logrus.FatalLevel {
			levelStr = "FATAL"
		}
		if levelStr != "" {
			metricsFunc(levelStr)
		}
	}
}

// 记录一条 LevelDebug 级别的日志
func (l *log) Log(level Level, args ...interface{}) {
	driverLevel, exists := levels[level]
	if !exists {
		l.log(logrus.WarnLevel, errUnknowLevel)
		return
	}
	if l.entry.Logger.IsLevelEnabled(driverLevel) {
		l.log(driverLevel, args...)
	}
}

// 记录一条 LevelDebug 级别的日志
func (l *log) Logf(level Level, format string, args ...interface{}) {
	driverLevel, exists := levels[level]
	if !exists {
		l.log(logrus.WarnLevel, errUnknowLevel)
		return
	}
	if l.entry.Logger.IsLevelEnabled(driverLevel) {
		l.log(driverLevel, fmt.Sprintf(format, args...))
	}
}

// 记录一条 LevelDebug 级别的日志
func (l *log) Trace(args ...interface{}) {
	l.Log(TraceLevel, args...)
}

// 格式化并记录一条 LevelDebug 级别的日志
func (l *log) Tracef(format string, args ...interface{}) {
	l.Logf(TraceLevel, format, args...)
}

// 记录一条 LevelDebug 级别的日志
func (l *log) Debug(args ...interface{}) {
	l.Log(DebugLevel, args...)
}

// 格式化并记录一条 LevelDebug 级别的日志
func (l *log) Debugf(format string, args ...interface{}) {
	l.Logf(DebugLevel, format, args...)
}

// 记录一条 LevelInfo 级别的日志
func (l *log) Info(args ...interface{}) {
	l.Log(InfoLevel, args...)
}

// 格式化并记录一条 LevelInfo 级别的日志
func (l *log) Infof(format string, args ...interface{}) {
	l.Logf(InfoLevel, format, args...)
}

// 记录一条 LevelWarn 级别的日志
func (l *log) Warn(args ...interface{}) {
	l.Log(WarnLevel, args...)
}

// 格式化并记录一条 LevelWarn 级别的日志
func (l *log) Warnf(format string, args ...interface{}) {
	l.Logf(WarnLevel, format, args...)
}

// 记录一条 LevelError 级别的日志
func (l *log) Error(args ...interface{}) {
	l.Log(ErrorLevel, args...)
}

// 格式化并记录一条 LevelError 级别的日志
func (l *log) Errorf(format string, args ...interface{}) {
	l.Logf(ErrorLevel, format, args...)
}

// 记录一条 LevelFatal 级别的日志
func (l *log) Fatal(args ...interface{}) {
	l.Log(FatalLevel, args...)
	l.entry.Logger.Exit(1) // 和logrus保持一致
}

// 格式化并记录一条 LevelFatal 级别的日志
func (l *log) Fatalf(format string, args ...interface{}) {
	l.Logf(FatalLevel, format, args...)
	l.entry.Logger.Exit(1)
}

// 记录一条 LevelPanic 级别的日志
func (l *log) Panic(args ...interface{}) {
	l.Log(PanicLevel, args...)
	panic(fmt.Sprint(args...))
}

// 格式化并记录一条 LevelPanic 级别的日志
func (l *log) Panicf(format string, args ...interface{}) {
	l.Logf(PanicLevel, format, args...)
	panic(fmt.Sprintf(format, args...))
}

// 为当前日志附加一个上下文数据
func (l *log) WithField(key string, value interface{}) Logger {
	return l.WithFields(Fields{key: value})
}

// 为当前日志附加一组上下文数据
func (l *log) WithFields(fields Fields) Logger {
	if l.reportCaller {
		if err, ok := fields[errorKey].(interface {
			Stack() []string
		}); ok {
			fields["err.stack"] = strings.Join(err.Stack(), ";")
		}
	}
	return &log{entry: l.entry.WithFields(logrus.Fields(fields)), reportCaller: l.reportCaller}
}

// 为当前日志附加一个错误
func (l *log) WithError(err error) Logger {
	return l.WithFields(Fields{errorKey: err})
}

// 设置日志等级
func (l *log) SetLevel(level Level) {
	driverLevel, exists := levels[level]
	if !exists {
		l.log(logrus.WarnLevel, errUnknowLevel)
	}
	l.entry.Logger.SetLevel(driverLevel)
}

// 获取日志等级
func (l *log) GetLevel() Level {
	return Level(l.entry.Logger.GetLevel())
}

// caller的显示形式为 File:Line
func caller(depth int) string {
	_, f, n, ok := runtime.Caller(1 + depth)
	if !ok {
		return ""
	}
	if ok {
		idx := strings.LastIndex(f, "leon-yc")
		if idx >= 0 {
			f = f[idx+18:]
		}
	}
	return fmt.Sprintf("%s:%d", f, n)
}
