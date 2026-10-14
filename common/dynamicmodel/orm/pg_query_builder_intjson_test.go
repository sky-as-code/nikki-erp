package orm

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// A search graph arrives as JSON, which decodes every number as float64. An integer column must
// still accept a whole number, and still refuse a fractional one.

func intJsonSchema(t *testing.T) (*dmodel.ModelSchema, *dmodel.SchemaRegistry) {
	t.Helper()
	reg := dmodel.NewSchemaRegistry()
	schema := dmodel.DefineModel("test_orm_intjson").ShouldBuildDb().TableName("test_orm_intjsons").
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("qty").DataType(dmodel.FieldDataTypeInt64(0, 1000))).
		Build()
	require.NoError(t, reg.Register(schema))
	require.NoError(t, reg.FinalizeRelations())
	return schema, reg
}

func graphFromJson(t *testing.T, raw string) *dmodel.SearchGraph {
	t.Helper()
	var graph dmodel.SearchGraph
	require.NoError(t, json.Unmarshal([]byte(raw), &graph))
	return &graph
}

func TestIntJson_WholeNumberFiltersAnIntegerColumn(t *testing.T) {
	schema, reg := intJsonSchema(t)
	graph := graphFromJson(t, `{"if":["qty", ">=", 3]}`)
	sql, cErrs, err := NewPgQueryBuilder().SqlSelectGraph(schema, reg, graph, SqlSelectGraphOpts{})
	require.NoError(t, err)
	require.Nil(t, cErrs)
	assert.Contains(t, *sql, `"qty" >= 3`)
}

func TestIntJson_InListOfWholeNumbers(t *testing.T) {
	schema, reg := intJsonSchema(t)
	graph := graphFromJson(t, `{"if":["qty", "in", 1, 2]}`)
	sql, cErrs, err := NewPgQueryBuilder().SqlSelectGraph(schema, reg, graph, SqlSelectGraphOpts{})
	require.NoError(t, err)
	require.Nil(t, cErrs)
	assert.Contains(t, *sql, `"qty" IN (1, 2)`)
}

func TestIntJson_FractionalNumberIsRefused(t *testing.T) {
	schema, reg := intJsonSchema(t)
	graph := graphFromJson(t, `{"if":["qty", "=", 2.5]}`)
	sql, cErrs, err := NewPgQueryBuilder().SqlSelectGraph(schema, reg, graph, SqlSelectGraphOpts{})
	require.NoError(t, err)
	assert.Nil(t, sql)
	require.NotNil(t, cErrs)
	assert.Contains(t, (*cErrs)[0].Key, "err_invalid_data_type")
}
