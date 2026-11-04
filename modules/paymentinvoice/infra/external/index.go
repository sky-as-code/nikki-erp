// Package external binds Payment & Invoice's local ports to the services other modules publish. It
// is the only package in the module that may import another feature module; everything else
// depends on interfaces/external, so a split into separate processes changes this package alone.
package external

import (
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	itSeq "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/sequence"
	itExt "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/external"
	itSettings "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settings"
)

// InitExternal binds what the invoice numbering and the module's org settings need. It runs before
// the domain services are built, because the invoice service resolves its ports at construction.
func InitExternal() error {
	return deps.Register(
		func(sequences itSeq.SequenceService) itExt.SequenceExtService {
			// Direct hand-over: Essential's port asserts nothing and settles the idempotency id.
			return sequences
		},
		func(settings itSettings.OrgSettingsAppService) itExt.OrgSettingsSeedExtService {
			return settings
		},
		func(settings itSettings.TenantSettingsAppService) itExt.SettingsRegistrationExtService {
			// The registry hangs off the TENANT service although the schema is org-level: declaring
			// what may be configured is tenant-wide whatever level the values live at.
			return settings
		},
	)
}
