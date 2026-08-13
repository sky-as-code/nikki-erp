package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// dateTimeSchema builds a table with one field of each date-family column type, so a filter graph
// condition can be built against each without a database.
func dateTimeSchema(t *testing.T) (*dmodel.ModelSchema, *dmodel.SchemaRegistry) {
	t.Helper()
	registry := dmodel.GetSchemaRegistry()
	if existing := registry.Get(dateTimeTestSchemaName); existing != nil {
		return existing, registry
	}

	builder := dmodel.DefineModel(dateTimeTestSchemaName).
		TableName("test_date_types").
		ShouldBuildDb().
		Field(dmodel.DefineField().Name("id").
			DataType(dmodel.FieldDataTypeUlid()).RequiredForCreate().PrimaryKey()).
		Field(dmodel.DefineField().Name("due_date").
			DataType(dmodel.FieldDataTypeDate())).
		Field(dmodel.DefineField().Name("due_time").
			DataType(dmodel.FieldDataTypeTime())).
		Field(dmodel.DefineField().Name("due_at").
			DataType(dmodel.FieldDataTypeDateTime()))

	require.NoError(t, dmodel.RegisterSchemaB(builder))
	return registry.Get(dateTimeTestSchemaName), registry
}

const dateTimeTestSchemaName = "test_date_type"

// A nikkiDate field filtered with the same YYYY-MM-DD format its own JSON layer accepts must not
// be rejected as "invalid data type" — the bug the counts-due worklist filter hit.
func TestDateFilter_AcceptsDateOnlyString(t *testing.T) {
	schema, registry := dateTimeSchema(t)

	graph := dmodel.NewSearchGraph()
	graph.NewCondition("due_date", dmodel.LessEqual, "2026-08-13")

	sql, cErrs, err := (&PgQueryBuilder{}).SqlSelectGraph(schema, registry, graph, SqlSelectGraphOpts{})

	require.NoError(t, err)
	require.Nil(t, cErrs, "expected no client error, got %v", cErrs)
	require.NotNil(t, sql)
	assert.Contains(t, *sql, `"due_date"`)
}

// Same gap existed for nikkiTime, not just nikkiDate.
func TestTimeFilter_AcceptsTimeOnlyString(t *testing.T) {
	schema, registry := dateTimeSchema(t)

	graph := dmodel.NewSearchGraph()
	graph.NewCondition("due_time", dmodel.LessEqual, "17:30:00")

	_, cErrs, err := (&PgQueryBuilder{}).SqlSelectGraph(schema, registry, graph, SqlSelectGraphOpts{})

	require.NoError(t, err)
	require.Nil(t, cErrs, "expected no client error, got %v", cErrs)
}

// nikkiDateTime already worked before this fix; pinned so the added switch does not regress it.
func TestDateTimeFilter_StillAcceptsRfc3339(t *testing.T) {
	schema, registry := dateTimeSchema(t)

	graph := dmodel.NewSearchGraph()
	graph.NewCondition("due_at", dmodel.LessEqual, "2026-08-13T00:00:00Z")

	_, cErrs, err := (&PgQueryBuilder{}).SqlSelectGraph(schema, registry, graph, SqlSelectGraphOpts{})

	require.NoError(t, err)
	require.Nil(t, cErrs, "expected no client error, got %v", cErrs)
}

// A genuinely malformed date must still be refused, and its message's {{typeName}} must now
// actually be interpolatable — the response the user saw had the template but no vars.
func TestDateFilter_RejectsMalformedDateWithVars(t *testing.T) {
	schema, registry := dateTimeSchema(t)

	graph := dmodel.NewSearchGraph()
	graph.NewCondition("due_date", dmodel.LessEqual, "not-a-date")

	_, cErrs, err := (&PgQueryBuilder{}).SqlSelectGraph(schema, registry, graph, SqlSelectGraphOpts{})

	require.NoError(t, err)
	require.NotNil(t, cErrs, "expected a client error")
	item := (*cErrs)[0]
	assert.Equal(t, "due_date", item.Field)
	assert.Contains(t, string(item.Key), "err_invalid_data_type")
	require.NotNil(t, item.Vars, "vars must be populated so {{typeName}} can be interpolated")
	assert.NotEmpty(t, item.Vars["typeName"])
}
