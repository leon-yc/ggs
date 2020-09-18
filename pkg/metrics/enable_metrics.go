package metrics

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/leon-yc/ggs/pkg/errors"

	"github.com/go-chassis/go-archaius"
	"github.com/leon-yc/ggs/internal/pkg/util/iputil"
	"github.com/leon-yc/ggs/pkg/qlog"
	"gopkg.in/yaml.v2"
)

var GgsBuckets = []float64{.0005, .001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}

func enableErrorLogMetrics() error {
	return CreateCounter(CounterOpts{
		Name:   ErrorNum,
		Help:   ErrorNumHelp,
		Labels: []string{ErrorNumLevel},
	})
}

func countErrorNumber(level string) {
	_ = CounterAdd(ErrorNum, 1, map[string]string{
		ErrorNumLevel: level,
	})
}

func enableProviderMetrics() error {
	//REST
	if err := CreateCounter(CounterOpts{
		Name:   ReqQPS,
		Help:   ReqQPSHelp,
		Labels: []string{ReqProtocolLable, RespUriLable, RespCodeLable},
	}); err != nil {
		return err
	}

	if err := CreateHistogram(HistogramOpts{
		Name:    ReqDuration,
		Help:    ReqDurationHelp,
		Labels:  []string{ReqProtocolLable, RespUriLable, RespCodeLable},
		Buckets: GgsBuckets,
	}); err != nil {
		return err
	}

	//GRPC
	if err := CreateCounter(CounterOpts{
		Name:   GrpcReqQPS,
		Help:   GrpcReqQPSHelp,
		Labels: []string{ReqProtocolLable, RespHandlerLable, RespCodeLable},
	}); err != nil {
		return err
	}

	if err := CreateHistogram(HistogramOpts{
		Name:    GrpcReqDuration,
		Help:    GrpcReqDurationHelp,
		Labels:  []string{ReqProtocolLable, RespHandlerLable, RespCodeLable},
		Buckets: GgsBuckets,
	}); err != nil {
		return err
	}

	return nil
}

func enableConsumerMetrics() error {
	//REST
	if err := CreateCounter(CounterOpts{
		Name:   ClientReqQPS,
		Help:   ClientReqQPSHelp,
		Labels: []string{RemoteLable, ReqProtocolLable, RespUriLable, RespCodeLable},
	}); err != nil {
		return err
	}

	if err := CreateHistogram(HistogramOpts{
		Name:    ClientReqDuration,
		Help:    ClientReqDurationHelp,
		Labels:  []string{RemoteLable, ReqProtocolLable, RespUriLable, RespCodeLable},
		Buckets: GgsBuckets,
	}); err != nil {
		return err
	}

	//GRPC
	if err := CreateCounter(CounterOpts{
		Name:   ClientGrpcReqQPS,
		Help:   ClientGrpcReqQPSHelp,
		Labels: []string{RemoteLable, ReqProtocolLable, RespHandlerLable, RespCodeLable},
	}); err != nil {
		return err
	}

	if err := CreateHistogram(HistogramOpts{
		Name:    ClientGrpcReqDuration,
		Help:    ClientGrpcReqDurationHelp,
		Labels:  []string{RemoteLable, ReqProtocolLable, RespHandlerLable, RespCodeLable},
		Buckets: GgsBuckets,
	}); err != nil {
		return err
	}

	return nil
}

func GetRegistryInstances() (instances []string) {
	la := archaius.GetString("ggs.protocols.rest.listenAddress", "")
	if len(la) == 0 {
		qlog.Errorf("ggs.protocols.rest.listenAddress is not set")
		return nil
	}
	las := strings.Split(la, ":")
	if len(las) <= 1 {
		qlog.Errorf("ggs.protocols.rest.listenAddress is wrong  %s", la)
		return nil
	}

	ip := iputil.GetLocalIP()
	if len(ip) == 0 {
		qlog.Errorf("get localup failed")
		return
	}
	instance := ip + ":" + las[1]
	instances = append(instances, instance)

	return
}

