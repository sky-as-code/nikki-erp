// Package settings gives every feature module one place to declare its configuration and have it
// stored, resolved and rendered, instead of each building its own settings table.
//
// A module registers a setting schema per level (tenant, org, user); settings owns persistence,
// the override rule and the read path. Two things shape the design:
//
//   - allow_override is stored per record, so a tenant admin decides it for their own tenant. A
//     module still declares a starting value as field metadata, which applies until someone rules
//     otherwise; only a tenant owner may write the column.
//   - Enforcement is a physical write rather than a read-time resolution. Saving a tenant setting
//     whose schema says allow_override=false fans the value out onto every child record, so reads
//     are plain reads with no precedence resolver behind them.
//
// Settings must never import iam. iam already declares settings in its Deps(), so the reverse edge
// would be a startup-aborting cycle: organization and user creation call *into* settings through a
// port iam owns, and the fan-out matches children by tenant_id rather than by joining iam tables.
package settings

import (
	stdErr "errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/settings/app"
	modconstants "github.com/sky-as-code/nikki-erp/modules/settings/constants"
	"github.com/sky-as-code/nikki-erp/modules/settings/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/settings/domain/services"
	"github.com/sky-as-code/nikki-erp/modules/settings/dynamicengines"
	"github.com/sky-as-code/nikki-erp/modules/settings/transport/restful"
)

// ModuleSingleton is the exported symbol that will be looked up by the plugin loader.
//
// It is typed DynamicModule rather than InCodeModule so that dropping RegisterModels fails the
// build. Under the wider interface the method is found by a type assertion instead, and a module
// that has lost it still compiles, still loads, and silently registers no schemas at all.
var ModuleSingleton modules.DynamicModule = &SettingsModule{}

type SettingsModule struct {
}

// LabelKey implements NikkiModule.
func (*SettingsModule) LabelKey() string {
	return "settings.moduleLabel"
}

// Name implements NikkiModule.
func (*SettingsModule) Name() string {
	return modconstants.SettingsModuleName
}

// Deps implements NikkiModule.
//
// core and essential are implicit and must not be listed. iam is deliberately absent and must stay
// that way: iam depends on settings, so naming it here would close a cycle and abort startup.
func (*SettingsModule) Deps() []string {
	return []string{
		"dynamicresource",
	}
}

// IsInternal implements InCodeModule.
func (*SettingsModule) IsInternal() bool {
	return false
}

// Version implements NikkiModule.
func (*SettingsModule) Version() semver.SemVer {
	return *semver.MustParseSemVer("v1.0.0")
}

// Init implements NikkiModule.
//
// The order is load-bearing: the engines must exist before the services that read and write
// through them, the application services bind the domain services they delegate to, and transport
// registers routes onto the application services last.
func (*SettingsModule) Init() error {
	if err := dynamicengines.InitDynamicEngines(); err != nil {
		return err
	}
	if err := services.InitDomainServices(); err != nil {
		return err
	}
	if err := app.InitApplicationServices(); err != nil {
		return err
	}
	return restful.InitRestfulHandlers()
}

// RegisterModels implements DynamicModule.
//
// settings_schema is registered first because settings_record points at it, and an edge is
// resolved against the schema registry at registration time.
func (*SettingsModule) RegisterModels() error {
	return stdErr.Join(
		dmodel.RegisterSchemaB(models.SettingsSchemaSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SettingsRecordSchemaBuilder()),
	)
}
