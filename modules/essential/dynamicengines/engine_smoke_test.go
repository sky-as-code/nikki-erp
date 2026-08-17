package dynamicengines

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
)

// EngineSchemaNames drives both engine creation and REST route registration, so a drift
// between it and the registered schemas silently unserves a resource.
func TestEngineSchemaNamesCoverEveryResource(t *testing.T) {
	assert.ElementsMatch(t,
		[]string{
			models.CurrencySchemaName,
			models.UomCatSchemaName,
			models.UomSchemaName,
		},
		EngineSchemaNames())
}

// The category must register before the UoM, since the UoM's edge points at it.
func TestSchemasRegisterInDependencyOrder(t *testing.T) {
	require.NoError(t, basemodel.RegisterJsonBaseSchemas())

	require.NoError(t, dmodel.RegisterSchemaB(models.UomCatSchemaBuilder()))
	require.NoError(t, dmodel.RegisterSchemaB(models.UomSchemaBuilder()))

	assert.NotNil(t, dmodel.GetSchema(models.UomSchemaName))
	assert.NotNil(t, dmodel.GetSchema(models.UomCatSchemaName))
}
