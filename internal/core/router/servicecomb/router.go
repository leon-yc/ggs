package servicecomb

import (
	"sync"

	"github.com/leon-yc/ggs/pkg/qlog"

	"github.com/leon-yc/ggs/internal/core/config"
	"github.com/leon-yc/ggs/internal/core/router"
	"github.com/go-chassis/go-archaius"
)

var cseRouter *Router

//Router is ggs router service
type Router struct {
	routeRule map[string][]*config.RouteRule
	lock      sync.RWMutex
}

//SetRouteRule set rules
func (r *Router) SetRouteRule(rr map[string][]*config.RouteRule) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.routeRule = rr
}

//FetchRouteRuleByServiceName get rules for service
func (r *Router) FetchRouteRuleByServiceName(service string) []*config.RouteRule {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.routeRule[service]
}

//Init init router config
func (r *Router) Init(o router.Options) error {
	archaius.RegisterListener(&routeRuleEventListener{}, DarkLaunchKey, DarkLaunchKeyV2)
	return r.LoadRules()
}

func newRouter() (router.Router, error) {
	cseRouter = &Router{
		routeRule: make(map[string][]*config.RouteRule, 0),
		lock:      sync.RWMutex{},
	}
	return cseRouter, nil
}

// LoadRules load all the router config
func (r *Router) LoadRules() error {
	configs, err := MergeLocalAndRemoteConfig()
	if err != nil {
		qlog.WithError(err).Error("init route rule failed")
	}

	if router.ValidateRule(configs) {
		r.routeRule = configs
		qlog.WithField("rule", r.routeRule).Debug("load route rule")
	}
	return nil
}

// SetRouteRuleByKey set route rule by key
func (r *Router) SetRouteRuleByKey(k string, rr []*config.RouteRule) {
	r.lock.Lock()
	r.routeRule[k] = rr
	r.lock.Unlock()
}

// DeleteRouteRuleByKey set route rule by key
func (r *Router) DeleteRouteRuleByKey(k string) {
	r.lock.Lock()
	delete(r.routeRule, k)
	r.lock.Unlock()
}

func init() {
	router.InstallRouterService("ggs", newRouter)
}
