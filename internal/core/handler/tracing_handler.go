package handler

import (
	"fmt"
	"io"
	"net/http"

	"github.com/leon-yc/ggs/internal/core/config"

	"github.com/leon-yc/ggs/internal/core/common"
	"github.com/leon-yc/ggs/internal/core/invocation"
	"github.com/leon-yc/ggs/internal/core/tracing"
	"github.com/leon-yc/ggs/pkg/qlog"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

var (
	httpTag = opentracing.Tag{Key: string(ext.Component), Value: "net/http"}
	grpcTag = opentracing.Tag{Key: string(ext.Component), Value: "gRPC"}
)

// TracingProviderHandler tracing provider handler
type TracingProviderHandler struct{}

// Handle is to handle the provider tracing related things
func (t *TracingProviderHandler) Handle(chain *Chain, i *invocation.Invocation, cb invocation.ResponseCallBack) {
	if config.GlobalDefinition.Ggs.Tracing.Disabled {
		chain.Next(i, cb)
		return
	}

	switch i.Protocol {
	case ProtocolRest:
		t.HttpHandle(chain, i, cb)
	case ProtocolGrpc:
		t.GrpcHandle(chain, i, cb)
	}
}

// Handle is to handle the http provider tracing related things
func (t *TracingProviderHandler) HttpHandle(chain *Chain, i *invocation.Invocation, cb invocation.ResponseCallBack) {
	var (
		wireContext opentracing.SpanContext
		span        opentracing.Span
	)

	// extract span context
	wireContext, _ = opentracing.GlobalTracer().Extract(opentracing.TextMap, opentracing.TextMapCarrier(i.Headers()))
	// start span
	span = opentracing.StartSpan(i.OperationID, ext.RPCServerOption(wireContext), httpTag)
	// set tags
	span.SetTag(tracing.HTTPMethod, i.Metadata[common.RestMethod])
	span.SetTag(tracing.HTTPPath, i.OperationID)

	// inject span context into carrier
	if err := opentracing.GlobalTracer().Inject(
		span.Context(),
		opentracing.TextMap,
		opentracing.TextMapCarrier(i.Headers()),
	); err != nil {
		qlog.Errorf("Inject span failed, err [%s]", err.Error())
	}

	// store span in ctx
	i.Ctx = opentracing.ContextWithSpan(i.Ctx, span)

	// To ensure accuracy, spans should finish immediately once server responds.
	// So the best way is that spans finish in the callback func, not after it.
	// But server may respond in the callback func too, that we have to remove
	// span finishing from callback func's inside to outside.
	chain.Next(i, func(r *invocation.Response) (err error) {
		err = cb(r)
		span.SetTag(tracing.HTTPStatusCode, r.Status)
		if err != nil {
			ext.Error.Set(span, true)
			span.LogKV("event", "error", "message", err.Error())
		}
		span.Finish()
		return
	})
}

// Handle is to handle the grpc provider tracing related things
func (t *TracingProviderHandler) GrpcHandle(chain *Chain, i *invocation.Invocation, cb invocation.ResponseCallBack) {
	var (
		err         error
		wireContext opentracing.SpanContext
		span        opentracing.Span
	)

	// extract span context
	wireContext, err = opentracing.GlobalTracer().Extract(opentracing.TextMap, opentracing.TextMapCarrier(i.Headers()))
	if err != nil && err != opentracing.ErrSpanContextNotFound {
		qlog.Errorf("Extract span failed, err [%s]", err.Error())
	}
	// start span
	span = opentracing.StartSpan(i.OperationID, ext.RPCServerOption(wireContext), grpcTag)

	// inject span context into carrier
	if err := opentracing.GlobalTracer().Inject(
		span.Context(),
		opentracing.TextMap,
		opentracing.TextMapCarrier(i.Headers()),
	); err != nil {
		qlog.Errorf("Inject span failed, err [%s]", err.Error())
	}

	// store span in ctx
	i.Ctx = opentracing.ContextWithSpan(i.Ctx, span)

	// To ensure accuracy, spans should finish immediately once server responds.
	// So the best way is that spans finish in the callback func, not after it.
	// But server may respond in the callback func too, that we have to remove
	// span finishing from callback func's inside to outside.
	chain.Next(i, func(r *invocation.Response) (err error) {
		err = cb(r)
		if err != nil {
			ext.Error.Set(span, true)
			span.LogKV("event", "error", "message", err.Error())
		}
		span.Finish()
		return
	})
}

// Name returns tracing-provider string
func (t *TracingProviderHandler) Name() string {
	return TracingProvider
}

func newTracingProviderHandler() Handler {
	return &TracingProviderHandler{}
}

// TracingConsumerHandler tracing consumer handler
type TracingConsumerHandler struct{}

// Handle is handle consumer tracing related things
func (t *TracingConsumerHandler) Handle(chain *Chain, i *invocation.Invocation, cb invocation.ResponseCallBack) {
	if config.GlobalDefinition.Ggs.Tracing.Disabled {
		chain.Next(i, cb)
		return
	}

	switch i.Protocol {
	case ProtocolRest:
		t.HttpHandle(chain, i, cb)
	case ProtocolGrpc:
		t.GrpcHandle(chain, i, cb)
	}
}

// Handle is handle consumer tracing related things
func (t *TracingConsumerHandler) HttpHandle(chain *Chain, i *invocation.Invocation, cb invocation.ResponseCallBack) {
	var (
		span        opentracing.Span
		wireContext opentracing.SpanContext
	)
	wireContext, _ = opentracing.GlobalTracer().Extract(opentracing.TextMap, opentracing.TextMapCarrier(i.Headers()))
	// store span in context
	span = opentracing.StartSpan(i.OperationID, opentracing.ChildOf(wireContext))

	ext.SpanKindRPCClient.Set(span)
	ext.Component.Set(span, "net/http")
	span.SetTag(tracing.HTTPMethod, i.Metadata[common.RestMethod])
	span.SetTag(tracing.HTTPPath, i.OperationID)

	// inject span context into carrier
	if err := opentracing.GlobalTracer().Inject(
		span.Context(),
		opentracing.TextMap,
		opentracing.TextMapCarrier(i.Headers()),
	); err != nil {
		qlog.Errorf("Inject span failed, err [%s]", err.Error())
	}

	// header stored in context
	i.Ctx = opentracing.ContextWithSpan(i.Ctx, span)

	// To ensure accuracy, spans should finish immediately once client send req.
	// So the best way is that spans finish in the callback func, not after it.
	// But client may send req in the callback func too, that we have to remove
	// span finishing from callback func's inside to outside.
	chain.Next(i, func(r *invocation.Response) (err error) {
		span.SetTag(tracing.HTTPStatusCode, r.Status)
		if r.Err != nil {
			ext.Error.Set(span, true)
			span.LogKV("event", "error", "message", r.Err.Error())
		} else if r.Status >= http.StatusInternalServerError {
			ext.Error.Set(span, true)
			span.LogKV("event", "error", "message", fmt.Sprintf("httpCodeError: %d", r.Status))
		}
		span.Finish()
		return cb(r)
	})
}

// Handle is handle consumer tracing related things
func (t *TracingConsumerHandler) GrpcHandle(chain *Chain, i *invocation.Invocation, cb invocation.ResponseCallBack) {
	wireContext, _ := opentracing.GlobalTracer().Extract(opentracing.TextMap, opentracing.TextMapCarrier(i.Headers()))
	opts := []opentracing.StartSpanOption{
		opentracing.ChildOf(wireContext),
		ext.SpanKindRPCClient,
		grpcTag,
	}
	span := opentracing.StartSpan(i.OperationID, opts...)
	// Make sure we add this to the metadata of the call, so it gets propagated:
	if err := opentracing.GlobalTracer().Inject(
		span.Context(),
		opentracing.TextMap,
		opentracing.TextMapCarrier(i.Headers()),
	); err != nil {
		qlog.Errorf("Inject span failed, err [%s]", err.Error())
	}

	// store span in context
	i.Ctx = opentracing.ContextWithSpan(i.Ctx, span)

	// To ensure accuracy, spans should finish immediately once client send req.
	// So the best way is that spans finish in the callback func, not after it.
	// But client may send req in the callback func too, that we have to remove
	// span finishing from callback func's inside to outside.
	chain.Next(i, func(r *invocation.Response) (err error) {
		if r.Err != nil && r.Err != io.EOF {
			ext.Error.Set(span, true)
			span.LogKV("event", "error", "message", r.Err.Error())
		}
		span.Finish()
		return cb(r)
	})
}

// Name returns tracing-consumer string
func (t *TracingConsumerHandler) Name() string {
	return TracingConsumer
}

func newTracingConsumerHandler() Handler {
	return &TracingConsumerHandler{}
}
