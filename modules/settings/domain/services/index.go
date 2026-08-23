package services

import (
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
)

// InitDomainServices registers this module's domain services.
//
// The application services that delegate to them are registered separately by app.
// InitApplicationServices, which the module's Init calls immediately after this one.
func InitDomainServices() error {
	return deps.Register(
		NewSettingsDomainServiceImpl,
	)
}
