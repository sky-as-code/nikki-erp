package orm

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// A LEFT JOIN across a one:many or many:many edge repeats each root row once per matching child.
// Without DISTINCT the list shows duplicates and COUNT(*) reports an inflated Total -- the caller
// sees "37 results" for 12 real records. These tests pin the fix and, just as importantly, pin
// that a many:one join still pays nothing for it.

const (
	distinctParentSchema = "test_distinct_parent"
	distinctChildSchema  = "test_distinct_child"
	distinctPeerSchema   = "test_distinct_peer"
)

func distinctSchemas(t *testing.T) (*dmodel.ModelSchema, *dmodel.SchemaRegistry) {
	t.Helper()
	registry := dmodel.GetSchemaRegistry()
	if registry.Get(distinctParentSchema) != nil {
		return registry.Get(distinctParentSchema), registry
	}

	require.NoError(t, dmodel.RegisterSchemaB(
		dmodel.DefineModel(distinctPeerSchema).
			TableName("test_distinct_peers").
			ShouldBuildDb().
			Field(dmodel.DefineField().Name("id").
				DataType(dmodel.FieldDataTypeUlid()).RequiredForCreate().PrimaryKey()).
			Field(dmodel.DefineField().Name("title").
				DataType(dmodel.FieldDataTypeString(0, 50)))))

	require.NoError(t, dmodel.RegisterSchemaB(
		dmodel.DefineModel(distinctParentSchema).
			TableName("test_distinct_parents").
			ShouldBuildDb().
			Field(dmodel.DefineField().Name("id").
				DataType(dmodel.FieldDataTypeUlid()).RequiredForCreate().PrimaryKey()).
			Field(dmodel.DefineField().Name("peer_id").
				DataType(dmodel.FieldDataTypeUlid()).RequiredForCreate()).
			Field(dmodel.DefineField().Name("code").
				DataType(dmodel.FieldDataTypeString(0, 50))).
			EdgeTo(dmodel.Edge("peer").ManyToOne(
				distinctPeerSchema, dmodel.DynamicFields{"peer_id": "id"})).
			EdgeFrom(dmodel.Edge("children").Existing(distinctChildSchema, "parent"))))

	require.NoError(t, dmodel.RegisterSchemaB(
		dmodel.DefineModel(distinctChildSchema).
			TableName("test_distinct_children").
			ShouldBuildDb().
			Field(dmodel.DefineField().Name("id").
				DataType(dmodel.FieldDataTypeUlid()).RequiredForCreate().PrimaryKey()).
			Field(dmodel.DefineField().Name("parent_id").
				DataType(dmodel.FieldDataTypeUlid()).RequiredForCreate()).
			Field(dmodel.DefineField().Name("label").
				DataType(dmodel.FieldDataTypeString(0, 50))).
			EdgeTo(dmodel.Edge("parent").ManyToOne(
				distinctParentSchema, dmodel.DynamicFields{"parent_id": "id"}))))

	require.NoError(t, registry.FinalizeRelations())
	return registry.Get(distinctParentSchema), registry
}

func selectSqlFiltering(t *testing.T, path string, columns []string) string {
	t.Helper()
	schema, registry := distinctSchemas(t)
	graph := dmodel.NewSearchGraph()
	graph.NewCondition(path, dmodel.Equals, "x")

	sql, cErrs, err := (&PgQueryBuilder{}).SqlSelectGraph(schema, registry, graph, SqlSelectGraphOpts{
		Columns: ToSelectColumns(columns),
	})
	require.NoError(t, err)
	require.Nil(t, cErrs)
	require.NotNil(t, sql)
	return *sql
}

// A many:one join is cardinality-preserving, so forcing DISTINCT on it would be pure cost.
func TestDistinct_ManyToOneJoinDoesNotForceDistinct(t *testing.T) {
	sql := selectSqlFiltering(t, "peer.title", []string{"id", "code"})

	assert.NotContains(t, sql, "DISTINCT")
	assert.Contains(t, sql, "LEFT JOIN")
}

func TestDistinct_OneToManyJoinForcesDistinct(t *testing.T) {
	sql := selectSqlFiltering(t, "children.label", []string{"id", "code"})

	assert.Contains(t, sql, "SELECT DISTINCT")
}

// The count must share the list's grain, or Total contradicts the rows beside it.
func TestDistinct_CountOverOneToManyUsesSubquery(t *testing.T) {
	schema, registry := distinctSchemas(t)
	graph := dmodel.NewSearchGraph()
	graph.NewCondition("children.label", dmodel.Equals, "x")

	sql, cErrs, err := (&PgQueryBuilder{}).SqlCountGraph(schema, registry, graph, SqlSelectGraphOpts{
		Columns: ToSelectColumns([]string{"id", "code"}),
	})

	require.NoError(t, err)
	require.Nil(t, cErrs)
	require.NotNil(t, sql)
	assert.Contains(t, *sql, "SELECT COUNT(*) FROM (SELECT DISTINCT")
}

