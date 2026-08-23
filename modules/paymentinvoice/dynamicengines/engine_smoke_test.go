package dynamicengines

import (
	"testing"

	"github.com/stretchr/testify/assert"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
)

// engineSchemaBuilders returns a builder per engine-backed schema, in no particular order --
// these assertions read one schema at a time.
func engineSchemaBuilders(t *testing.T) []*dmodel.ModelSchemaBuilder {
	t.Helper()
	_ = basemodel.RegisterJsonBaseSchemas()
	return []*dmodel.ModelSchemaBuilder{
		models.PaymentMethodSchemaBuilder(),
		models.OrderSchemaBuilder(),
		models.TransactionSchemaBuilder(),
		models.InvoiceSchemaBuilder(),
		models.InvoiceLineSchemaBuilder(),
	}
}

// EngineSchemaNames drives both engine creation and REST route registration, so a drift between
// it and the registered schemas silently unserves a resource.
func TestEngineSchemaNamesCoverEveryResource(t *testing.T) {
	assert.ElementsMatch(t,
		[]string{
			models.PaymentMethodSchemaName,
			models.OrderSchemaName,
			models.TransactionSchemaName,
			models.InvoiceSchemaName,
			models.InvoiceLineSchemaName,
		},
		EngineSchemaNames())
}

// A listing that returns no fields renders as a table of empty rows, and the failure looks like
// missing data rather than a missing declaration. The list now lives in the model JSON as
// `default_search_fields`, and an omitted one is legal to the schema builder, so nothing else
// catches a schema that was missed.
func TestEverySchemaDeclaresDefaultSearchFields(t *testing.T) {
	for _, builder := range engineSchemaBuilders(t) {
		schema := builder.Build()
		assert.NotEmptyf(t, schema.DefaultSearchFields(),
			"schema %q declares no default_search_fields", schema.Name())
	}
}
