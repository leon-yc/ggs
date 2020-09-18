package client

import (
	"net/url"

	"github.com/leon-yc/ggs/internal/pkg/scclient/proto"
	"github.com/leon-yc/ggs/pkg/qlog"
)

func getProtocolMap(eps []string) map[string]string {
	m := make(map[string]string)
	for _, ep := range eps {
		u, err := url.Parse(ep)
		if err != nil {
			qlog.Error("url err: " + err.Error())
			continue
		}
		m[u.Scheme] = u.Host
	}
	return m
}

//RegroupInstances organize raw data to better format
func RegroupInstances(keys []*proto.FindService, response proto.BatchFindInstancesResponse) map[string][]*proto.MicroServiceInstance {
	instanceMap := make(map[string][]*proto.MicroServiceInstance, 0)
	if response.Services != nil {
		for _, result := range response.Services.Updated {
			if len(result.Instances) == 0 {
				continue
			}
			for _, instance := range result.Instances {
				instance.ServiceName = keys[result.Index].Service.ServiceName
				instance.App = keys[result.Index].Service.AppId
				instances, ok := instanceMap[instance.ServiceName]
				if !ok {
					instances = make([]*proto.MicroServiceInstance, 0)
					instanceMap[instance.ServiceName] = instances
				}
				instanceMap[instance.ServiceName] = append(instances, instance)
			}

		}
	}
	return instanceMap
}
