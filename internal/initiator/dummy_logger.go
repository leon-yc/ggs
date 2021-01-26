package initiator

import (
	"github.com/go-chassis/openlog"
)

type dummpLogger struct {
}

func (logger dummpLogger) Debug(message string, opts ...openlog.Option) {
}

func (logger dummpLogger) Info(message string, opts ...openlog.Option) {
}

func (logger dummpLogger) Warn(message string, opts ...openlog.Option) {
}

func (logger dummpLogger) Error(message string, opts ...openlog.Option) {
}

func (logger dummpLogger) Fatal(message string, opts ...openlog.Option) {
}

func (logger dummpLogger) Debugf(template string, args ...interface{}) {
}

func (logger dummpLogger) Infof(template string, args ...interface{}) {
}

func (logger dummpLogger) Warnf(template string, args ...interface{}) {
}

func (logger dummpLogger) Errorf(template string, args ...interface{}) {
}

func (logger dummpLogger) Fatalf(template string, args ...interface{}) {
}
