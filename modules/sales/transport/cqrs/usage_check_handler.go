package cqrs

import (
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
)

// Sales' answer to "is this still in use", asked before another module deletes something Sales
// references.
//
// Essential refuses to delete or re-scale a unit of measure while sales history names it,
// because a recorded quantity without its unit is no longer a quantity. The foreign keys on
// uom_id are ON DELETE SET NULL, so the database alone would accept the delete and blank the
// unit from every line; this check is what keeps that from happening.
//
// Every table carrying the column is listed, regardless of the document's status. A draft
// quotation and a completed order both reference the unit, and a pricelist item is live
// configuration - none of them survives the unit being erased.

// variantTables are the Sales schemas holding a product_variant_id.
//
// Every one of them counts, whatever state its document is in. A confirmed order is history; a
// draft quotation is an offer someone may still accept; a pricelist item and a combo component
// are live configuration. None of them survives the product they name being deleted, and the
// foreign keys are ON DELETE RESTRICT precisely because none of them may be silently blanked.
func variantTables() []usagecheck.ReferencingTable {
	return []usagecheck.ReferencingTable{
		{SchemaName: models.SalesOrderLineSchemaName, Field: models.SalesOrderLineFieldProductVariantId},
		{SchemaName: models.SalesOrderLineComponentSchemaName, Field: models.SalesOrderLineComponentFieldProductVariantId},
		{SchemaName: models.SalesOrderFulfillmentItemSchemaName, Field: models.SalesOrderFulfillmentItemFieldProductVariantId},
		{SchemaName: models.SalesQuotationLineSchemaName, Field: models.SalesQuotationLineFieldVariantId},
		{SchemaName: models.SalesComboComponentSchemaName, Field: models.SalesComboComponentFieldProductVariantId},
		{SchemaName: models.SalesPricelistItemSchemaName, Field: models.SalesPricelistItemFieldProductVariantId},
	}
}

// uomTables are the Sales schemas holding a uom_id.
func uomTables() []usagecheck.ReferencingTable {
	return []usagecheck.ReferencingTable{
		{SchemaName: models.SalesOrderLineSchemaName, Field: models.SalesOrderLineFieldUomId},
		{SchemaName: models.SalesOrderLineComponentSchemaName, Field: models.SalesOrderLineComponentFieldUomId},
		{SchemaName: models.SalesOrderFulfillmentItemSchemaName, Field: models.SalesOrderFulfillmentItemFieldUomId},
		{SchemaName: models.SalesQuotationLineSchemaName, Field: models.SalesQuotationLineFieldUomId},
		{SchemaName: models.SalesComboComponentSchemaName, Field: models.SalesComboComponentFieldUomId},
		{SchemaName: models.SalesPricelistItemSchemaName, Field: models.SalesPricelistItemFieldUomId},
	}
}

// resolveRepository answers which repository serves a schema, at check time rather than at
// registration: checkers are registered during Init, while the onions are built afterwards.
func resolveRepository(schemaName string) (usagecheck.RowSearcher, error) {
	repo, err := services.RepositoryFor(schemaName)
	if err != nil {
		return nil, errors.Wrapf(err, "no repository for %s", schemaName)
	}
	return repo, nil
}
