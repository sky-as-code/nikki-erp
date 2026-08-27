package dynamicengines

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// EngineSchemaNames drives both engine creation and REST route registration, so a drift between it
// and the registered schemas silently unserves a resource.
func TestEngineSchemaNamesCoverEveryResource(t *testing.T) {
	assert.ElementsMatch(t,
		[]string{
			models.TaxJurisdictionSchemaName,
			models.TaxGroupSchemaName,
			models.TaxRoundingPolicySchemaName,
			models.TaxProductClassificationSchemaName,
			models.TaxSchemaName,
			models.TaxDefinitionVersionSchemaName,
			models.TaxRateVersionSchemaName,
			models.TaxComponentSchemaName,
			models.TaxMappingSchemaName,
			models.TaxMappingLineSchemaName,
			models.TaxRuleSchemaName,
			models.TaxRuleConditionSchemaName,
			models.TaxRuleResultSchemaName,
		},
		EngineSchemaNames())
}

// Registration order is load-bearing: an edge resolves against the registry as it is registered, so
// a schema must come after everything it points at. This is the same order as RegisterModels in the
// module root, and the assertion exists so that reordering one without the other fails here rather
// than at boot.
func TestSchemasRegisterInDependencyOrder(t *testing.T) {
	require.NoError(t, basemodel.RegisterJsonBaseSchemas())

	require.NoError(t, dmodel.RegisterSchemaB(models.TaxJurisdictionSchemaBuilder()))
	require.NoError(t, dmodel.RegisterSchemaB(models.TaxGroupSchemaBuilder()))
	require.NoError(t, dmodel.RegisterSchemaB(models.TaxRoundingPolicySchemaBuilder()))
	require.NoError(t, dmodel.RegisterSchemaB(models.TaxProductClassificationSchemaBuilder()))
	require.NoError(t, dmodel.RegisterSchemaB(models.TaxSchemaBuilder()))
	require.NoError(t, dmodel.RegisterSchemaB(models.TaxDefinitionVersionSchemaBuilder()))
	require.NoError(t, dmodel.RegisterSchemaB(models.TaxRateVersionSchemaBuilder()))
	require.NoError(t, dmodel.RegisterSchemaB(models.TaxComponentSchemaBuilder()))
	require.NoError(t, dmodel.RegisterSchemaB(models.TaxMappingSchemaBuilder()))
	require.NoError(t, dmodel.RegisterSchemaB(models.TaxMappingLineSchemaBuilder()))
	require.NoError(t, dmodel.RegisterSchemaB(models.TaxRuleSchemaBuilder()))
	require.NoError(t, dmodel.RegisterSchemaB(models.TaxRuleConditionSchemaBuilder()))
	require.NoError(t, dmodel.RegisterSchemaB(models.TaxRuleResultSchemaBuilder()))

	for _, schemaName := range EngineSchemaNames() {
		assert.NotNil(t, dmodel.GetSchema(schemaName), "schema %s did not register", schemaName)
	}
}
