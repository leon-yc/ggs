package consul

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/leon-yc/ggs/pkg/qlog"

	qudiscovery "github.com/leon-gopher/discovery"
	"github.com/leon-gopher/discovery/errors"
	quregistry "github.com/leon-gopher/discovery/registry"
	"github.com/leon-yc/ggs/internal/core/config"
	chregistry "github.com/leon-yc/ggs/internal/core/registry"
	"github.com/leon-yc/ggs/internal/pkg/util/iputil"
)

const (
	ServiceCenter     = "consul"
	ConsulDegratePath = "/data/cache/ggs-consul/"
)

func init() {
	chregistry.InstallRegistrator(ServiceCenter, NewRegistrator)
	chregistry.InstallServiceDiscovery(ServiceCenter, NewServiceDiscovery)
}

type Registrator struct {
	Name         string
	deRegistors  []qudiscovery.ServiceRegister
	microService *chregistry.MicroService
}

//NewRegistrator new Service center registrator
func NewRegistrator(options chregistry.Options) chregistry.Registrator {
	return &Registrator{
		Name: "consul",
	}
}

func (r *Registrator) RegisterService(microService *chregistry.MicroService) (string, error) {
	r.microService = microService
	return microService.ServiceName, nil
}

func (re *Registrator) RegisterServiceInstance(sid string, instance *chregistry.MicroServiceInstance) (string, error) {
	if len(instance.EndpointsMap) == 0 {
		return "", errors.New("endpoints is empty")
	}

	consulAddr := config.GetRegistratorAddress()

	//r, err := qudiscovery.NewRegistryWithConsul(consulAddr)
	r, err := qudiscovery.NewRegistryWithConsulAndFile(consulAddr, ConsulDegratePath)
	if err != nil {
		//handle err
		return "", err
	}

	//map[]
	//注册服务名规则：
	// 1. 只注册一个服务,而且是grpc协议,注册的服务名不加 -grpc,
	// 2. 同时注册rest,grpc服务,grpc的服务需要加上 -grpc
	//2019-11-19 修改规则：rest不加后缀，其他协议的服务，要在服务名后面加上协议名，比如：
	instanceID := ""
	for k, v := range instance.EndpointsMap {
		serviceName := config.MicroserviceDefinition.ServiceDescription.Name
		if k != "rest" {
			serviceName = config.MicroserviceDefinition.ServiceDescription.Name + "-" + k
		}

		ipPort := strings.Split(v, ":")
		if len(ipPort) != 2 {
			return "", errors.Errorf("failed to get port %#v ", instance.EndpointsMap)
		}

		//使用传入的endpoint,假如没有传入,就用获取的ip
		var ip string
		if len(ipPort[0]) != 0 {
			ip = ipPort[0]
		} else {
			ip = iputil.GetLocalIP()
		}
		instanceID = ip

		port, err := strconv.Atoi(ipPort[1])
		if err != err {
			return "", errors.Errorf("recover port failed %s ", ipPort[1])
		}

		tags := []string{
			k,
			fmt.Sprintf("%s:%s", re.microService.Framework.Name, re.microService.Framework.Version),
			config.MicroserviceDefinition.ServiceDescription.Environment,
			config.MicroserviceDefinition.ServiceDescription.Version,
		}
		//注册服务
		deRegistorService, err := r.Register(&quregistry.Service{
			//服务名: 建议ops项目名，不能使用下换线且任何非url safe的字符
			Name: serviceName,
			//服务注册ip地址
			IP: ip,
			//服务端口
			Port: port,
			Tags: tags,
			Meta: map[string]string{
				"protoc": k,
			},
		},
		)

		//annotation: 注册多个，有失败就返回
		if err != nil {
			qlog.Errorf("registor failed v: err:%s", err.Error())
			return "", err
		}

		re.deRegistors = append(re.deRegistors, deRegistorService)
	}

	return instanceID, nil
}

func (r *Registrator) RegisterServiceAndInstance(microService *chregistry.MicroService, instance *chregistry.MicroServiceInstance) (string, string, error) {
	return "", "", nil
}

func (r *Registrator) Heartbeat(microServiceID, microServiceInstanceID string) (bool, error) {
	return false, nil
}

func (r *Registrator) AddDependencies(dep *chregistry.MicroServiceDependency) error {
	return nil
}

func (r *Registrator) UnRegisterMicroServiceInstance(microServiceID, microServiceInstanceID string) (err error) {
	for _, v := range r.deRegistors {
		err = v.Deregister()
		if err != nil {
			qlog.Errorf("deregister service failed ")
		}

		qlog.Info("deregister service success")
	}

	return nil
}

func (r *Registrator) UpdateMicroServiceInstanceStatus(microServiceID, microServiceInstanceID, status string) error {
	return nil
}

func (r *Registrator) UpdateMicroServiceProperties(microServiceID string, properties map[string]string) error {
	return nil
}

func (r *Registrator) UpdateMicroServiceInstanceProperties(microServiceID, microServiceInstanceID string, properties map[string]string) error {
	return nil
}

func (r *Registrator) AddSchemas(microServiceID, schemaName, schemaInfo string) error {
	return nil
}

func (r *Registrator) Close() error {
	return nil
}
