package handler

import (
	"strconv"
	"time"

	"github.com/leon-yc/ggs/internal/core/config"

	"github.com/leon-yc/ggs/internal/core/invocation"
	"github.com/leon-yc/ggs/pkg/metrics"
	"github.com/leon-yc/ggs/pkg/qlog"
)

type MetricsProviderHandler struct{}

func (m *MetricsProviderHandler) Name() string {
	return MetricsProvider
}

func newMetricsProviderHandler() Handler {
	return &MetricsProviderHandler{}
}

func (m *MetricsProviderHandler) Handle(chain *Chain, i *invocation.Invocation, cb invocation.ResponseCallBack) {
	if !config.GlobalDefinition.Ggs.Metrics.Enabled {
		chain.Next(i, cb)
		return
	}

	st := time.Now()
	uri := i.URLPathFormat

	chain.Next(i, func(r *invocation.Response) (err error) {
		defer func() {
			if i.Protocol == ProtocolRest {
				err := metrics.HistogramObserve(metrics.ReqDuration, time.Since(st).Seconds(),
					map[string]string{metrics.ReqProtocolLable: i.Protocol,
						metrics.RespUriLable:  uri,
						metrics.RespCodeLable: strconv.Itoa(r.Status)})
				if err != nil {
					qlog.Errorf("HistogramObserve, uri:%s status:%d err:%s", uri, r.Status, err.Error())
				}

				err = metrics.CounterAdd(metrics.ReqQPS, 1,
					map[string]string{metrics.ReqProtocolLable: i.Protocol,
						metrics.RespUriLable:  uri,
						metrics.RespCodeLable: strconv.Itoa(r.Status)})
				if err != nil {
					qlog.Errorf("CounterAdd, uri:%s  status:%d err:%s", uri, r.Status, err.Error())
				}

			} else if i.Protocol == ProtocolGrpc {
				err := metrics.HistogramObserve(metrics.GrpcReqDuration, time.Since(st).Seconds(),
					map[string]string{metrics.ReqProtocolLable: i.Protocol,
						metrics.RespHandlerLable: i.OperationID,
						metrics.RespCodeLable:    strconv.Itoa(r.Status)})
				if err != nil {
					qlog.Errorf("HistogramObserve, RespHandler:%s  status:%d err:%s",
						i.OperationID, r.Status, err.Error())
				}

				err = metrics.CounterAdd(metrics.GrpcReqQPS, 1,
					map[string]string{metrics.ReqProtocolLable: i.Protocol,
						metrics.RespHandlerLable: i.OperationID,
						metrics.RespCodeLable:    strconv.Itoa(r.Status)})
				if err != nil {
					qlog.Errorf("HistogramObserve, RespHandler:%s  status:%d err:%s",
						i.OperationID, r.Status, err.Error())
				}
			}
		}()

		return cb(r)
	})
}

//consumer
type MetricsConsumerHandler struct{}

func (m *MetricsConsumerHandler) Name() string {
	return MetricsProvider
}

func newMetricsConsumerHandler() Handler {
	return &MetricsConsumerHandler{}
}

func (m *MetricsConsumerHandler) Handle(chain *Chain, i *invocation.Invocation, cb invocation.ResponseCallBack) {
	if !config.GlobalDefinition.Ggs.Metrics.Enabled {
		chain.Next(i, cb)
		return
	}

	st := time.Now()
	uri := i.URLPathFormat

	chain.Next(i, func(r *invocation.Response) (err error) {
		defer func() {
			if i.Protocol == ProtocolRest {
				err := metrics.HistogramObserve(metrics.ClientReqDuration, time.Since(st).Seconds(),
					map[string]string{metrics.RemoteLable: i.MicroServiceName,
						metrics.ReqProtocolLable: i.Protocol,
						metrics.RespUriLable:     uri,
						metrics.RespCodeLable:    strconv.Itoa(r.Status)})
				if err != nil {
					qlog.Errorf("HistogramObserve, uri:%s status:%d err:%s", uri, r.Status, err.Error())
				}

				err = metrics.CounterAdd(metrics.ClientReqQPS, 1,
					map[string]string{metrics.RemoteLable: i.MicroServiceName,
						metrics.ReqProtocolLable: i.Protocol,
						metrics.RespUriLable:     uri,
						metrics.RespCodeLable:    strconv.Itoa(r.Status)})
				if err != nil {
					qlog.Errorf("CounterAdd, uri:%s  status:%d err:%s", uri, r.Status, err.Error())
				}

			} else if i.Protocol == ProtocolGrpc {
				err := metrics.HistogramObserve(metrics.ClientGrpcReqDuration, time.Since(st).Seconds(),
					map[string]string{metrics.RemoteLable: i.MicroServiceName,
						metrics.ReqProtocolLable: i.Protocol,
						metrics.RespHandlerLable: i.OperationID,
						metrics.RespCodeLable:    strconv.Itoa(r.Status)})
				if err != nil {
					qlog.Errorf("HistogramObserve, RespHandler:%s  status:%d err:%s",
						i.OperationID, r.Status, err.Error())
				}

				err = metrics.CounterAdd(metrics.ClientGrpcReqQPS, 1,
					map[string]string{metrics.RemoteLable: i.MicroServiceName,
						metrics.ReqProtocolLable: i.Protocol,
						metrics.RespHandlerLable: i.OperationID,
						metrics.RespCodeLable:    strconv.Itoa(r.Status)})
				if err != nil {
					qlog.Errorf("HistogramObserve, RespHandler:%s  status:%d err:%s",
						i.OperationID, r.Status, err.Error())
				}
			}
		}()

		return cb(r)
	})
}
