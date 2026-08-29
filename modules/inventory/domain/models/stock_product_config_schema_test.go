package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The stock settings of a product line, and the two things about its shape that carry meaning.

// The schema must name itself exactly as the Go constant does; the registry compares the two at
// runtime, so a mismatch is not caught until a request arrives.
func TestStockProductConfigSchemaParses(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := StockProductConfigSchemaBuilder().Build()

	require.NotNil(t, schema)
	assert.Equal(t, StockProductConfigSchemaName, schema.Name())
}

// The unit is held as a plain id, not an edge: the UoM belongs to Essential, and a foreign key
// across that boundary would couple Inventory's schema to another module's table and make the two
// undeployable apart.
func TestStockProductConfigReferencesUomWithoutAnEdge(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := StockProductConfigSchemaBuilder().Build()

	_, hasUomField := schema.Fields()[StockProductConfigFieldInventoryUomId]
	assert.True(t, hasUomField, "the configuration must name the unit its balances are counted in")

	// Asserted against the schema JSON, where edges are declared: the built ModelSchema exposes no
	// accessor for them.
	assert.NotContains(t, stockProductConfigSchemaJson, "essential_uom",
		"a cross-module edge would couple Inventory's schema to Essential's table")
}

// There is deliberately no is_archived: the configuration lives and dies with its template, and an
// archivable one would let a template's unit be hidden while balances still referenced it, leaving
// historical quantities in a unit the UI cannot resolve.
func TestStockProductConfigIsNotArchivable(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := StockProductConfigSchemaBuilder().Build()

	_, archivable := schema.Fields()["is_archived"]
	assert.False(t, archivable,
		"the configuration follows its template rather than being archived independently")
}
