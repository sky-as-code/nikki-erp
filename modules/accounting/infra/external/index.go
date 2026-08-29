// Package external binds Accounting's local ports to the services other modules publish.
//
// This is the ONLY package in Accounting that may import another module. Everything else depends
// on the interfaces in interfaces/external, so splitting a module into its own process changes
// this file and nothing else — the bindings become REST or CQRS clients and every caller is
// unaffected.
package external

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	itExt "github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/external"
	itCurrency "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/currency"
	itUom "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uom"
	itSettings "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settings"
)

// InitExternal binds every port Accounting consumes.
func InitExternal() error {
	return stdErr.Join(
		deps.Register(func(uomSvc itUom.UomConversionAppService) itExt.UomExtService {
			// The upstream service already has both methods the port declares, so this is a
			// direct hand-over rather than an adapter. It becomes a client when this application
			// is split into separate microservices.
			return uomSvc
		}),
		deps.Register(func(settingsSvc itSettings.TenantSettingsAppService) itExt.SettingsRegistrationExtService {
			return settingsSvc
		}),
		deps.Register(func(settingsSvc itSettings.EffectiveSettingsAppService) itExt.EffectiveSettingsExtService {
			return settingsSvc
		}),
		deps.Register(func(currencySvc itCurrency.CurrencyAppService) itExt.CurrencyExtService {
			// A direct hand-over: the upstream service already has GetCurrency. The port stays
			// narrower than the service it binds, so Accounting cannot start rounding through it.
			return currencySvc
		}),
	)
}
