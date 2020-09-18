package qlog

type MetricsFunc func(string)

var logger = New()
var metricsFunc MetricsFunc

func SetMetricsFunc(f func(string)) {
	metricsFunc = f
}

func SetLogger(l Logger) {
	logger = l
}

func GetLogger() Logger {
	return logger
}

func SetLevel(level Level) {
	logger.SetLevel(level)
}

func GetLevel() Level {
	return logger.GetLevel()
}

func Log(level Level, args ...interface{}) {
	logger.Log(level, args...)
}

func Logf(level Level, format string, args ...interface{}) {
	logger.Logf(level, format, args...)
}

func Trace(args ...interface{}) {
	logger.Trace(args...)
}

func Tracef(format string, args ...interface{}) {
	logger.Tracef(format, args...)
}

func Debug(args ...interface{}) {
	logger.Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	logger.Debugf(format, args...)
}

func Info(args ...interface{}) {
	logger.Info(args...)
}

func Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

func Warn(args ...interface{}) {
	logger.Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	logger.Warnf(format, args...)
}

func Error(args ...interface{}) {
	logger.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}

func Fatal(args ...interface{}) {
	logger.Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	logger.Fatalf(format, args...)
}

func Panic(args ...interface{}) {
	logger.Panic(args...)
}

func Panicf(format string, args ...interface{}) {
	logger.Panicf(format, args...)
}

func WithField(key string, value interface{}) Logger {
	return logger.WithField(key, value)
}

func WithFields(fields Fields) Logger {
	return logger.WithFields(fields)
}

func WithError(err error) Logger {
	return logger.WithError(err)
}
