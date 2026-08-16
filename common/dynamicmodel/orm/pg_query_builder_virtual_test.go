package orm

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/computed"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// virtualSchema builds a table with one physical column and one virtual scalar. The registry is
// a package singleton, so the schema is registered once and reused across these tests.
func virtualSchema(t *testing.T) (*dmodel.ModelSchema, *dmodel.SchemaRegistry) {
	t.Helper()
	registry := dmodel.GetSchemaRegistry()
	if existing := registry.Get(virtualTestSchemaName); existing != nil {
		return existing, registry
	}

	builder := dmodel.DefineModel(virtualTestSchemaName).
		TableName("test_virts").
		ShouldBuildDb().
		Field(dmodel.DefineField().Name("id").
			DataType(dmodel.FieldDataTypeUlid()).RequiredForCreate().PrimaryKey()).
		Field(dmodel.DefineField().Name("sku").
			DataType(dmodel.FieldDataTypeString(1, 100)).RequiredForCreate()).
		Field(dmodel.DefineField().Name("template_name").
			DataType(dmodel.FieldDataTypeString(0, 200)).
			Computed(false, computed.Related("template.name")))

	require.NoError(t, dmodel.RegisterSchemaB(builder))
	return registry.Get(virtualTestSchemaName), registry
}

const virtualTestSchemaName = "test_virt"

// A virtual field must never reach DDL: it has no SQL type, so emitting it would fail the
// migration outright.
func TestVirtual_CreateTableOmitsColumn(t *testing.T) {
	schema, registry := virtualSchema(t)

	sqls, cErrs, err := (&PgQueryBuilder{}).SqlCreateTable(schema, registry)

	require.NoError(t, err)
	require.Nil(t, cErrs)
	joined := strings.Join(sqls, "\n")
	assert.Contains(t, joined, `"sku"`)
	assert.NotContains(t, joined, "template_name")
}

// The write path drops the field rather than rejecting it, so a client that round-trips a read
// into a write is not punished for echoing back what it was given.
func TestVirtual_InsertDropsValueWithoutError(t *testing.T) {
	schema, _ := virtualSchema(t)

	sql, cErrs, err := (&PgQueryBuilder{}).SqlInsert(schema, dmodel.DynamicFields{
		"id":            "01J000000000000000000000",
		"sku":           "SKU-1",
		"template_name": "Classic T-Shirt",
	}, false)

	require.NoError(t, err)
	require.Nil(t, cErrs)
	require.NotNil(t, sql)
	assert.Contains(t, *sql, `"sku"`)
	assert.NotContains(t, *sql, "template_name")
}

// Selecting a virtual field is legal, but there is no column to project: it is dropped from the
// SELECT and filled by a service afterwards.
func TestVirtual_SelectOmitsColumnButSucceeds(t *testing.T) {
	schema, registry := virtualSchema(t)

	sql, cErrs, err := (&PgQueryBuilder{}).SqlSelectGraph(schema, registry, nil, SqlSelectGraphOpts{
		Columns: ToSelectColumns([]string{"id", "sku", "template_name"}),
	})

	require.NoError(t, err)
	require.Nil(t, cErrs)
	require.NotNil(t, sql)
	assert.Contains(t, *sql, `"sku"`)
	assert.NotContains(t, *sql, "template_name")
}

// Asking for only virtual fields would leave the projection empty. Falling through to "SELECT *"
// would contradict the caller, so the query is anchored on the primary key instead.
func TestVirtual_SelectOnlyVirtualFallsBackToPrimaryKey(t *testing.T) {
	schema, registry := virtualSchema(t)

	sql, cErrs, err := (&PgQueryBuilder{}).SqlSelectGraph(schema, registry, nil, SqlSelectGraphOpts{
		Columns: ToSelectColumns([]string{"template_name"}),
	})

	require.NoError(t, err)
	require.Nil(t, cErrs)
	require.NotNil(t, sql)
	assert.Contains(t, *sql, `"id"`)
	assert.NotContains(t, *sql, "template_name")
	assert.NotContains(t, *sql, "SELECT  FROM", "an empty projection would be invalid SQL")
}

// "You may not filter on this" and "no such field" are different facts with different fixes, so
// they must not collapse into one error key.
func TestVirtual_FilterReportsUnavailableNotUnknown(t *testing.T) {
	schema, registry := virtualSchema(t)

	graph := dmodel.NewSearchGraph()
	graph.NewCondition("template_name", dmodel.Equals, "Classic T-Shirt")

	_, cErrs, err := (&PgQueryBuilder{}).SqlSelectGraph(schema, registry, graph, SqlSelectGraphOpts{})

	require.NotNil(t, cErrs, "expected a client error, got err=%v", err)
	assert.Contains(t, string((*cErrs)[0].Key), "err_virtual_field_unavailable")
}

func TestVirtual_UnknownFieldStillReportsUnknown(t *testing.T) {
	schema, registry := virtualSchema(t)

	graph := dmodel.NewSearchGraph()
	graph.NewCondition("no_such_field", dmodel.Equals, "x")

	_, cErrs, err := (&PgQueryBuilder{}).SqlSelectGraph(schema, registry, graph, SqlSelectGraphOpts{})

	require.NotNil(t, cErrs, "expected a client error, got err=%v", err)
	assert.Contains(t, string((*cErrs)[0].Key), "err_unknown_schema_field")
}

// Ordering on a columnless field used to raise a 500. The caller can fix it by dropping the sort,
// so it belongs in ClientErrors.
func TestVirtual_OrderByReportsClientErrorNotPanic(t *testing.T) {
	schema, registry := virtualSchema(t)

	graph := dmodel.NewSearchGraph()
	graph.OrderBy("template_name", dmodel.Asc)

	_, cErrs, err := (&PgQueryBuilder{}).SqlSelectGraph(schema, registry, graph, SqlSelectGraphOpts{})

	require.NotNil(t, cErrs, "expected a client error, got err=%v", err)
	assert.Contains(t, string((*cErrs)[0].Key), "err_field_not_sortable")
}
