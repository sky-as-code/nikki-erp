package external

import (
	"context"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
	essconstants "github.com/sky-as-code/nikki-erp/modules/essential/constants"
	invconstants "github.com/sky-as-code/nikki-erp/modules/inventory/constants"
	"github.com/sky-as-code/nikki-erp/modules/purchase/constants"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

// Purchase's answer to "is this unit of measure still in use", asked before Essential deletes or
// re-scales one.
//
// Agreement lines count as much as order lines: an agreement commits to a quantity at a price in
// a stated unit, and the commitment is meaningless once the unit is gone. Vendor prices count
// too - a price per case stops being a price if "case" is erased - and were missed by the probe
// this replaces, so a unit reachable only from vendor pricing used to look unused.

// variantTables are the Purchase schemas holding a product_variant_id.
//
// A vendor price is configuration rather than history, but it counts all the same: a price per
// variant means nothing once the variant is gone, and leaving the row behind would make the
// catalogue quote a product nobody can buy.
func variantTables() []usagecheck.ReferencingTable {
	return []usagecheck.ReferencingTable{
		{SchemaName: models.PurchaseOrderLineSchemaName, Field: models.PurchaseOrderLineFieldProductVariantId},
		{SchemaName: models.AgreementLineSchemaName, Field: models.AgreementLineFieldProductVariantId},
		{SchemaName: models.VendorProductPriceSchemaName, Field: models.VendorProductPriceFieldProductVariantId},
	}
}

func uomTables() []usagecheck.ReferencingTable {
	return []usagecheck.ReferencingTable{
		{SchemaName: models.PurchaseOrderLineSchemaName, Field: models.PurchaseOrderLineFieldUomId},
		{SchemaName: models.AgreementLineSchemaName, Field: models.AgreementLineFieldUomId},
		{SchemaName: models.VendorProductPriceSchemaName, Field: models.VendorProductPriceFieldPurchaseUomId},
	}
}

// resolveRepository answers which repository serves a schema. Purchase is still on the legacy
// resource engine, so the lookup goes through its registry; resolving at check time rather than
// at registration keeps that detail from mattering to when the checker is registered.
func resolveRepository(schemaName string) (usagecheck.RowSearcher, error) {
	engine, ok := engineFor(schemaName)
	if !ok {
		return nil, errors.Errorf("no resource engine for %s", schemaName)
	}
	return engine.ResourceRepository(), nil
}

// RegisterUsageCheckers subscribes Purchase's answer to usage checks.
func RegisterUsageCheckers(bus cqrs.CqrsBus) error {
	registry := usagecheck.NewCheckerRegistry()

	if err := registry.Register(
		essconstants.EssentialModuleName,
		usagecheck.ResourceUom,
		usagecheck.NewTableChecker(resolveRepository, uomTables()...),
	); err != nil {
		return err
	}

	if err := registry.Register(
		invconstants.InventoryModuleName,
		usagecheck.ResourceProductVariant,
		usagecheck.NewTableChecker(resolveRepository, variantTables()...),
	); err != nil {
		return err
	}

	return usagecheck.SubscribeHandler(
		context.Background(), bus, constants.PurchaseModuleName, registry)
}
