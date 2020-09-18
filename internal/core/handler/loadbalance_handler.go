package handler

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/leon-yc/ggs/internal/control"
	"github.com/leon-yc/ggs/internal/core/common"
	"github.com/leon-yc/ggs/internal/core/invocation"
	"github.com/leon-yc/ggs/internal/core/loadbalancer"
	backoffUtil "github.com/leon-yc/ggs/internal/pkg/backoff"
	"github.com/leon-yc/ggs/internal/pkg/util"
	"github.com/leon-yc/ggs/pkg/qlog"
	"github.com/cenkalti/backoff"
	"github.com/go-chassis/go-archaius"
)

// LBHandler loadbalancer handler struct
type LBHandler struct {
	failureMap map[string]bool
}

func (lb *LBHandler) getEndpoint(i *invocation.Invocation, lbConfig control.LoadBalancingConfig) (string, error) {
	if i.NoDiscovery {
		// do not using discovery, so skiping consul
		ep := i.MicroServiceName
		if i.RouteType == common.RouteSidecar && archaius.GetBool("ggs.sidecar.enabled", false) {
			i.Ctx = common.WithContext(i.Ctx, common.HeaderXSidecar, strings.ReplaceAll(i.MicroServiceName, "_", "-"))
			ep = common.SidecarAddress
		}
		return ep, nil
	}

	var strategyFun func() loadbalancer.Strategy
	var err error
	if i.Strategy == "" {
		i.Strategy = lbConfig.Strategy
		strategyFun, err = loadbalancer.GetStrategyPlugin(i.Strategy)
		if err != nil {
			qlog.Errorf("lb error [%s] because of [%s]", loadbalancer.LBError{
				Message: "Get strategy [" + i.Strategy + "] failed."}.Error(), err.Error())
		}
	} else {
		strategyFun, err = loadbalancer.GetStrategyPlugin(i.Strategy)
		if err != nil {
			qlog.Errorf("lb error [%s] because of [%s]", loadbalancer.LBError{
				Message: "Get strategy [" + i.Strategy + "] failed."}.Error(), err.Error())
		}
	}
	if len(i.Filters) == 0 {
		i.Filters = lbConfig.Filters
	}

	s, err := loadbalancer.BuildStrategy(i, strategyFun())
	if err != nil {
		return "", err
	}

	ins, err := s.Pick()
	if err != nil {
		lbErr := loadbalancer.LBError{Message: err.Error()}
		return "", lbErr
	}

	var ep string
	if i.Protocol == "" {
		i.Protocol = archaius.GetString("ggs.references."+i.MicroServiceName+".transport", ins.DefaultProtocol)
	}
	if i.Protocol == "" {
		for k := range ins.EndpointsMap {
			i.Protocol = k
			break
		}
	}
	protocolServer := util.GenProtoEndPoint(i.Protocol, i.Port)
	ep, ok := ins.EndpointsMap[protocolServer]
	if !ok {
		errStr := fmt.Sprintf(
			"No available instance for protocol server [%s] , microservice: %s has %v",
			protocolServer, i.MicroServiceName, ins.EndpointsMap)
		lbErr := loadbalancer.LBError{Message: errStr}
		qlog.Errorf(lbErr.Error())
		return "", lbErr
	}
	return ep, nil
}

// Handle to handle the load balancing
func (lb *LBHandler) Handle(chain *Chain, i *invocation.Invocation, cb invocation.ResponseCallBack) {
	lbConfig := control.DefaultPanel.GetLoadBalancing(*i)
	if !lbConfig.RetryEnabled {
		lb.handleWithNoRetry(chain, i, lbConfig, cb)
	} else {
		lb.handleWithRetry(chain, i, lbConfig, cb)
	}
}

func (lb *LBHandler) handleWithNoRetry(chain *Chain, i *invocation.Invocation, lbConfig control.LoadBalancingConfig, cb invocation.ResponseCallBack) {
	ep, err := lb.getEndpoint(i, lbConfig)
	if err != nil {
		writeErr(err, cb)
		return
	}

	i.Endpoint = ep
	chain.Next(i, cb)
}

func (lb *LBHandler) handleWithRetry(chain *Chain, i *invocation.Invocation, lbConfig control.LoadBalancingConfig, cb invocation.ResponseCallBack) {
	retryOnSame := lbConfig.RetryOnSame
	retryOnNext := lbConfig.RetryOnNext
	handlerIndex := i.HandlerIndex
	var invResp *invocation.Response
	var reqBytes []byte
	if req, ok := i.Args.(*http.Request); ok {
		if req != nil {
			if req.Body != nil {
				reqBytes, _ = ioutil.ReadAll(req.Body)
			}
		}
	}
	// get retry func
	lbBackoff := backoffUtil.GetBackOff(lbConfig.BackOffKind, lbConfig.BackOffMin, lbConfig.BackOffMax)
	callTimes := 0

	ep, err := lb.getEndpoint(i, lbConfig)
	if err != nil {
		// if get endpoint failed, no need to retry
		writeErr(err, cb)
		return
	}
	operation := func() error {
		i.Endpoint = ep
		callTimes++
		var retErr error
		i.HandlerIndex = handlerIndex

		if _, ok := i.Args.(*http.Request); ok {
			i.Args.(*http.Request).Body = ioutil.NopCloser(bytes.NewBuffer(reqBytes))
		}

		chain.Next(i, func(r *invocation.Response) error {
			if r != nil {
				invResp = r
				//respErr = invResp.Err
				if lb.needRetry(r.Err, r.Status) {
					retErr = fmt.Errorf("need retry")
				}
				return invResp.Err
			}
			return nil
		})

		if callTimes >= retryOnSame+1 {
			if retryOnNext <= 0 {
				return backoff.Permanent(errors.New("retry times expires"))
			}
			ep, err = lb.getEndpoint(i, lbConfig)
			if err != nil {
				// if get endpoint failed, no need to retry
				return backoff.Permanent(err)
			}
			callTimes = 0
			retryOnNext--
		}
		return retErr
	}
	if err := backoff.Retry(operation, lbBackoff); err != nil {
		qlog.Tracef("stop retry , error : %v", err)
	}

	if invResp == nil {
		invResp = &invocation.Response{}
	}
	cb(invResp)
}

// Name returns loadbalancer string
func (lb *LBHandler) Name() string {
	return "loadbalancer"
}

func (lb *LBHandler) needRetry(err error, status int) bool {
	if err != nil {
		if e, ok := err.(*url.Error); ok && e.Timeout() {
			return lb.failureMap["timeout"]
		}
		return true
	}

	if status < 400 {
		return false
	}
	codestr := fmt.Sprintf("http_%d", status)
	return lb.failureMap[codestr]
}

func newLBHandler() Handler {
	statuses := archaius.GetString("ggs.loadbalance.retryCondition", "http_500,http_502,http_503,timeout")
	failureList := strings.Split(statuses, ",")
	failureMap := make(map[string]bool)
	for _, v := range failureList {
		failureMap[strings.TrimSpace(v)] = true
	}

	return &LBHandler{
		failureMap: failureMap,
	}
}
