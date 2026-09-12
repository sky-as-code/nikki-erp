package external

import (
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/notification/infra/external/channels"
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

		// Attaching a distribution channel is this one line plus its entry in
		// newNotificationChannels. Detaching it is removing them. Nothing in the core changes
		// either way, which is the property the channel interface exists to give (BR 5).
		channels.NewWebChannel,

		newNotificationChannels,
		NewChannelDispatcher,

		// The settings module's own service, narrowed to the registration this module needs. The
		// narrowing is the point: an alias of the full contract would re-export every method added
		// to it later, whether or not Notification should be able to call it.
		func(settingsSvc itSettings.TenantSettingsAppService) itExt.SettingsRegistrationExtService {
			return settingsSvc
		},

		// Reading an organization's own settings is a second, separately narrowed port: the
		// organization contract cannot reach a tenant's or a user's rows, which is the same
		// separation the settings module makes between its own three services.
		func(settingsSvc itSettings.OrgSettingsAppService) itExt.OrgSettingsReadExtService {
			return settingsSvc
		},

		NewOrgSettingsReader,
		func(reader *OrgSettingsReader) channels.OrgSettingsReader {
			return reader
		},
	)
}

// The compile guards. A channel that stops satisfying the interface should fail the build here,
// next to where it is attached, rather than at the point it is first used.
var (
	_ itExt.NotificationChannel = (*channels.WebChannel)(nil)
)

// newNotificationChannels collects the attached channels into the set the dispatcher works from.
//
// Each channel is a separate constructor argument rather than a variadic slice so that dig reports
// a missing one by name.
func newNotificationChannels(web *channels.WebChannel) NotificationChannels {
	return NotificationChannels{
		web.Name(): web,
	}
}
