// Package purchase covers the procurement cycle from request for quotation through to a confirmed
// purchase order, per docs/requirements/purchase/business-requirements.md.
//
// One resource carries both halves of that cycle: an RFQ and a PO are the same purchase_order at
// different points of its status, so confirming one changes its status and nothing about its
// identity. It deliberately owns none of Vendor, Product, UoM, Warehouse, Stock, Receipt, Vendor
// Bill, Accounting, Tax or Payment — it holds ids into those modules and reads them through ports.
package purchase

import (
	stdErr "errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/modules"
	modconstants "github.com/sky-as-code/nikki-erp/modules/purchase/constants"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/purchase/dynamicengines"
	"github.com/sky-as-code/nikki-erp/modules/purchase/infra/external"
	"github.com/sky-as-code/nikki-erp/modules/purchase/transport/restful"
)

// ModuleSingleton is the exported symbol that will be looked up by the plugin loader.
//
// It is typed DynamicModule rather than InCodeModule so that dropping RegisterModels fails the
// build. Under the wider interface the method is found by a type assertion instead, and a module
// that has lost it still compiles, still loads, and silently registers no schemas at all.
var ModuleSingleton modules.DynamicModule = &PurchaseModule{}

type PurchaseModule struct{}

func (*PurchaseModule) LabelKey() string {
	return "purchase.moduleLabel"
}

func (*PurchaseModule) Name() string {
	return modconstants.PurchaseModuleName
}

// Deps names every module Purchase reads through a port.
//
// dynamicresource hosts the resource engines. essential supplies UoM conversion and currency,
// inventory the product variant and its purchase_ok flag, contacts the vendor party and its vendor
// profile. The last three were already declared before this module was rebuilt, but nothing
// realized them — there was no interfaces/external and no infra/external at all.
func (*PurchaseModule) Deps() []string {
	return []string{
		"dynamicresource",
		"essential",
		"inventory",
		"contacts",
	}
}

func (*PurchaseModule) IsInternal() bool {
	return false
}

func (*PurchaseModule) Version() semver.SemVer {
	return *semver.MustParseSemVer("v1.0.0")
}

// Init implements DynamicModule.
//
// The order is load-bearing three times over. The external ports are bound first, because the
// derived line service needs them; the engines are created before the derived services, because a
// derived service wraps the engine's own; and the REST layer is registered last, because it
// registers the engines' routes and so cannot run before they exist.
func (*PurchaseModule) Init() error {
	if err := external.InitExternal(); err != nil {
		return err
	}
	if err := dynamicengines.InitDynamicEngines(); err != nil {
		return err
	}
	if err := dynamicengines.InitDomainServices(); err != nil {
		return err
	}
	return restful.InitRestfulHandlers()
}

// RegisterModels implements DynamicModule.
//
// The order is load-bearing: an edge is resolved against the schema registry at registration time,
// so the agreement must exist before the line that points at it, and the order before its own line.
//
// The four schemas this module used to register — purchase_request, request_for_quote,
// request_for_proposal and vendor — were deleted rather than carried forward: the requirement
// collapses the first three into one purchase_order carrying status, and says outright that
// Purchase must not own the vendor.
func (*PurchaseModule) RegisterModels() error {
	return stdErr.Join(
		dmodel.RegisterSchemaB(models.ConfigurationSchemaBuilder()),
		// Vendor prices reference nothing in Purchase — the vendor is Contacts', the product is
		// Inventory's, both as plain ulids — so they may be registered as early as anything else.
		dmodel.RegisterSchemaB(models.VendorProductPriceSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SourcingGroupSchemaBuilder()),
		dmodel.RegisterSchemaB(models.AgreementSchemaBuilder()),
		dmodel.RegisterSchemaB(models.AgreementLineSchemaBuilder()),
		dmodel.RegisterSchemaB(models.PurchaseOrderSchemaBuilder()),
		dmodel.RegisterSchemaB(models.PurchaseOrderLineSchemaBuilder()),
		dmodel.RegisterSchemaB(models.AuditEventSchemaBuilder()),
	)
}
