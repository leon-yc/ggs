package handler

import (
	"fmt"
	"net/http"
	"path"
	"time"

	"github.com/leon-yc/ggs/internal/core/config"
	"github.com/leon-yc/ggs/internal/core/invocation"
	"github.com/leon-yc/ggs/internal/pkg/util/httputil"
	"github.com/leon-yc/ggs/pkg/qlog"
	"google.golang.org/grpc/codes"
)

// LogProviderHandler tracing provider handler
type LogProviderHandler struct{}

// Handle is to handle the provider tracing related things
func (t *LogProviderHandler) Handle(chain *Chain, i *invocation.Invocation, cb invocation.ResponseCallBack) {
	if !config.GlobalDefinition.Ggs.AccessLog.Enabled {
		chain.Next(i, cb)
		return
	}

	l, err := newLogParams(i)
	if err != nil {
		chain.Next(i, cb)
		return
	}

	chain.Next(i, func(r *invocation.Response) (err error) {
		err = cb(r)

		l.format(r.Status, r.Err)
		return
	})
}

// Name returns tracing-provider string
func (t *LogProviderHandler) Name() string {
	return LogProvider
}

func newLogProviderHandler() Handler {
	return &LogProviderHandler{}
}

func init() {
	RegisterHandler(LogProvider, newLogProviderHandler)
}

type logParams struct {
	protocol string
	start    time.Time
	fields   qlog.Fields
}

func newLogParams(i *invocation.Invocation) (*logParams, error) {
	l := &logParams{
		protocol: i.Protocol,
		start:    time.Now(),
		fields:   make(qlog.Fields, 10),
	}

	switch i.Protocol {
	case ProtocolRest:
		request, err := httputil.HTTPRequest(i)
		if err != nil {
			qlog.Error("extract request from invocation failed")
			return nil, err
		}

		path := request.URL.Path
		raw := request.URL.RawQuery
		if raw != "" {
			path = path + "?" + raw
		}
		l.fields["component"] = "net/http"
		l.fields["path"] = path
		l.fields["method"] = request.Method
	case ProtocolGrpc:
		l.fields["component"] = "grpc"
		l.fields["service"] = path.Dir(i.SchemaID)[1:]
		l.fields["method"] = path.Base(i.SchemaID)
	}
	return l, nil
}

func (l *logParams) format(code int, err error) {
	l.fields["duration"] = fmt.Sprintf("%v", time.Since(l.start))
	var level qlog.Level

	switch l.protocol {
	case ProtocolRest:
		l.fields["code"] = code
		level = l.httpCode2Level(code)
	case ProtocolGrpc:
		code := codes.Code(code)
		l.fields["code"] = code.String()
		level = l.grpcCode2Level(code)
	}
	if err != nil {
		level = qlog.ErrorLevel
		l.fields["error"] = err
	}
	qlog.WithFields(l.fields).Logf(level, "finished call with code %v", code)
}

// StatusCodeColor is the ANSI color for appropriately logging http status code to a terminal.
func (l *logParams) httpCode2Level(code int) qlog.Level {
	switch {
	case code >= http.StatusOK && code < http.StatusMultipleChoices:
		return qlog.InfoLevel
	case code >= http.StatusMultipleChoices && code < http.StatusBadRequest:
		return qlog.WarnLevel
	case code >= http.StatusBadRequest && code < http.StatusInternalServerError:
		return qlog.WarnLevel
	default:
		return qlog.ErrorLevel
	}
}

// code2Level is the default implementation of gRPC return codes to log levels for server side.
func (l *logParams) grpcCode2Level(code codes.Code) qlog.Level {
	switch code {
	case codes.OK, codes.Canceled, codes.InvalidArgument, codes.NotFound, codes.AlreadyExists, codes.Unauthenticated:
		return qlog.InfoLevel
	case codes.DeadlineExceeded, codes.PermissionDenied, codes.ResourceExhausted, codes.FailedPrecondition, codes.Aborted, codes.OutOfRange, codes.Unavailable:
		return qlog.WarnLevel
	default:
		return qlog.ErrorLevel
	}
}
