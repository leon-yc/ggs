package jaeger

import (
	"strconv"

	"github.com/leon-gopher/gtracing"
	"github.com/leon-yc/ggs/internal/core/tracing"
	"github.com/leon-yc/ggs/internal/pkg/runtime"
	"github.com/leon-yc/ggs/pkg/qlog"
	"github.com/opentracing/opentracing-go"
)

const (
	TraceFileName    = "traceFileName"
	SamplingRate     = "samplingRate"
	BufferSize       = "bufferSize"
	SamplingRateDef  = 1.0
	TraceFileNameDef = "/data/logs/trace/trace.log"
)

func init() {
	tracing.InstallTracer("jaeger", NewTracer)
}

func NewTracer(options map[string]string) (opentracing.Tracer, error) {
	var err error
	samplingRate := SamplingRateDef
	if options[SamplingRate] != "" {
		samplingRate, err = strconv.ParseFloat(options[SamplingRate], 64)
		if err != nil {
			qlog.Errorf("parse sampling rate failed: %v", err)
			return nil, err
		}
	}

	var traceFileName = TraceFileNameDef
	if options[TraceFileName] != "" {
		traceFileName = options[TraceFileName]
	}

	var bufferSize int64
	if options[BufferSize] != "" {
		bufferSize, err = strconv.ParseInt(options[BufferSize], 10, 64)
		if err != nil {
			qlog.Errorf("parse buffer size failed: %v", err)
			return nil, err
		}
	}

	tracer, _, err := gtracing.NewJaegerTracer(runtime.ServiceName, traceFileName, samplingRate, nil, int(bufferSize))
	if err != nil {
		return nil, err
	}

	return tracer, nil
}
