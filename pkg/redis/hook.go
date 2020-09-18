package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/leon-yc/ggs/pkg/metrics"
	"github.com/go-redis/redis/v7"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

const (
	RequestRedisSuccess = "200"
	RequestRedisFail    = "400"
	RequestRedisNil     = "404"
)

type contextKey struct{}

var activeSpanKey = contextKey{}

type ContextVal struct {
	Span opentracing.Span
	St   time.Time
}

type hook struct {
	optName        string
	cfg            *Config
	disableMetrics bool
}

func newHook(name string, cfg *Config, disableMetrics bool) *hook {
	return &hook{
		optName:        name,
		cfg:            cfg,
		disableMetrics: disableMetrics,
	}
}

func (h *hook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	if ctx == nil {
		return ctx, nil
	}
	parentSpan := opentracing.SpanFromContext(ctx)
	if parentSpan == nil {
		return ctx, nil
	}
	tracer := opentracing.GlobalTracer()
	if tracer == nil {
		return ctx, nil
	}

	optName := fmt.Sprintf("redis_%s", cmd.Name())
	span := tracer.StartSpan(optName, opentracing.ChildOf(parentSpan.Context()))
	ext.DBType.Set(span, "redis")
	ext.DBInstance.Set(span, strconv.Itoa(h.cfg.DB))
	ext.PeerService.Set(span, "redis")
	ext.PeerHostname.Set(span, h.cfg.Addr)
	ext.SpanKindRPCClient.Set(span)

	contextVal := &ContextVal{
		Span: span,
		St:   time.Now(),
	}
	return context.WithValue(ctx, activeSpanKey, contextVal), nil
}

func (h *hook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	if ctx == nil || cmd == nil {
		return nil
	}

	cmdName := cmd.Name()
	status := RequestRedisSuccess
	if cmd.Err() != nil {
		if cmd.Err() == redis.Nil {
			status = RequestRedisNil
		} else {
			status = RequestRedisFail
		}
	}

	val := ctx.Value(activeSpanKey)
	ctxVal, ok := val.(*ContextVal)
	if !ok || ctxVal == nil {
		return nil
	}

	if !h.disableMetrics {
		metrics.CounterAdd(metrics.RedisReqCount, 1,
			map[string]string{metrics.RedisRedisName: h.optName, metrics.RedisReqCMD: cmdName, metrics.RedisRespStatus: status})

		metrics.HistogramObserve(metrics.RedisReqDurationSecond, time.Since(ctxVal.St).Seconds(),
			map[string]string{metrics.RedisRedisName: h.optName, metrics.RedisReqCMD: cmdName, metrics.RedisRespStatus: status})
	}

	span := ctxVal.Span
	defer span.Finish()

	state := fmt.Sprintf("%v", cmd.Args())
	ext.DBStatement.Set(span, state)
	if err := cmd.Err(); err != nil {
		if err != redis.Nil {
			ext.Error.Set(span, true)
			span.LogKV("event", "error", "message", err.Error())
		}
	}
	return nil
}

func (h *hook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	if ctx == nil || len(cmds) == 0 {
		return ctx, nil
	}
	parentSpan := opentracing.SpanFromContext(ctx)
	if parentSpan == nil {
		return ctx, nil
	}
	tracer := opentracing.GlobalTracer()
	if tracer == nil {
		return ctx, nil
	}

	optName := fmt.Sprintf("redis_pipeline_%s", cmds[0].Name())
	span := tracer.StartSpan(optName, opentracing.ChildOf(parentSpan.Context()))
	ext.DBType.Set(span, "redis")
	ext.DBInstance.Set(span, strconv.Itoa(h.cfg.DB))
	ext.PeerService.Set(span, "redis")
	ext.PeerHostname.Set(span, h.cfg.Addr)
	ext.SpanKindRPCClient.Set(span)

	contextVal := &ContextVal{
		Span: span,
		St:   time.Now(),
	}
	return context.WithValue(ctx, activeSpanKey, contextVal), nil
}

func (h *hook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	if ctx == nil || len(cmds) == 0 {
		return nil
	}

	cmdName := cmds[0].Name()
	status := RequestRedisSuccess
	if cmds[0].Err() != nil {
		if cmds[0].Err() == redis.Nil {
			status = RequestRedisNil
		} else {
			status = RequestRedisFail
		}
	}

	val := ctx.Value(activeSpanKey)
	ctxVal, ok := val.(*ContextVal)
	if !ok || ctxVal == nil {
		return nil
	}

	if !h.disableMetrics {
		metrics.CounterAdd(metrics.RedisReqCount, 1,
			map[string]string{metrics.RedisRedisName: h.optName, metrics.RedisReqCMD: cmdName, metrics.RedisRespStatus: status})

		metrics.HistogramObserve(metrics.RedisReqDurationSecond, time.Since(ctxVal.St).Seconds(),
			map[string]string{metrics.RedisRedisName: h.optName, metrics.RedisReqCMD: cmdName, metrics.RedisRespStatus: status})
	}

	span := ctxVal.Span
	defer span.Finish()

	ext.DBStatement.Set(span, fmt.Sprintf("pipeline %v", cmds[0].Args()))
	if err := cmds[0].Err(); err != nil {
		if err != redis.Nil {
			ext.Error.Set(span, true)
			span.LogKV("event", "error", "message", err.Error())
		}
	}
	return nil
}
