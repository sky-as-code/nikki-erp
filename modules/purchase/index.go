// Package purchase covers the procurement cycle from request for quotation to confirmed purchase
// order. An RFQ and a PO are the same purchase_order record at different statuses, so confirming
// changes the status and nothing about its identity. Vendor, Product, UoM, Warehouse, Stock,
// Receipt, Vendor Bill, Accounting, Tax and Payment are owned elsewhere and reached through ports.
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

// ModuleSingleton is the symbol the plugin loader looks up. It is typed DynamicModule, not
// InCodeModule, so that dropping RegisterModels fails the build instead of silently registering
// no schemas.
var ModuleSingleton modules.DynamicModule = &PurchaseModule{}

type PurchaseModule struct{}

func (*PurchaseModule) LabelKey() string {
	return "purchase.moduleLabel"
}

func (*PurchaseModule) Name() string {
	return modconstants.PurchaseModuleName
}

// Deps names every module Purchase reads through a port: dynamicresource hosts the resource
// engines, essential supplies UoM conversion and currency, inventory the product variant and its
// purchase_ok flag, contacts the vendor party and its vendor profile.
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

// Init implements DynamicModule. The order is load-bearing: external ports first because the
// derived line service needs them, then the engines because a derived service wraps the engine's
// own, then REST last because it registers the engines' routes.
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

// RegisterModels implements DynamicModule. The order is load-bearing: an edge is resolved against
// the schema registry at registration time, so the agreement must exist before the line pointing at
// it, and the order before its own line.
func (*PurchaseModule) RegisterModels() error {
	return stdErr.Join(
		dmodel.RegisterSchemaB(models.ConfigurationSchemaBuilder()),
		// Vendor prices reference nothing in Purchase (vendor and product are plain ulids into other
		// modules), so their position here is free.
		dmodel.RegisterSchemaB(models.VendorProductPriceSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SourcingGroupSchemaBuilder()),
		dmodel.RegisterSchemaB(models.AgreementSchemaBuilder()),
		dmodel.RegisterSchemaB(models.AgreementLineSchemaBuilder()),
		dmodel.RegisterSchemaB(models.PurchaseOrderSchemaBuilder()),
		dmodel.RegisterSchemaB(models.PurchaseOrderLineSchemaBuilder()),
		dmodel.RegisterSchemaB(models.AuditEventSchemaBuilder()),
	)
}
