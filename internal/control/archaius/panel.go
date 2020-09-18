package archaius

import (
	"github.com/leon-yc/ggs/internal/control"
	"github.com/leon-yc/ggs/internal/core/common"
	"github.com/leon-yc/ggs/internal/core/config"
	"github.com/leon-yc/ggs/internal/core/config/model"
	"github.com/leon-yc/ggs/internal/core/invocation"
	"github.com/leon-yc/ggs/internal/core/qpslimiter"
	"github.com/leon-yc/ggs/third_party/forked/afex/hystrix-go/hystrix"
	"github.com/go-chassis/go-archaius"
)

//Panel pull configs from archaius
type Panel struct {
}

func newPanel(options control.Options) control.Panel {
	SaveToLBCache(config.GetLoadBalancing())
	SaveToCBCache(config.GetHystrixConfig())
	return &Panel{}
}

//GetCircuitBreaker return command , and circuit breaker settings
func (p *Panel) GetCircuitBreaker(inv invocation.Invocation, serviceType string) (string, hystrix.CommandConfig) {
	key := GetCBCacheKey(inv.MicroServiceName, serviceType)
	command := control.NewCircuitName(serviceType, config.GetHystrixConfig().CircuitBreakerProperties.Scope, inv)
	c, ok := CBConfigCache.Get(key)
	if !ok {
		c, _ := CBConfigCache.Get(serviceType)
		return command, c.(hystrix.CommandConfig)

	}
	return command, c.(hystrix.CommandConfig)
}

//GetLoadBalancing get load balancing config
func (p *Panel) GetLoadBalancing(inv invocation.Invocation) control.LoadBalancingConfig {
	c, ok := LBConfigCache.Get(inv.MicroServiceName)
	if !ok {
		c, ok := LBConfigCache.Get("")
		if !ok {
			return DefaultLB

		}
		return c.(control.LoadBalancingConfig)

	}
	return c.(control.LoadBalancingConfig)

}

//GetRateLimiting get rate limiting config
func (p *Panel) GetRateLimiting(inv invocation.Invocation, serviceType string) control.RateLimitingConfig {
	rl := control.RateLimitingConfig{}
	rl.Enabled = archaius.GetBool("ggs.flowcontrol."+serviceType+".qps.enabled", false)
	if serviceType == common.Consumer {
		keys := qpslimiter.GetConsumerKey(inv.SourceMicroService, inv.MicroServiceName, inv.SchemaID, inv.OperationID)
		rl.Rate, rl.Key = qpslimiter.GetQPSTrafficLimiter().GetQPSRateWithPriority(
			keys.OperationQualifiedName, keys.MicroServiceName)
	} else {
		keys := qpslimiter.GetProviderKey(inv.OperationID)
		rl.Rate, rl.Key = qpslimiter.GetQPSTrafficLimiter().GetQPSRateWithPriority(
			keys.Api, keys.Global)
	}

	return rl
}

//GetFaultInjection get Fault injection config
func (p *Panel) GetFaultInjection(inv invocation.Invocation) model.Fault {
	return model.Fault{}

}

//GetEgressRule get egress config
func (p *Panel) GetEgressRule() []control.EgressConfig {
	return []control.EgressConfig{}
}

func init() {
	control.InstallPlugin("archaius", newPanel)
}
