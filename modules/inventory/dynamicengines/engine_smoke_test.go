package dynamicengines

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// EngineSchemaNames drives both engine creation and REST route registration, so a drift between
// it and the registered schemas silently unserves a resource.
func TestEngineSchemaNamesCoverEveryProductResource(t *testing.T) {
	assert.ElementsMatch(t,
		[]string{
			models.ProductTypeSchemaName,
			models.ProductCategorySchemaName,
			models.BrandSchemaName,
			models.ProductAttributeSchemaName,
			models.ProductAttributeValueSchemaName,
			models.ProductTemplateSchemaName,
			models.ProductTemplateAttributeSchemaName,
			models.ProductTemplateAttributeValueSchemaName,
			models.ProductVariantSchemaName,
			models.ProductVariantAttributeValueSchemaName,
			models.StockLocationSchemaName,
			models.StockOperationTypeSchemaName,
			models.StockQuantSchemaName,
			models.StockTransferSchemaName,
			models.StockMoveSchemaName,
			models.StockMoveLineSchemaName,
			models.StockMoveDependencySchemaName,
			models.StockScrapSchemaName,
		},
		EngineSchemaNames())
}

// The two resources carrying business rules must declare DefineActions; a spec that loses it
// still serves CRUD, so the rules would go missing without any visible failure.
func TestRuleBearingResourcesDefineActions(t *testing.T) {
	withActions := map[string]bool{
		models.ProductCategorySchemaName: true,
		models.ProductTemplateSchemaName: true,
		models.ProductVariantSchemaName:  true,
		// Losing this one would silently reopen client writes to stock balances.
		models.StockQuantSchemaName: true,
		// The transfer's six movement operations are defined here; without them the resource
		// still serves CRUD and no stock can ever move.
		models.StockTransferSchemaName: true,
		// Losing this one would let a client write an allocation the balance knows nothing about.
		models.StockMoveLineSchemaName: true,
		// Do Scrap lives here; without it a scrap could be raised but never executed, and the
		// resource would still serve CRUD so nothing would look broken.
		models.StockScrapSchemaName: true,
	}

	for _, spec := range engineSpecs {
		if withActions[spec.SchemaName] {
			assert.NotNilf(t, spec.DefineActions,
				"schema %q carries business rules and must define actions", spec.SchemaName)
		}
	}
}

// A listing with no default fields renders as a table of bare ids.
func TestEverySpecDeclaresDefaultFields(t *testing.T) {
	for _, spec := range engineSpecs {
		assert.NotEmptyf(t, spec.DefaultFields,
			"schema %q must declare the fields its listing shows", spec.SchemaName)
	}
}

func TestDerefId(t *testing.T) {
	assert.Equal(t, "", derefId(nil))

	id := "01ABC"
	assert.Equal(t, "01ABC", derefId(&id))
}
