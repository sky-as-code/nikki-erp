package cqrs

import (
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/services"
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

// resolveRepository answers which repository serves a schema, through the module's resource hub.
// It resolves at check time rather than at registration, so it does not matter whether the onions
// are built before the checker is registered; a schema not installed answers an error, which the
// dispatcher counts as "in use" and so refuses the delete.
func resolveRepository(schemaName string) (usagecheck.RowSearcher, error) {
	return services.RepositoryFor(schemaName)
}
