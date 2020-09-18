package eventlistener

import (
	"github.com/leon-yc/ggs/internal/core/common"
	"github.com/leon-yc/ggs/pkg/qlog"
	"github.com/go-chassis/go-archaius/event"
)

const (
	//LagerLevelKey is a variable of type string
	LoggerLevelKey = "logger_level"
)

//LagerEventListener is a struct used for Event listener
type LoggerEventListener struct {
	//Key []string
	Key string
}

//Event is a method for Lager event listening
func (el *LoggerEventListener) Event(e *event.Event) {
	qlog.WithFields(qlog.Fields{
		"key":   e.Key,
		"value": e.Value,
		"type":  e.EventType,
	}).Info("Get logger e")

	v, ok := e.Value.(string)
	if !ok {
		return
	}

	switch e.EventType {
	case common.Update:
		level, err := qlog.ParseLevel(v)
		if err != nil {
			qlog.Error(err)
			return
		}
		qlog.SetLevel(level)
	}
}
