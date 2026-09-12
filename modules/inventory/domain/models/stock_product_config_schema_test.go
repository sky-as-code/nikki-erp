package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
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

// The unit is a real foreign key into Essential's table.
//
// This reverses what this test asserted before: the reference used to be a plain id, on the
// grounds that a cross-module constraint would couple the two schemas. Referential integrity won
// that argument — an inventory_uom_id naming a unit that no longer exists is a balance counted in
// nothing, which no amount of module purity buys back. Isolation remains a rule about CODE: the
// constraint lives in the schema, while reads and writes still cross the boundary only through a
// port.
//
// ON DELETE SET NULL is the policy, so the column must accept NULL from the database while
// staying required of a client — which is what db_nullable expresses.
func TestStockProductConfigReferencesUomWithAForeignKey(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := StockProductConfigSchemaBuilder().Build()

	uomField, hasUomField := schema.Fields()[StockProductConfigFieldInventoryUomId]
	require.True(t, hasUomField, "the configuration must name the unit its balances are counted in")
	assert.True(t, uomField.IsRequiredForCreate(), "a client may not create a configuration without it")
	assert.True(t, uomField.IsNullable(), "ON DELETE SET NULL must be able to clear the column")

	var uomEdge *dmodel.ModelRelation
	for i, relation := range schema.ToRelations() {
		if relation.DestSchemaName == "essential_uom" {
			uomEdge = &schema.ToRelations()[i]
			break
		}
	}

	require.NotNil(t, uomEdge, "the unit reference must be a declared edge")
	assert.Equal(t, dmodel.RelationCascadeSetNull, uomEdge.OnDelete,
		"deleting a unit must clear the reference, never delete the configuration")
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
