package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The stock settings of a product line, and the two things about its shape that carry meaning.

// The schema must parse and name itself exactly as the Go constant does. A mismatch is not caught
// until a request arrives, because the registry compares the two at runtime.
func TestStockProductConfigSchemaParses(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := StockProductConfigSchemaBuilder().Build()

	require.NotNil(t, schema)
	assert.Equal(t, StockProductConfigSchemaName, schema.Name())
}

// The unit is held as a plain id, not an edge.
//
// The UoM belongs to Essential. A foreign key across that boundary would make Inventory's schema
// depend on another module's table, which is the coupling the module's ports exist to prevent —
// and it would make the two modules undeployable apart.
func TestStockProductConfigReferencesUomWithoutAnEdge(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := StockProductConfigSchemaBuilder().Build()

	_, hasUomField := schema.Fields()[StockProductConfigFieldInventoryUomId]
	assert.True(t, hasUomField, "the configuration must name the unit its balances are counted in")

	// Asserted against the schema JSON, which is where edges are declared: the built ModelSchema
	// exposes no accessor for them, and the declaration is the artifact that would be wrong.
	assert.NotContains(t, stockProductConfigSchemaJson, "essential_uom",
		"a cross-module edge would couple Inventory's schema to Essential's table")
}

// There is no is_archived, deliberately.
//
// The configuration is not retired on its own: it lives and dies with the template it configures.
// An archivable configuration would also let a template's unit be hidden while its balances still
// referenced it, leaving historical quantities expressed in a unit the UI would not resolve.
func TestStockProductConfigIsNotArchivable(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := StockProductConfigSchemaBuilder().Build()

	_, archivable := schema.Fields()["is_archived"]
	assert.False(t, archivable,
		"the configuration follows its template rather than being archived independently")
}
