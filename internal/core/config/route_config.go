package config

//DefaultRouterType set the default router type
const DefaultRouterType = "ggs"

// GetRouterType returns the type of router
func GetRouterType() string {
	if OldRouterDefinition.Router.Infra != "" {
		return OldRouterDefinition.Router.Infra
	}
	return DefaultRouterType
}

// GetRouterEndpoints returns the router address
func GetRouterEndpoints() string {
	return OldRouterDefinition.Router.Address
}
