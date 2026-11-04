package external

import (
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	itExt "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/interfaces/external"
	itSeq "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/sequence"
	itSettings "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settings"
)

// registerNumberingPorts binds what Purchase's code numbering and its org settings need.
func registerNumberingPorts() error {
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
