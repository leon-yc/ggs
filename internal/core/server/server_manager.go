package server

import (
	"fmt"
	"net"

	"github.com/leon-yc/ggs/internal/core/common"
	"github.com/leon-yc/ggs/internal/core/config"
	"github.com/leon-yc/ggs/internal/core/config/model"
	"github.com/leon-yc/ggs/internal/core/registry"
	ggsTLS "github.com/leon-yc/ggs/internal/core/tls"
	"github.com/leon-yc/ggs/internal/pkg/runtime"
	"github.com/leon-yc/ggs/internal/pkg/util"
	"github.com/leon-yc/ggs/internal/pkg/util/iputil"
	"github.com/leon-yc/ggs/pkg/qlog"
)

//NewFunc returns a ProtocolServer
type NewFunc func(Options) ProtocolServer

var serverPlugins = make(map[string]NewFunc)
var servers = make(map[string]ProtocolServer)

//InstallPlugin For developer
func InstallPlugin(protocol string, newFunc NewFunc) {
	serverPlugins[protocol] = newFunc
	qlog.Trace("Installed Server Plugin, protocol:" + protocol)
}

//GetServerFunc returns the server function
func GetServerFunc(protocol string) (NewFunc, error) {
	f, ok := serverPlugins[protocol]
	if !ok {
		return nil, fmt.Errorf("unknown protocol server [%s]", protocol)
	}
	return f, nil
}

//GetServer return the server based on protocol
func GetServer(protocol string) (ProtocolServer, error) {
	s, ok := servers[protocol]
	if !ok {
		return nil, fmt.Errorf("[%s] server isn't running ", protocol)
	}
	return s, nil
}

//GetServers returns the map of servers
func GetServers() map[string]ProtocolServer {
	return servers
}

//ErrRuntime is an error channel, if it receive any signal will cause graceful shutdown of go chassis, process will exit
var ErrRuntime = make(chan error)

//StartServer starting the server
func StartServer() error {
	for name, server := range servers {
		qlog.Info("starting server " + name + "...")
		err := server.Start()
		if err != nil {
			qlog.Errorf("servers failed to start, err %s", err)
			return fmt.Errorf("can not start [%s] server,%s", name, err.Error())
		}
		qlog.Info(name + " server start success")
	}
	qlog.Info("All server Start Completed")

	return nil
}

//UnRegistrySelfInstances this function removes the self instance
func UnRegistrySelfInstances() error {
	if err := registry.DefaultRegistrator.UnRegisterMicroServiceInstance(runtime.ServiceID, runtime.InstanceID); err != nil {
		qlog.Errorf("StartServer() UnregisterMicroServiceInstance failed, sid/iid: %s/%s: %s",
			runtime.ServiceID, runtime.InstanceID, err)
		return err
	}
	return nil
}

//Init initializes
func Init(listens map[string]net.Listener) error {
	qlog.Tracef("listens: %v", listens)
	var err error
	for k, v := range config.GlobalDefinition.Ggs.Protocols {
		var listen net.Listener
		if listens != nil {
			listen = listens[v.Listen]
		}
		if err = initialServer(config.GlobalDefinition.Ggs.Handler.Chain.Provider, v, k, listen); err != nil {
			qlog.Info(err)
			return err
		}
	}
	return nil

}

func initialServer(providerMap map[string]string, p model.Protocol, name string, listen net.Listener) error {
	protocolName, _, err := util.ParsePortName(name)
	if err != nil {
		return err
	}
	qlog.Tracef("Init server [%s], protocol is [%s]", name, protocolName)
	f, err := GetServerFunc(protocolName)
	if err != nil {
		return fmt.Errorf("do not support [%s] server", name)
	}

	sslTag := name + "." + common.Provider
	tlsConfig, sslConfig, err := ggsTLS.GetTLSConfigByService("", name, common.Provider)
	if err != nil {
		if !ggsTLS.IsSSLConfigNotExist(err) {
			return err
		}
	} else {
		if listen != nil {
			return fmt.Errorf("grace restart and tls are not supported at the same time")
		}
		qlog.Warnf("%s TLS mode, verify peer: %t, cipher plugin: %s.",
			sslTag, sslConfig.VerifyPeer, sslConfig.CipherPlugin)
	}

	if p.Listen == "" {
		if p.Advertise != "" {
			p.Listen = p.Advertise
		} else {
			p.Listen = iputil.DefaultEndpoint4Protocol(name)
		}
	}

	chainName := common.DefaultChainName
	if _, ok := providerMap[name]; ok {
		chainName = name
	}

	var s ProtocolServer
	o := Options{
		Address:            p.Listen,
		Listen:             listen,
		ProtocolServerName: name,
		ChainName:          chainName,
		TLSConfig:          tlsConfig,
		BodyLimit:          config.GlobalDefinition.Ggs.Transport.MaxBodyBytes["rest"],
		EnableGrpcurl:      p.EnableGrpcurl,
	}
	s = f(o)
	servers[name] = s
	return nil
}
