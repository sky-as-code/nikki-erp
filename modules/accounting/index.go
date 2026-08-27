// Package accounting owns tax: the master data, the effective-dated configuration, the
// determination rules and the calculation engine that turns them into an amount.
//
// It is a foundation service. Sales, and later Purchase and the point-of-sale surfaces, ask it what
// tax applies and what it comes to; none of them compute VAT themselves, and none of them are
// visible from here. The dependency runs one way on purpose (TAX-INV-01 to TAX-INV-03): Accounting
// holds no foreign key into a sales order, queries no downstream module to ask whether a
// configuration is in use, and would still calculate correctly in a deployment where Sales does
// not exist.
//
// What it deliberately does not do: post to a ledger, issue a VAT invoice, decide a discount, or
// convert a currency. Each of those either belongs to another module or has no agreed contract yet,
// and the requirement is explicit that shipping a half-defined version of one is worse than
// shipping none.
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

// ModuleSingleton is the exported symbol that will be looked up by the plugin loader.
//
// It is typed DynamicModule rather than InCodeModule so that dropping RegisterModels fails the
// build. Under the wider interface the method is found by a type assertion instead, and a module
// that has lost it still compiles, still loads, and silently registers no schemas at all.
var ModuleSingleton modules.DynamicModule = &AccountingModule{}

// OnAppStarted is found by a type assertion rather than by the interface above, so a rename or a
// changed signature would not fail the build — the module would simply load with its settings
// schema never registered, and nothing would say so. This assertion is what turns that into a
// compile error.
var _ modules.InCodeModuleAppStarted = &AccountingModule{}

type AccountingModule struct{}

func (*AccountingModule) LabelKey() string {
	return "accounting.moduleLabel"
}

func (*AccountingModule) Name() string {
	return modconstants.AccountingModuleName
}

// Deps names every module Accounting reads through a port.
//
// dynamicresource hosts the resource engines. settings stores the organization's tax policy.
// essential supplies the UoM conversion a fixed tax needs to turn a transaction quantity into the
// unit its rate is quoted in — Accounting must not implement that conversion itself
// (BR-TAX-ESS-SUP-014), which is the whole reason essential is named here.
//
// Sales is conspicuously absent and must stay so.
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

// Init implements DynamicModule.
//
// The order is load-bearing: the external ports bind first, because a derived service resolves its
// ports when it is constructed; the engines are created before the application services, because
// those wrap the engines; and the REST layer is registered last, because it registers the engines'
// routes and so cannot run before they exist.
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

// OnAppStarted implements InCodeModuleAppStarted.
//
// The settings schema is registered here rather than in Init() because peer module init order is
// nondeterministic: Init() cannot assume the settings module has built its engines yet, while
// OnAppStarted runs after every module has initialized. Registration is idempotent, so it runs
// unconditionally.
func (*AccountingModule) OnAppStarted() error {
	return deps.Invoke(func(settingsSvc itExt.SettingsRegistrationExtService) error {
		return registerOrgSettings(corectx.NewRequestContext(context.Background()), settingsSvc)
	})
}

// RegisterModels implements DynamicModule.
//
// Registration order is load-bearing: an edge is resolved against the schema registry at
// registration time, so a referenced schema must be registered before the one pointing at it — the
// jurisdiction before the taxes sited in it, the tax before its versions, the rule before its
// conditions and results.
//
// The schemas are listed here rather than scattered across the packages that own them, so that the
// order is visible in a single place and a missing registration is a gap in a list rather than an
// absence nobody can see.
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