// The guard against over-application: a many:one filter must keep the cheap plain count.
func TestDistinct_CountOverManyToOneStaysPlain(t *testing.T) {
	schema, registry := distinctSchemas(t)
	graph := dmodel.NewSearchGraph()
	graph.NewCondition("peer.title", dmodel.Equals, "x")

	sql, cErrs, err := (&PgQueryBuilder{}).SqlCountGraph(schema, registry, graph, SqlSelectGraphOpts{
		Columns: ToSelectColumns([]string{"id", "code"}),
	})

	require.NoError(t, err)
	require.Nil(t, cErrs)
	require.NotNil(t, sql)
	assert.Contains(t, *sql, "SELECT COUNT(*)")
	assert.NotContains(t, *sql, "_distinct_count")
}

// The explicit opt-in predates this change and must keep working on its own.
func TestDistinct_ExplicitTokenStillWorksWithoutJoins(t *testing.T) {
	schema, registry := distinctSchemas(t)

	sql, cErrs, err := (&PgQueryBuilder{}).SqlSelectGraph(schema, registry, nil, SqlSelectGraphOpts{
		Columns: []SelectColumn{SelectColumn("id").AsDistinct(), SelectColumn("code")},
	})

	require.NoError(t, err)
	require.Nil(t, cErrs)
	require.NotNil(t, sql)
	assert.Contains(t, *sql, "SELECT DISTINCT")
}

// Both routes to DISTINCT together must still yield exactly one.
func TestDistinct_ExplicitTokenPlusFanOutYieldsOneDistinct(t *testing.T) {
	schema, registry := distinctSchemas(t)
	graph := dmodel.NewSearchGraph()
	graph.NewCondition("children.label", dmodel.Equals, "x")

	sql, cErrs, err := (&PgQueryBuilder{}).SqlSelectGraph(schema, registry, graph, SqlSelectGraphOpts{
		Columns: []SelectColumn{SelectColumn("id").AsDistinct(), SelectColumn("code")},
	})

	require.NoError(t, err)
	require.Nil(t, cErrs)
	require.NotNil(t, sql)
	assert.Equal(t, 1, strings.Count(*sql, "DISTINCT"))
}

// PostgreSQL rejects "SELECT DISTINCT a ORDER BY b" outright. Auto-DISTINCT would introduce
// exactly that whenever a caller sorts on a joined column, so the ordered ref is projected.
func TestDistinct_OrderedJoinedColumnIsProjected(t *testing.T) {
	schema, registry := distinctSchemas(t)
	graph := dmodel.NewSearchGraph()
	graph.NewCondition("children.label", dmodel.Equals, "x")
	graph.OrderBy("children.label", dmodel.Asc)

	sql, cErrs, err := (&PgQueryBuilder{}).SqlSelectGraph(schema, registry, graph, SqlSelectGraphOpts{
		Columns: ToSelectColumns([]string{"id"}),
	})

	require.NoError(t, err)
	require.Nil(t, cErrs)
	require.NotNil(t, sql)

	selectClause := (*sql)[:strings.Index(*sql, " FROM ")]
	assert.Contains(t, selectClause, `"label"`,
		"the ordered column must appear in the select list under DISTINCT")
}

// Without DISTINCT there is nothing to reconcile, so the projection must be left alone.
func TestDistinct_OrderedColumnNotProjectedWhenNoFanOut(t *testing.T) {
	schema, registry := distinctSchemas(t)
	graph := dmodel.NewSearchGraph()
	graph.NewCondition("peer.title", dmodel.Equals, "x")
	graph.OrderBy("peer.title", dmodel.Asc)

	sql, cErrs, err := (&PgQueryBuilder{}).SqlSelectGraph(schema, registry, graph, SqlSelectGraphOpts{
		Columns: ToSelectColumns([]string{"id"}),
	})

	require.NoError(t, err)
	require.Nil(t, cErrs)
	require.NotNil(t, sql)

	selectClause := (*sql)[:strings.Index(*sql, " FROM ")]
	assert.NotContains(t, selectClause, `"title"`)
}

// EXISTS is unaffected by duplicates, so it must not pay for DISTINCT.
func TestDistinct_ExistsGraphUnaffected(t *testing.T) {
	schema, registry := distinctSchemas(t)
	graph := dmodel.NewSearchGraph()
	graph.NewCondition("children.label", dmodel.Equals, "x")

	sql, cErrs, err := (&PgQueryBuilder{}).SqlExistsGraph(schema, registry, graph)

	require.NoError(t, err)
	require.Nil(t, cErrs)
	require.NotNil(t, sql)
	assert.NotContains(t, *sql, "DISTINCT")
}
