package baserepo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	archivableSchemaName    = "test_baserepo_archivable"
	nonArchivableSchemaName = "test_baserepo_plain"
)

// archivedRepos builds one repository over a schema carrying is_archived and one over a schema
// without it, which is the minimum needed to prove the column guard.
func archivedRepos(t *testing.T) (archivable *BaseDynamicRepositoryImpl, plain *BaseDynamicRepositoryImpl) {
	t.Helper()
	registry := dmodel.GetSchemaRegistry()
	if registry.Get(archivableSchemaName) == nil {
		require.NoError(t, dmodel.RegisterSchemaB(
			dmodel.DefineModel(archivableSchemaName).
				TableName("test_baserepo_archivables").
				ShouldBuildDb().
				Field(dmodel.DefineField().Name("id").
					DataType(dmodel.FieldDataTypeUlid()).RequiredForCreate().PrimaryKey()).
				Field(dmodel.DefineField().Name("name").
					DataType(dmodel.FieldDataTypeString(0, 200))).
				Field(dmodel.DefineField().Name(basemodel.FieldIsArchived).
					DataType(dmodel.FieldDataTypeBoolean()).Default(false))))
	}
	if registry.Get(nonArchivableSchemaName) == nil {
		require.NoError(t, dmodel.RegisterSchemaB(
			dmodel.DefineModel(nonArchivableSchemaName).
				TableName("test_baserepo_plains").
				ShouldBuildDb().
				Field(dmodel.DefineField().Name("id").
					DataType(dmodel.FieldDataTypeUlid()).RequiredForCreate().PrimaryKey()).
				Field(dmodel.DefineField().Name("name").
					DataType(dmodel.FieldDataTypeString(0, 200)))))
	}

	return &BaseDynamicRepositoryImpl{schema: registry.Get(archivableSchemaName)},
		&BaseDynamicRepositoryImpl{schema: registry.Get(nonArchivableSchemaName)}
}

// The single most important guard in this change.
//
// Around 29 call sites reuse Search as an internal lookup and several of them need archived rows
// by design (models.FindTemplateVariants documents exactly that). They all leave IncludeArchived
// unset, so a nil must be returned untouched -- identical graph, therefore identical SQL.
func TestIsArchived_NilLeavesGraphUntouched(t *testing.T) {
	repo, _ := archivedRepos(t)
	graph := dmodel.NewSearchGraph().NewCondition("name", dmodel.Equals, "Anna")

	got := repo.injectIsArchivedIntoGraph(graph, nil)

	assert.Same(t, graph, got, "a nil IncludeArchived must not rebuild the graph")
}

func TestIsArchived_NilLeavesNilGraphNil(t *testing.T) {
	repo, _ := archivedRepos(t)

	assert.Nil(t, repo.injectIsArchivedIntoGraph(nil, nil))
}

// True means "archived alongside active", which is the same absence of a filter as nil.
func TestIsArchived_TrueLeavesGraphUntouched(t *testing.T) {
	repo, _ := archivedRepos(t)
	graph := dmodel.NewSearchGraph().NewCondition("name", dmodel.Equals, "Anna")

	got := repo.injectIsArchivedIntoGraph(graph, util.ToPtr(true))

	assert.Same(t, graph, got)
}

// "Only archived" is expressed as IncludeArchived=true plus an explicit is_archived=true in the
// caller's own graph: true suppresses the injection, so the caller's condition is the only one
// left and narrows the result to archived records. This is the supported way to list the archive,
// and it is why true must return the graph byte-identical rather than merging anything into it.
func TestIsArchived_TruePreservesCallerOnlyArchivedFilter(t *testing.T) {
	repo, _ := archivedRepos(t)
	graph := dmodel.NewSearchGraph().NewCondition(basemodel.FieldIsArchived, dmodel.Equals, true)

	got := repo.injectIsArchivedIntoGraph(graph, util.ToPtr(true))

	require.Same(t, graph, got, "the caller's own is_archived filter must survive untouched")
	cond := got.GetCondition()
	assert.Equal(t, basemodel.FieldIsArchived, cond.Field())
	assert.Equal(t, true, cond.Value(), "an only-archived search must stay only-archived")
}

func TestIsArchived_FalsePrependsConditionToNilGraph(t *testing.T) {
	repo, _ := archivedRepos(t)

	got := repo.injectIsArchivedIntoGraph(nil, util.ToPtr(false))

	require.NotNil(t, got)
	require.Len(t, got.GetAnd(), 1)
	cond := got.GetAnd()[0].GetCondition()
	assert.Equal(t, basemodel.FieldIsArchived, cond.Field())
	assert.Equal(t, dmodel.Equals, cond.Operator())
	assert.Equal(t, false, cond.Value())
}

// The caller's graph must be ANDed in, not replaced -- SearchGraph.And is a destructive setter,
// so the old graph has to be wrapped via ToSearchNode rather than mutated.
func TestIsArchived_FalseAndsWithCallerGraph(t *testing.T) {
	repo, _ := archivedRepos(t)
	graph := dmodel.NewSearchGraph().NewCondition("name", dmodel.Equals, "Anna")

	got := repo.injectIsArchivedIntoGraph(graph, util.ToPtr(false))

	require.NotNil(t, got)
	require.Len(t, got.GetAnd(), 2, "both the injected condition and the caller graph must survive")
	assert.Equal(t, basemodel.FieldIsArchived, got.GetAnd()[0].GetCondition().Field())
	assert.Equal(t, "name", got.GetAnd()[1].GetCondition().Field())
}

// ToSearchNode drops the order, so it has to be re-applied onto the rebuilt graph.
func TestIsArchived_FalsePreservesOrder(t *testing.T) {
	repo, _ := archivedRepos(t)
	graph := dmodel.NewSearchGraph().
		NewCondition("name", dmodel.Equals, "Anna").
		OrderBy("name", dmodel.Desc)

	got := repo.injectIsArchivedIntoGraph(graph, util.ToPtr(false))

	require.NotNil(t, got)
	assert.Equal(t, graph.GetOrder(), got.GetOrder(), "the caller's order must survive the rewrite")
}

// An order-only graph carries no condition/and/or. Folding it in as an AND member would emit an
// empty predicate, and SearchGraph.validate accepts zero-of-three, so nothing else catches it.
// This is the shape a plain listing request produces, so it is the common case, not an edge case.
func TestIsArchived_FalseOnOrderOnlyGraphEmitsNoEmptyNode(t *testing.T) {
	repo, _ := archivedRepos(t)
	graph := dmodel.NewSearchGraph().OrderBy("name", dmodel.Desc)

	got := repo.injectIsArchivedIntoGraph(graph, util.ToPtr(false))

	require.NotNil(t, got)
	require.Len(t, got.GetAnd(), 1, "an order-only graph must not contribute an empty AND member")
	assert.Equal(t, basemodel.FieldIsArchived, got.GetAnd()[0].GetCondition().Field())
	assert.Equal(t, graph.GetOrder(), got.GetOrder())
}

// is_archived comes from the optional ArchivableModelSchemaBuilder mixin, so a schema without
// the column must be left alone rather than handed a condition the query builder would reject.
func TestIsArchived_FalseSkippedOnSchemaWithoutColumn(t *testing.T) {
	_, plain := archivedRepos(t)
	graph := dmodel.NewSearchGraph().NewCondition("name", dmodel.Equals, "Anna")

	got := plain.injectIsArchivedIntoGraph(graph, util.ToPtr(false))

	assert.Same(t, graph, got, "a schema without is_archived must not be filtered")
}