func enableAutoRegistryMetrics() error {
	instances := GetRegistryInstances()
	if instances == nil {
		return errors.New("metircs registry instances not available")
	}

	type RegistryMetrics struct {
		Token        string   `json:"token"`
		Servicename  string   `json:"servicename"`
		Servicetype  string   `json:"servicetype"`
		Instancelist []string `json:"instancelist"`
		Tagnames     []string `json:"tagnames"`
	}

	servicename := archaius.GetString("service.name", "ggs")
	serviceverion := archaius.GetString("service.version", "")
	reqInfo := RegistryMetrics{
		Token:        archaius.GetString("ggs.metrics.autometrics.token", "80D2A4851C90AB1CC0842D55F409C518D"),
		Servicename:  archaius.GetString("ggs.metrics.autometrics.servicename", "other_exporter"),
		Servicetype:  archaius.GetString("ggs.metrics.autometrics.servicetype", "prometheus_service"),
		Instancelist: instances,
		Tagnames:     []string{"servicename=" + servicename, "serviceverion=" + serviceverion},
	}

	reqData, err := json.Marshal(reqInfo)
	if err != nil {
		return errors.Wrap(err, "marshal reqInfo failed")
	}
	url := archaius.GetString("ggs.metrics.autometrics.url", "")
	if url == "" {
		return errors.New("metrics url empty")
	}

	cli := &http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqData))
	if err != nil {
		return errors.Wrap(err, "new request failed")
	}

	resp, err := cli.Do(req)
	//resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqData))
	if err != nil {
		return errors.Wrap(err, "req post failed")
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "read resp body failed")
	}

	OutRespInfo(respBody)
	return nil
}

func DeAutoRegistryMetrics() {
	instances := GetRegistryInstances()
	if instances == nil {
		return
	}

	type DeMetrics struct {
		Token        string   `json:"token"`
		Instancelist []string `json:"instancelist"`
	}

	reqInfo := DeMetrics{
		Token:        archaius.GetString("ggs.metrics.autometrics.token", "80D2A4851C90AB1CC0842D55F409C518D"),
		Instancelist: instances,
	}

	reqData, err := json.Marshal(reqInfo)
	if err != nil {
		qlog.Errorf("Marshal reqInfo err:", err.Error())
		return
	}
	deurl := archaius.GetString("ggs.metrics.autometrics.deurl", "")
	if deurl == "" {
		qlog.Errorf("metrics url empty")
		return
	}

	cli := &http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest("POST", deurl, bytes.NewBuffer(reqData))
	if err != nil {
		qlog.Errorf("new request failed, err:", err.Error())
		return
	}

	resp, err := cli.Do(req)
	if err != nil {
		qlog.Errorf("req post failed, err:", err.Error())
		return
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		qlog.Errorf("read resp body ", err.Error())
		return
	}

	OutRespInfo(respBody)
}

func OutRespInfo(respBody []byte) {
	type InstancsInfo struct {
		Status   string
		Msg      string
		Instance string
	}
	type RespInfo struct {
		Total       int
		Requestid   string
		Servicetype string
		Instances   []InstancsInfo
	}

	var respInfo RespInfo
	if yaml.Unmarshal(respBody, &respInfo) != nil {
		qlog.Errorf("unmashal failed")
		return
	}
	for _, in := range respInfo.Instances {
		qlog.Infof("auto metrics info: status:%s  msg:%s ", in.Status, in.Msg)
	}
}

func enableRedisMetrics() error {
	if err := CreateCounter(CounterOpts{
		Name:   RedisReqCount,
		Help:   RedisReqCountHelp,
		Labels: []string{RedisRedisName, RedisReqCMD, RedisRespStatus},
	}); err != nil {
		return err
	}

	if err := CreateHistogram(HistogramOpts{
		Name:    RedisReqDurationSecond,
		Help:    RedisReqDurationSecondHelp,
		Labels:  []string{RedisRedisName, RedisReqCMD, RedisRespStatus},
		Buckets: GgsBuckets,
	}); err != nil {
		return err
	}

	return nil
}
