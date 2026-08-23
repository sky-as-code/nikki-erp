// Package external binds Essential's local ports to the services other modules publish.
//
// This is the ONLY package in Essential that may import another module. Everything else depends on
// the interfaces in interfaces/external, so splitting a module into its own process changes this
// file and nothing else.
package external

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	itExt "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/external"
	itSettings "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settings"
)

func InitExternalServices() error {
	return stdErr.Join(
		deps.Register(func(settingsSvc itSettings.TenantSettingsAppService) itExt.SettingsRegistrationExtService {
			return settingsSvc
		}),
		deps.Register(func(settingsSvc itSettings.EffectiveSettingsAppService) itExt.EffectiveSettingsExtService {
			return settingsSvc
		}),
	)
}
