package config

import (
	"github.com/go-chassis/go-archaius"
	"strings"
)

// GetServiceDiscoveryType returns the Type of SD registry
func GetServiceDiscoveryType() string {
	if GlobalDefinition.Ggs.Service.Registry.ServiceDiscovery.Type != "" {
		return GlobalDefinition.Ggs.Service.Registry.ServiceDiscovery.Type
	}
	return GlobalDefinition.Ggs.Service.Registry.Type
}

// GetServiceDiscoveryAddress returns the Address of SD registry
func GetServiceDiscoveryAddress() string {
	if GlobalDefinition.Ggs.Service.Registry.ServiceDiscovery.Address != "" {
		return GlobalDefinition.Ggs.Service.Registry.ServiceDiscovery.Address
	}
	if GlobalDefinition.Ggs.Service.Registry.Address == "" {
		e := strings.ToLower(archaius.GetString("service.environment", ""))
		if e == "prd" || e == "pre" {
			GlobalDefinition.Ggs.Service.Registry.Address = DefaultSRAddressPRD
		} else {
			GlobalDefinition.Ggs.Service.Registry.Address = DefaultSRAddressQA
		}
	}
	return GlobalDefinition.Ggs.Service.Registry.Address
}

// GetServiceDiscoveryRefreshInterval returns the RefreshInterval of SD registry
func GetServiceDiscoveryRefreshInterval() string {
	if GlobalDefinition.Ggs.Service.Registry.ServiceDiscovery.RefreshInterval != "" {
		return GlobalDefinition.Ggs.Service.Registry.ServiceDiscovery.RefreshInterval
	}
	return GlobalDefinition.Ggs.Service.Registry.RefreshInterval
}

// GetServiceDiscoveryWatch returns the Watch of SD registry
func GetServiceDiscoveryWatch() bool {
	if GlobalDefinition.Ggs.Service.Registry.ServiceDiscovery.Watch {
		return GlobalDefinition.Ggs.Service.Registry.ServiceDiscovery.Watch
	}
	return GlobalDefinition.Ggs.Service.Registry.Watch
}

// GetServiceDiscoveryTenant returns the Tenant of SD registry
func GetServiceDiscoveryTenant() string {
	if GlobalDefinition.Ggs.Service.Registry.ServiceDiscovery.Tenant != "" {
		return GlobalDefinition.Ggs.Service.Registry.ServiceDiscovery.Tenant
	}
	return GlobalDefinition.Ggs.Service.Registry.Tenant
}

// GetServiceDiscoveryAPIVersion returns the APIVersion of SD registry
func GetServiceDiscoveryAPIVersion() string {
	if GlobalDefinition.Ggs.Service.Registry.ServiceDiscovery.APIVersion.Version != "" {
		return GlobalDefinition.Ggs.Service.Registry.ServiceDiscovery.APIVersion.Version
	}
	return GlobalDefinition.Ggs.Service.Registry.APIVersion.Version
}

// GetServiceDiscoveryDisable returns the Disable of SD registry
func GetServiceDiscoveryDisable() bool {
	return archaius.GetBool("ggs.service.registry.serviceDiscovery.disabled", false)
}

// GetServiceDiscoveryHealthCheck returns the HealthCheck of SD registry
func GetServiceDiscoveryHealthCheck() bool {
	if b := archaius.GetBool("ggs.service.registry.serviceDiscovery.healthCheck", false); b {
		return b
	}
	return archaius.GetBool("ggs.service.registry.healthCheck", false)
}

// DefaultConfigPath set the default config path
const DefaultConfigPath = "/etc/.kube/config"

// GetServiceDiscoveryConfigPath returns the configpath of SD registry
func GetServiceDiscoveryConfigPath() string {
	if GlobalDefinition.Ggs.Service.Registry.ServiceDiscovery.ConfigPath != "" {
		return GlobalDefinition.Ggs.Service.Registry.ServiceDiscovery.ConfigPath
	}
	return DefaultConfigPath
}
