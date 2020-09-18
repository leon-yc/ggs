package servicecomb

import (
	"errors"
	"strings"

	"github.com/leon-yc/ggs/internal/core/config"
	"github.com/leon-yc/ggs/pkg/qlog"
	"github.com/go-chassis/go-archaius"
)

// constant for route rule keys
const (
	DarkLaunchKey      = "^ggs\\.darklaunch\\.policy\\."
	DarkLaunchKeyV2    = "^servicecomb\\.routeRule\\."
	DarkLaunchPrefix   = "ggs.darklaunch.policy."
	DarkLaunchPrefixV2 = "servicecomb.routeRule."
	DarkLaunchTypeRule = "RULE"
	DarkLaunchTypeRate = "RATE"
)

//MergeLocalAndRemoteConfig get router config from archaius,
//including local file,memory and config center
func MergeLocalAndRemoteConfig() (map[string][]*config.RouteRule, error) {
	destinations := make(map[string][]*config.RouteRule, 0)
	//set config from file first
	for k, v := range config.OldRouterDefinition.Destinations {
		destinations[k] = v
	}
	//then get config from archaius and simply overwrite rule from file
	ruleV1Map := make(map[string]interface{})
	ruleV2Map := make(map[string]interface{})
	configMap := archaius.GetConfigs()
	//filter out key:value pairs which are not route rules
	prepareRule(configMap, ruleV1Map, ruleV2Map)
	rules, e := processV1Rule(ruleV1Map, destinations)
	if e != nil {
		return rules, e
	}
	routeRules, e := processV2Rule(ruleV2Map, destinations)
	if e != nil {
		return routeRules, e
	}
	return destinations, nil
}

func processV2Rule(ruleV2Map map[string]interface{}, destinations map[string][]*config.RouteRule) (map[string][]*config.RouteRule, error) {
	for k, v := range ruleV2Map {
		value, ok := v.(string)
		if !ok {
			return nil, errors.New("route rule is not a yaml string format, please check the configuration in config center")
		}

		service := strings.Replace(k, DarkLaunchPrefixV2, "", 1)
		r, err := config.NewServiceRule(value)
		if err != nil {
			qlog.Error("convert failed: " + err.Error())
		}
		destinations[service] = r.Value()
	}
	return nil, nil
}

func processV1Rule(ruleV1Map map[string]interface{}, destinations map[string][]*config.RouteRule) (map[string][]*config.RouteRule, error) {
	for k, v := range ruleV1Map {
		value, ok := v.(string)
		if !ok {
			return nil, errors.New("route rule is not a json string format, please check the configuration in config center")
		}

		service := strings.Replace(k, DarkLaunchPrefix, "", 1)
		r, err := ConvertJSON2RouteRule(value)
		if err != nil {
			qlog.Error("convert failed: " + err.Error())
		}
		destinations[service] = r
	}
	return nil, nil
}

func prepareRule(configMap map[string]interface{}, ruleV1Map map[string]interface{}, ruleV2Map map[string]interface{}) {
	for k, v := range configMap {
		if strings.HasPrefix(k, DarkLaunchPrefix) {
			ruleV1Map[k] = v
			continue
		}
		if strings.HasPrefix(k, DarkLaunchPrefixV2) {
			ruleV2Map[k] = v
			continue
		}
	}
}
