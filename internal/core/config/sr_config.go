package config

import (
	"strings"

	"github.com/leon-yc/ggs/internal/core/common"
	"github.com/go-chassis/go-archaius"
)

// constant for service registry parameters
const (
	DefaultSRAddressQA  = "http://10.0.1.101:8500"
	DefaultSRAddressPRD = "http://127.0.0.1:8500"
)

// GetRegistratorType returns the Type of service registry
func GetRegistratorType() string {
	if GlobalDefinition.Ggs.Service.Registry.Registrator.Type != "" {
		return GlobalDefinition.Ggs.Service.Registry.Registrator.Type
	}
	return GlobalDefinition.Ggs.Service.Registry.Type
}

// GetRegistratorAddress returns the Address of service registry
func GetRegistratorAddress() string {
	if GlobalDefinition.Ggs.Service.Registry.Registrator.Address != "" {
		return GlobalDefinition.Ggs.Service.Registry.Registrator.Address
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

// GetRegistratorScope returns the Scope of service registry
func GetRegistratorScope() string {
	if GlobalDefinition.Ggs.Service.Registry.Registrator.Scope == "" {
		GlobalDefinition.Ggs.Service.Registry.Registrator.Scope = common.ScopeFull
	}
	return GlobalDefinition.Ggs.Service.Registry.Scope
}

// GetRegistratorAutoRegister returns the AutoRegister of service registry
func GetRegistratorAutoRegister() string {
	if GlobalDefinition.Ggs.Service.Registry.Registrator.AutoRegister != "" {
		return GlobalDefinition.Ggs.Service.Registry.Registrator.AutoRegister
	}
	return GlobalDefinition.Ggs.Service.Registry.AutoRegister
}

// GetRegistratorTenant returns the Tenant of service registry
func GetRegistratorTenant() string {
	if GlobalDefinition.Ggs.Service.Registry.Registrator.Tenant != "" {
		return GlobalDefinition.Ggs.Service.Registry.Registrator.Tenant
	}
	return GlobalDefinition.Ggs.Service.Registry.Tenant
}

// GetRegistratorAPIVersion returns the APIVersion of service registry
func GetRegistratorAPIVersion() string {
	if GlobalDefinition.Ggs.Service.Registry.Registrator.APIVersion.Version != "" {
		return GlobalDefinition.Ggs.Service.Registry.Registrator.APIVersion.Version
	}
	return GlobalDefinition.Ggs.Service.Registry.APIVersion.Version
}

// GetRegistratorDisable returns the Disable of service registry
func GetRegistratorDisable() bool {
	if b := archaius.GetBool("ggs.service.registry.registrator.disabled", false); b {
		return b
	}
	return archaius.GetBool("ggs.service.registry.disabled", false)
}
