package external

import (
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	itExt "github.com/sky-as-code/nikki-erp/modules/notification/interfaces/external"
	itSettings "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settings"
)

// InitExternalServices binds Notification to the infrastructure it depends on.
//
// Both bindings are of an interface this module declares onto an implementation built from the
// shared services, which is what keeps the rest of the module unaware of whether the broker is
// Redis, MQTT or anything else.
func InitExternalServices() error {
	return deps.Register(
		NewPubSubBroker,
		NewMemoryStreamRegistry,

		// The settings module's own service, narrowed to the registration this module needs. The
		// narrowing is the point: an alias of the full contract would re-export every method added
		// to it later, whether or not Notification should be able to call it.
		func(settingsSvc itSettings.TenantSettingsAppService) itExt.SettingsRegistrationExtService {
			return settingsSvc
		},
	)
}
