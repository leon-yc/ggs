package handler

import (
	"fmt"
	"net/http"

	"github.com/leon-yc/ggs/internal/control"
	"github.com/leon-yc/ggs/internal/core/common"
	"github.com/leon-yc/ggs/internal/core/invocation"
	"github.com/leon-yc/ggs/internal/core/qpslimiter"
	pkgerr "github.com/leon-yc/ggs/pkg/errors"
)

// ConsumerRateLimiterHandler consumer rate limiter handler
type ConsumerRateLimiterHandler struct{}

// Handle is handles the consumer rate limiter APIs
func (rl *ConsumerRateLimiterHandler) Handle(chain *Chain, i *invocation.Invocation, cb invocation.ResponseCallBack) {
	rlc := control.DefaultPanel.GetRateLimiting(*i, common.Consumer)
	if !rlc.Enabled {
		chain.Next(i, cb)

		return
	}

	limited := false
	if rlc.Rate <= 0 {
		limited = true
	} else {
		//get operation meta info ms.schema, ms.schema.operation, ms
		allowed := rl.GetOrCreate(rlc)
		if !allowed {
			limited = true
		}
	}

	if limited {
		switch i.Reply.(type) {
		case *http.Response:
			resp := i.Reply.(*http.Response)
			resp.StatusCode = http.StatusTooManyRequests
		}
		r := &invocation.Response{}
		r.Status = http.StatusTooManyRequests
		r.Err = pkgerr.WithMessage(pkgerr.ErrRateLimit, fmt.Sprintf("ratelimit: %s|%v", rlc.Key, rlc.Rate))
		cb(r)
		return
	}

	chain.Next(i, cb)
}

func newConsumerRateLimiterHandler() Handler {
	return &ConsumerRateLimiterHandler{}
}

// Name returns consumerratelimiter string
func (rl *ConsumerRateLimiterHandler) Name() string {
	return "consumerratelimiter"
}

// GetOrCreate is for getting or creating qps limiter meta data
func (rl *ConsumerRateLimiterHandler) GetOrCreate(rlc control.RateLimitingConfig) bool {
	return qpslimiter.GetQPSTrafficLimiter().ProcessQPSTokenReq(rlc.Key, rlc.Rate)
}
