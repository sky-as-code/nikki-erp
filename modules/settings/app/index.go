package app

import (
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
)

// InitApplicationServices registers this module's application services.
//
// They must be registered alongside the domain services they delegate to: registering only a
// domain service leaves its application counterpart unprovided, which is not a compile error and
// surfaces only as a boot failure in the module that tried to bind it.
func InitApplicationServices() error {
	return deps.Register(
		NewTenantSettingsAppServiceImpl,
		NewOrgSettingsAppServiceImpl,
		NewUserPreferencesAppServiceImpl,
	)
}
