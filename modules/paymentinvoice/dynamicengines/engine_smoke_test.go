package dynamicengines

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
)

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
// missing data rather than a missing declaration.
func TestEverySpecDeclaresDefaultFields(t *testing.T) {
	for _, spec := range engineSpecs {
		assert.NotEmpty(t, spec.DefaultFields, "engine '%s' declares no default search fields", spec.SchemaName)
	}
}
