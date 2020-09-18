package eventlistener

import (
	"strings"

	"github.com/go-chassis/go-archaius/event"

	"github.com/leon-yc/ggs/internal/core/common"
	"github.com/leon-yc/ggs/internal/core/qpslimiter"
)

const (
	//QPSLimitKey is a variable of type string
	QPSLimitKey = "ggs.flowcontrol"
)

//QPSEventListener is a struct used for Event listener
type QPSEventListener struct {
	//Key []string
	Key string
}

//Event is a method for QPS event listening
func (el *QPSEventListener) Event(e *event.Event) {
	qpsLimiter := qpslimiter.GetQPSTrafficLimiter()

	if strings.Contains(e.Key, "enabled") {
		return
	}

	switch e.EventType {
	case common.Update:
		qpsLimiter.UpdateRateLimit(e.Key, e.Value)
	case common.Create:
		qpsLimiter.UpdateRateLimit(e.Key, e.Value)
	case common.Delete:
		qpsLimiter.DeleteRateLimiter(e.Key)
	}
}
