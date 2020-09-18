package consul

import (
	"github.com/leon-gopher/discovery/logger"
	"github.com/leon-yc/ggs/pkg/qlog"
)

func init() {
	logger.SetLogger(&SdkLog{})
}

type SdkLog struct{}

func (*SdkLog) Errorf(format string, v ...interface{}) {
	qlog.Errorf(format, v...)
}

func (*SdkLog) Warnf(format string, v ...interface{}) {
	qlog.Warnf(format, v...)
}

func (*SdkLog) Infof(format string, v ...interface{}) {
	qlog.Infof(format, v...)
}

func (*SdkLog) Debugf(format string, v ...interface{}) {
	qlog.Tracef(format, v...)
}
