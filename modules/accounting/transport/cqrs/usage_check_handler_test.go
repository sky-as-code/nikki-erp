package cqrs

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
)

// The probe covers exactly the rate version's rate_uom_id, and a schema it does not know answers
// an error, which the usage check counts as "cannot answer" and so refuses the delete.
func TestUomProbeCoversRateVersionsOnly(t *testing.T) {
	tables := uomTables()
	assert.Len(t, tables, 1)
	assert.Equal(t, models.TaxRateVersionSchemaName, tables[0].SchemaName)
	assert.Equal(t, models.TaxRateVersionFieldRateUomId, tables[0].Field)

	_, err := resolveRepository(models.TaxSchemaName)
	assert.Error(t, err)
}
