package orm

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

func placementSchema(name, table, refSchema string, loose bool) *dmodel.ModelSchema {
	edge := dmodel.Edge("ref").
		ManyToOne(refSchema, dmodel.DynamicFields{"ref_id": "id"}).
		OnDelete(dmodel.RelationCascadeSetNull)
	if loose {
		edge.NoDbConstraint()
	}
	return dmodel.DefineModel(name).
		TableName(table).
		ShouldBuildDb().
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("ref_id").DataType(dmodel.FieldDataTypeUlid())).
		EdgeTo(edge).
		Build()
}

func TestLooseEdge_NoForeignKeyGenerated(t *testing.T) {
	child := placementSchema("place_loose", "place_looses", "refact_parent", true)
	registry := refActionRegistry(t, refActionParent(), child)

	sqls, _, err := NewPgQueryBuilder().SqlCreateTable(child, registry)
	require.NoError(t, err)
	assert.NotContains(t, sqls[0], "FOREIGN KEY")

	require.Len(t, child.ToRelations(), 1, "a loose edge stays queryable")
	assert.True(t, child.ToRelations()[0].NoDbConstraint)
}

func TestLooseEdge_PrunedWhenDestinationNotRegistered(t *testing.T) {
	child := placementSchema("place_orphan", "place_orphans", "not_registered", true)
	registry := refActionRegistry(t, child)

	assert.Empty(t, child.ToRelations())
	_, err := GenCreateSql(registry, DialectPostgres)
	require.NoError(t, err)
}

func TestStrictEdge_FailsWhenDestinationNotRegistered(t *testing.T) {
	child := placementSchema("place_strict", "place_stricts", "not_registered", false)
	registry := dmodel.NewSchemaRegistry()
	require.NoError(t, registry.Register(child))
	assert.Error(t, registry.FinalizeRelations())
}

// A cycle has no creation order, so the constraint pointing forward is moved after every
// CREATE TABLE instead of naming a table that does not exist yet.
func TestGenCreateSql_CyclicForeignKeyIsDeferred(t *testing.T) {
	first := placementSchema("place_a", "place_as", "place_b", false)
	second := placementSchema("place_b", "place_bs", "place_a", false)
	registry := refActionRegistry(t, first, second)

	sqls, err := GenCreateSql(registry, DialectPostgres)
	require.NoError(t, err)

	creates, alters := 0, 0
	for i, sql := range sqls {
		if strings.HasPrefix(sql, "CREATE TABLE") {
			creates++
			assert.Zero(t, alters, "every ALTER comes after the CREATEs, statement %d", i)
		}
		if strings.HasPrefix(sql, "ALTER TABLE") {
			alters++
			assert.Contains(t, sql, "ADD CONSTRAINT")
			assert.Contains(t, sql, "ON DELETE SET NULL")
		}
	}
	assert.Equal(t, 2, creates)
	assert.Equal(t, 1, alters)
}
