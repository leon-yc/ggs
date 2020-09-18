package config

import "github.com/go-chassis/go-archaius"

// GetContractDiscoveryType returns the Type of contract discovery registry
func GetContractDiscoveryType() string {
	if GlobalDefinition.Ggs.Service.Registry.ContractDiscovery.Type != "" {
		return GlobalDefinition.Ggs.Service.Registry.ContractDiscovery.Type
	}
	return GlobalDefinition.Ggs.Service.Registry.Type
}

// GetContractDiscoveryAddress returns the Address of contract discovery registry
func GetContractDiscoveryAddress() string {
	if GlobalDefinition.Ggs.Service.Registry.ContractDiscovery.Address != "" {
		return GlobalDefinition.Ggs.Service.Registry.ContractDiscovery.Address
	}
	return GlobalDefinition.Ggs.Service.Registry.Address
}

// GetContractDiscoveryTenant returns the Tenant of contract discovery registry
func GetContractDiscoveryTenant() string {
	if GlobalDefinition.Ggs.Service.Registry.ContractDiscovery.Tenant != "" {
		return GlobalDefinition.Ggs.Service.Registry.ContractDiscovery.Tenant
	}
	return GlobalDefinition.Ggs.Service.Registry.Tenant
}

// GetContractDiscoveryAPIVersion returns the APIVersion of contract discovery registry
func GetContractDiscoveryAPIVersion() string {
	if GlobalDefinition.Ggs.Service.Registry.ContractDiscovery.APIVersion.Version != "" {
		return GlobalDefinition.Ggs.Service.Registry.ContractDiscovery.APIVersion.Version
	}
	return GlobalDefinition.Ggs.Service.Registry.APIVersion.Version
}

// GetContractDiscoveryDisable returns the Disable of contract discovery registry
func GetContractDiscoveryDisable() bool {
	if b := archaius.GetBool("ggs.service.registry.contractDiscovery.disabled", false); b {
		return b
	}
	return archaius.GetBool("ggs.service.registry.disabled", false)
}
