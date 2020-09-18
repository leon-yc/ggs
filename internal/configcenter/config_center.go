package configcenter

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/leon-yc/ggs/internal/core/common"
	"github.com/leon-yc/ggs/internal/core/config"
	"github.com/leon-yc/ggs/internal/core/endpoint"
	ggsTLS "github.com/leon-yc/ggs/internal/core/tls"
	configclient "github.com/go-chassis/go-chassis-config"

	"github.com/leon-yc/ggs/internal/pkg/runtime"
	"github.com/leon-yc/ggs/pkg/qlog"
	"github.com/go-chassis/go-archaius"
)

const (
	//Name is a variable of type string
	Name = "configcenter"
	//DefaultConfigCenter is config center
	DefaultConfigCenter = "config_center"
)

//ErrRefreshMode means config is mis used
var (
	ErrRefreshMode      = errors.New("refreshMode must be 0 or 1")
	ErrRegistryDisabled = errors.New("discovery is disabled")
)

// Init initialize config center
func Init() error {
	configCenterURL, err := GetConfigCenterEndpoint()
	if err != nil {
		qlog.Trace("can not get config server endpoint: " + err.Error())
		return nil
	}

	var enableSSL bool
	tlsConfig, tlsError := getTLSForClient(configCenterURL)
	if tlsError != nil {
		qlog.Errorf("Get %s.%s TLS config failed, err:[%s]", Name, common.Consumer, tlsError.Error())
		return tlsError
	}

	/*This condition added because member discovery can have multiple ip's with IsHTTPS
	having both true and false value.*/
	if tlsConfig != nil {
		enableSSL = true
	}

	TenantName := config.GetConfigCenterConf().TenantName
	if TenantName == "" {
		TenantName = common.DefaultTenant
	}
	interval := config.GetConfigCenterConf().RefreshInterval
	if interval == 0 {
		interval = 30
	}

	err = initConfigCenter(configCenterURL, TenantName,
		enableSSL, tlsConfig, interval)
	if err != nil {
		qlog.Error("failed to init config center" + err.Error())
		return err
	}

	qlog.Warn("config center init success")
	return nil
}

//GetConfigCenterEndpoint will read local config center uri first, if there is not,
// it will try to discover config center from registry
func GetConfigCenterEndpoint() (string, error) {
	configCenterURL := config.GetConfigCenterConf().ServerURI
	if configCenterURL == "" {
		if false /*registry.DefaultServiceDiscoveryService != nil*/ { //TODO: register ConfigCenter
			qlog.Trace("find config server in registry")
			ccURL, err := endpoint.GetEndpointFromServiceCenter("default", "ConfigCenter", "latest")
			if err != nil {
				qlog.Warn("failed to find config center endpoints, err: " + err.Error())
				return "", err
			}
			configCenterURL = ccURL
		} else {
			return "", ErrRegistryDisabled
		}

	}

	return configCenterURL, nil
}

func getTLSForClient(configCenterURL string) (*tls.Config, error) {
	if !strings.Contains(configCenterURL, "://") {
		return nil, nil
	}
	ccURL, err := url.Parse(configCenterURL)
	if err != nil {
		qlog.Error("Error occurred while parsing config center Server Uri" + err.Error())
		return nil, err
	}
	if ccURL.Scheme == common.HTTP {
		return nil, nil
	}

	sslTag := Name + "." + common.Consumer
	tlsConfig, sslConfig, err := ggsTLS.GetTLSConfigByService(Name, "", common.Consumer)
	if err != nil {
		if ggsTLS.IsSSLConfigNotExist(err) {
			return nil, fmt.Errorf("%s TLS mode, but no ssl config", sslTag)
		}
		return nil, err
	}
	qlog.Warnf("%s TLS mode, verify peer: %t, cipher plugin: %s.",
		sslTag, sslConfig.VerifyPeer, sslConfig.CipherPlugin)

	return tlsConfig, nil
}

func initConfigCenter(ccEndpoint, tenantName string,
	enableSSL bool, tlsConfig *tls.Config, interval int) error {

	refreshMode := archaius.GetInt("ggs.config.client.refreshMode", common.DefaultRefreshMode)
	if refreshMode != 0 && refreshMode != 1 {
		qlog.Error(ErrRefreshMode.Error())
		return ErrRefreshMode
	}

	clientType := config.GlobalDefinition.Ggs.Config.Client.Type
	if clientType == "" {
		clientType = DefaultConfigCenter

	}

	var ccObj = &archaius.RemoteInfo{
		DefaultDimension: map[string]string{
			configclient.LabelApp:         runtime.App,
			configclient.LabelService:     runtime.ServiceName,
			configclient.LabelVersion:     runtime.Version,
			configclient.LabelEnvironment: runtime.Environment,
		},
		URL:             ccEndpoint,
		TenantName:      tenantName,
		EnableSSL:       enableSSL,
		TLSConfig:       tlsConfig,
		RefreshMode:     refreshMode,
		RefreshInterval: interval,
		AutoDiscovery:   config.GetConfigCenterConf().Autodiscovery,
		ClientType:      clientType,
		APIVersion:      config.GetConfigCenterConf().APIVersion.Version,
		RefreshPort:     config.GetConfigCenterConf().RefreshPort,
	}

	err := archaius.EnableRemoteSource(ccObj, nil)

	if err != nil {
		return err
	}

	if err := refreshGlobalConfig(); err != nil {
		qlog.Error("failed to refresh global config for lb and cb:" + err.Error())
		return err
	}
	return nil
}

func refreshGlobalConfig() error {
	err := config.ReadHystrixFromArchaius()
	if err != nil {
		return err
	}
	return config.ReadLBFromArchaius()
}
