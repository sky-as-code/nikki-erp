// Package purchase covers the procurement cycle from request for quotation to confirmed purchase
// order. An RFQ and a PO are the same purchase_order record at different statuses, so confirming
// changes the status and nothing about its identity. Vendor, Product, UoM, Warehouse, Stock,
// Receipt, Vendor Bill, Accounting, Tax and Payment are owned elsewhere and reached through ports.
package purchase

import (
	stdErr "errors"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/commerce/purchase/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/commerce/purchase/domain/services"
	"github.com/sky-as-code/nikki-erp/modules/commerce/purchase/dynamicengines"
	"github.com/sky-as-code/nikki-erp/modules/commerce/purchase/transport/restful"
)

// InitSubModule builds the Purchase part of Commerce. The order is load-bearing: the external ports
// the onions inject are bound beforehand by commerce/infra.InitSubModule, then the onion
// registrations, then the routes.
// BuildAllEngines forces every onion no route resolved, so the resource hub the lifecycle
// operations and the usage checker read through is complete before the first request.
func InitSubModule() error {
	if err := dynamicengines.InitDynamicEngines(); err != nil {
		return err
	}
	if err := restful.InitRestfulHandlers(); err != nil {
		return err
	}
	if err := dynamicengines.BuildAllEngines(); err != nil {
		return err
	}
	return assertEveryResourceInstalled()
}

// assertEveryResourceInstalled fails the boot when an onion was declared but never built: the first
// caller to notice would otherwise be a lifecycle action or a usage check.
func assertEveryResourceInstalled() error {
	installed := map[string]bool{}
	for _, name := range services.InstalledResourceNames() {
		installed[name] = true
	}
	var missing []error
	for _, name := range dynamicengines.SchemaNames() {
		if !installed[name] {
			missing = append(missing,
				errors.Errorf("the '%s' purchase resource was declared but never built", name))
		}
	}
	return stdErr.Join(missing...)
}

// RegisterModels registers the Purchase schemas. The order is load-bearing: an edge is resolved against
// the schema registry at registration time, so the agreement must exist before the line pointing at
// it, and the order before its own line.
func RegisterModels() error {
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
