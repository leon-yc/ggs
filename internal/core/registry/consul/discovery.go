package consul

import (
	"fmt"

	"github.com/leon-yc/ggs/pkg/qlog"

	client "github.com/leon-yc/ggs/internal/pkg/scclient"
	utiltags "github.com/leon-yc/ggs/internal/pkg/util/tags"

	qudiscovery "github.com/leon-gopher/discovery"
	quregistry "github.com/leon-gopher/discovery/registry"
	"github.com/leon-yc/ggs/internal/core/config"
	chregistry "github.com/leon-yc/ggs/internal/core/registry"
)

type ServiceDiscovery struct {
	Name           string
	registryClient *client.RegistryClient
	opts           client.Options

	r *qudiscovery.Registry //for discovery
}

//NewServiceDiscovery new service center discovery
func NewServiceDiscovery(options chregistry.Options) chregistry.ServiceDiscovery {
	//sco := ToSCOptions(options)
	//r := &client.RegistryClient{}
	//if err := r.Initialize(sco); err != nil {
	//	qlog.Errorf("RegistryClient initialization failed. %s", err)
	//}

	consulAddr := config.GetServiceDiscoveryAddress()
	r, err := qudiscovery.NewRegistryWithConsul(consulAddr)
	if err != nil {
		qlog.Errorf("new discovery object faild,consuladdr: %s err:%s", consulAddr, err.Error())
		return nil
	}

	return &ServiceDiscovery{
		Name: ServiceCenter,
		r:    r,
		//registryClient: r,
		//opts:           sco,
	}
}

func (dis *ServiceDiscovery) GetMicroServiceID(appID, microServiceName, version, env string) (string, error) {
	return microServiceName, nil
}

func (dis *ServiceDiscovery) GetAllMicroServices() ([]*chregistry.MicroService, error) {
	return nil, nil
}

func (dis *ServiceDiscovery) GetMicroService(microServiceID string) (*chregistry.MicroService, error) {
	return nil, nil
}
func (dis *ServiceDiscovery) GetMicroServiceInstances(consumerID, providerID string) ([]*chregistry.MicroServiceInstance, error) {
	//consulAddr := config.GetServiceDiscoveryAddress()
	//client, err := consul.New(consulAddr)
	//if err != nil {
	//	return nil, err
	//}
	//
	//r, err := qudiscovery.NewRegistry(qudiscovery.WithDiscoverys(client), qudiscovery.WithRegisters(client))
	//if err != nil {
	//	return nil, err
	//}
	//
	//srvlist, err := r.LookupServices(providerID)
	//if err != nil {
	//	return nil, err
	//}
	//
	////var microinss []*chregistry.MicroServiceInstance
	//for _, v := range srvlist {
	//	var mi chregistry.MicroServiceInstance
	//
	//	mi.Metadata = v.Meta
	//}

	return nil, nil
}

func (dis *ServiceDiscovery) FindMicroServiceInstances(consumerID, microServiceName string, tags utiltags.Tags) (
	microinss []*chregistry.MicroServiceInstance, err error) {

	var srvlist []*quregistry.Service
	tenant := config.GlobalDefinition.Ggs.Service.Registry.Tenant //tenant -> dc
	if tenant != "" && tenant != "default" {
		opt := quregistry.WithDC(tenant)
		srvlist, err = dis.r.LookupServices(microServiceName, opt) //opt 不能为nil
	} else {
		srvlist, err = dis.r.LookupServices(microServiceName)
	}

	if err != nil {
		return nil, err
	}

	for _, v := range srvlist {
		var mi chregistry.MicroServiceInstance
		mi.EndpointsMap = make(map[string]string)

		mi.Metadata = v.Meta
		mi.HostName = v.IP
		mi.InstanceID = v.ID
		mi.ServiceID = v.ID

		ep := v.IP + ":" + fmt.Sprintf("%d", v.Port)
		mi.EndpointsMap["rest"] = ep
		mi.EndpointsMap["grpc"] = ep

		microinss = append(microinss, &mi)
	}

	return microinss, err
}

func (dis *ServiceDiscovery) AutoSync() {

}

func (dis *ServiceDiscovery) Close() error {
	return nil
}
