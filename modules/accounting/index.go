// Package accounting owns tax: master data, effective-dated configuration, determination rules and
// the calculation engine.
//
// The dependency runs one way: Accounting holds no foreign key into a downstream module and must
// stay calculable in a deployment without Sales. It does not post to a ledger, issue a VAT invoice,
// decide a discount, or convert a currency.
package accounting

import (
	"context"
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/accounting/app"
	modconstants "github.com/sky-as-code/nikki-erp/modules/accounting/constants"
	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/accounting/dynamicengines"
	"github.com/sky-as-code/nikki-erp/modules/accounting/infra/external"
	itExt "github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/external"
	"github.com/sky-as-code/nikki-erp/modules/accounting/transport/restful"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// ModuleSingleton is the symbol the plugin loader looks up. It is typed DynamicModule rather than
// InCodeModule so that dropping RegisterModels fails the build instead of silently registering no
// schemas.
var ModuleSingleton modules.DynamicModule = &AccountingModule{}

// OnAppStarted is found by a type assertion, so this assertion is what turns a rename or signature
// change into a compile error instead of a silently unregistered settings schema.
var _ modules.InCodeModuleAppStarted = &AccountingModule{}

type AccountingModule struct{}

func (*AccountingModule) LabelKey() string {
	return "accounting.moduleLabel"
}

func (*AccountingModule) Name() string {
	return modconstants.AccountingModuleName
}

// Deps names every module Accounting reads through a port. essential supplies the UoM conversion a
// fixed tax needs to reach the unit its rate is quoted in; Accounting must not implement that
// itself. Sales is absent and must stay so.
func (*AccountingModule) Deps() []string {
	return []string{
		"dynamicresource",
		"essential",
		"settings",
	}
}

func (*AccountingModule) IsInternal() bool {
	return false
}

func (*AccountingModule) Version() semver.SemVer {
	return *semver.MustParseSemVer("v1.0.0")
}

// Init implements DynamicModule. The order is load-bearing: external ports bind first (a derived
// service resolves its ports at construction), engines before the application services that wrap
// them, and REST last because it registers the engines' routes.
func (*AccountingModule) Init() error {
	if err := external.InitExternal(); err != nil {
		return err
	}
	if err := dynamicengines.InitDynamicEngines(); err != nil {
		return err
	}
	if err := app.InitApplicationServices(); err != nil {
		return err
	}
	return restful.InitRestfulHandlers()
}

// OnAppStarted implements InCodeModuleAppStarted. The settings schema is registered here, not in
// Init(), because peer module init order is nondeterministic and Init() cannot assume the settings
// module has built its engines. Registration is idempotent.
func (*AccountingModule) OnAppStarted() error {
	return deps.Invoke(func(settingsSvc itExt.SettingsRegistrationExtService) error {
		return registerOrgSettings(corectx.NewRequestContext(context.Background()), settingsSvc)
	})
}

// RegisterModels implements DynamicModule. Registration order is load-bearing: edges resolve
// against the schema registry at registration time, so a referenced schema must be registered
// before the one pointing at it (jurisdiction before taxes, tax before its versions, rule before
// its conditions and results).
func (*AccountingModule) RegisterModels() error {
	return stdErr.Join(
		dmodel.RegisterSchemaB(models.TaxJurisdictionSchemaBuilder()),
		dmodel.RegisterSchemaB(models.TaxGroupSchemaBuilder()),
		dmodel.RegisterSchemaB(models.TaxRoundingPolicySchemaBuilder()),
		dmodel.RegisterSchemaB(models.TaxProductClassificationSchemaBuilder()),
		dmodel.RegisterSchemaB(models.TaxSchemaBuilder()),
		dmodel.RegisterSchemaB(models.TaxDefinitionVersionSchemaBuilder()),
		dmodel.RegisterSchemaB(models.TaxRateVersionSchemaBuilder()),
		dmodel.RegisterSchemaB(models.TaxComponentSchemaBuilder()),
		dmodel.RegisterSchemaB(models.TaxMappingSchemaBuilder()),
		dmodel.RegisterSchemaB(models.TaxMappingLineSchemaBuilder()),
		dmodel.RegisterSchemaB(models.TaxRuleSchemaBuilder()),
		dmodel.RegisterSchemaB(models.TaxRuleConditionSchemaBuilder()),
		dmodel.RegisterSchemaB(models.TaxRuleResultSchemaBuilder()),
	)
}
