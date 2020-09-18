package eventlistener

import (
	"github.com/go-chassis/go-archaius"
	"github.com/go-chassis/go-archaius/event"
)

//RegisterKeys registers a config key to the archaius
func RegisterKeys(eventListener event.Listener, keys ...string) {

	archaius.RegisterListener(eventListener, keys...)
}

//Init is a function
func Init() {
	//采用平滑重启的方式来做热更新，所以暂时去掉配置模块自动加载功能
	//qpsEventListener := &QPSEventListener{}
	//circuitBreakerEventListener := &CircuitBreakerEventListener{}
	//lbEventListener := &LoadbalancingEventListener{}
	//
	//RegisterKeys(qpsEventListener, QPSLimitKey)
	//RegisterKeys(circuitBreakerEventListener, ConsumerFallbackKey, ConsumerFallbackPolicyKey, ConsumerIsolationKey, ConsumerCircuitbreakerKey)
	//RegisterKeys(lbEventListener, LoadBalanceKey)
	//RegisterKeys(&LoggerEventListener{}, LoggerLevelKey)

}
