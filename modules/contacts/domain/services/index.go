package services

import deps "github.com/sky-as-code/nikki-erp/common/deps_inject"

// InitDomainServices publishes the module's domain services into the dependency container.
//
// The vendor service is registered rather than only constructed, because Purchase resolves it
// through the container to back its own local port. Constructing it here and not registering it
// would leave that Invoke with nothing to find — which fails at boot, taking the whole application
// down rather than one module.
func InitDomainServices() error {
	return deps.Register(
		NewVendorDomainServiceImpl,
		NewVendorApplicationServiceImpl,
	)
}
