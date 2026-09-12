package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// A cross-module foreign key states its delete policy explicitly, and RESTRICT is the one that
// refuses the parent delete outright. These tests pin the emitted DDL, because the keyword
// reaches PostgreSQL verbatim: a value the builder accepts but the database does not recognize
// fails at migration time rather than at build time.

func refActionRegistry(t *testing.T, schemas ...*dmodel.ModelSchema) *dmodel.SchemaRegistry {
	t.Helper()
	registry := dmodel.NewSchemaRegistry()
	for _, schema := range schemas {
		require.NoError(t, registry.Register(schema))
	}
	require.NoError(t, registry.FinalizeRelations())
	return registry
}

func refActionParent() *dmodel.ModelSchema {
	return dmodel.DefineModel("refact_parent").
		TableName("refact_parents").
		ShouldBuildDb().
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Build()
}

func refActionChild(onDelete dmodel.RelationCascade) *dmodel.ModelSchema {
	return dmodel.DefineModel("refact_child").
		TableName("refact_children").
		ShouldBuildDb().
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("parent_id").DataType(dmodel.FieldDataTypeUlid())).
		EdgeTo(dmodel.Edge("parent").
			ManyToOne("refact_parent", dmodel.DynamicFields{"parent_id": "id"}).
			OnDelete(onDelete)).
		Build()
}

func refActionChildDdl(t *testing.T, onDelete dmodel.RelationCascade) string {
	t.Helper()
	child := refActionChild(onDelete)
	registry := refActionRegistry(t, refActionParent(), child)
	sqls, clientErrs, err := NewPgQueryBuilder().SqlCreateTable(child, registry)
	require.NoError(t, err)
	require.True(t, clientErrs == nil || clientErrs.Count() == 0)
	require.NotEmpty(t, sqls)
	return sqls[0]
}

func TestReferentialAction_RestrictReachesTheDdl(t *testing.T) {
	assert.Contains(t, refActionChildDdl(t, dmodel.RelationCascadeRestrict),
		`FOREIGN KEY ("parent_id") REFERENCES "refact_parents" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT`)
}

func TestReferentialAction_SetNullReachesTheDdl(t *testing.T) {
	assert.Contains(t, refActionChildDdl(t, dmodel.RelationCascadeSetNull),
		`ON DELETE SET NULL`)
}

// The zero value must keep meaning NO ACTION: most existing relations declare no policy at all,
// and a change here would silently rewrite every one of them.
func TestReferentialAction_UnsetStaysNoAction(t *testing.T) {
	assert.Contains(t, refActionChildDdl(t, ""), `ON DELETE NO ACTION`)
}

func TestReferentialAction_RejectsUnknownKeyword(t *testing.T) {
	assert.False(t, dmodel.RelationCascade("RESTRCIT").IsValid(), "a typo is not a valid action")
	assert.True(t, dmodel.RelationCascade("").IsValid(), "unset is valid and means NO ACTION")

	assert.Panics(t, func() {
		dmodel.Edge("parent").
			ManyToOne("refact_parent", dmodel.DynamicFields{"parent_id": "id"}).
			OnDelete(dmodel.RelationCascade("RESTRCIT"))
	}, "an unknown referential action must fail while the schema is built")
}

// A column whose foreign key is ON DELETE SET NULL must accept NULL from the database even
// though a client may not omit it: the constraint clears the column when the parent goes away,
// and a NOT NULL column would refuse that. db_nullable separates the two statements, which
// required_for_create otherwise conflates.
func TestReferentialAction_DbNullableDropsNotNullButKeepsRequired(t *testing.T) {
	child := dmodel.DefineModel("refact_child").
		TableName("refact_children").
		ShouldBuildDb().
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("parent_id").DataType(dmodel.FieldDataTypeUlid()).
			RequiredForCreate().DbNullable()).
		EdgeTo(dmodel.Edge("parent").
			ManyToOne("refact_parent", dmodel.DynamicFields{"parent_id": "id"}).
			OnDelete(dmodel.RelationCascadeSetNull)).
		Build()
	registry := refActionRegistry(t, refActionParent(), child)

	sqls, _, err := NewPgQueryBuilder().SqlCreateTable(child, registry)
	require.NoError(t, err)

	assert.Contains(t, sqls[0], `"parent_id" character varying NULL`,
		"the column must accept the NULL the constraint writes")
	assert.Contains(t, sqls[0], `ON DELETE SET NULL`)

	field, ok := child.Field("parent_id")
	require.True(t, ok)
	assert.True(t, field.IsRequiredForCreate(), "a client still may not omit it")
	assert.True(t, field.IsNullable())
}

// Without the override, required_for_create still means NOT NULL.
func TestReferentialAction_RequiredStaysNotNullByDefault(t *testing.T) {
	child := dmodel.DefineModel("refact_child2").
		TableName("refact_children2").
		ShouldBuildDb().
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("parent_id").DataType(dmodel.FieldDataTypeUlid()).
			RequiredForCreate()).
		Build()

	sqls, _, err := NewPgQueryBuilder().SqlCreateTable(child, dmodel.NewSchemaRegistry())
	require.NoError(t, err)

	assert.Contains(t, sqls[0], `"parent_id" character varying NOT NULL`)
}
