package usagecheck

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

// recordingSearcher answers from a fixed set of rows, applying the graph it is given the way the
// database would. That is the point: a checker whose conditions are dropped reads as "every row
// matches", which is exactly the defect these tests exist to catch.
type recordingSearcher struct {
	rows        []dmodel.DynamicFields
	schema      *dmodel.ModelSchema
	lastGraph   *dmodel.SearchGraph
	searchErr   error
	searchCount int
}

func (this *recordingSearcher) Schema() *dmodel.ModelSchema { return this.schema }

func (this *recordingSearcher) Search(
	_ corectx.Context, param dyn.RepoSearchParam,
) (*dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]], error) {
	this.searchCount++
	this.lastGraph = param.Graph
	if this.searchErr != nil {
		return nil, this.searchErr
	}

	matched := make([]dmodel.DynamicFields, 0, len(this.rows))
	for _, row := range this.rows {
		if rowMatches(row, param.Graph) {
			matched = append(matched, row)
		}
	}
	if param.Size > 0 && len(matched) > param.Size {
		matched = matched[:param.Size]
	}

	return &dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]]{
		HasData: len(matched) > 0,
		Data:    dyn.PagedResultData[dmodel.DynamicFields]{Items: matched},
	}, nil
}

// rowMatches applies every equality condition in the graph. A graph carrying no conditions
// matches everything, which is how the dropped-condition bug presented.
func rowMatches(row dmodel.DynamicFields, graph *dmodel.SearchGraph) bool {
	if graph == nil {
		return true
	}
	for _, node := range graph.GetAnd() {
		condition := node.GetCondition()
		if condition.Field() == "" {
			continue
		}
		if row[condition.Field()] != condition.Value() {
			return false
		}
	}
	return true
}

func searcherWith(rows ...dmodel.DynamicFields) *recordingSearcher {
	return &recordingSearcher{
		rows: rows,
		schema: dmodel.DefineModel("test_referencing").
			TableName("test_referencing").
			ShouldBuildDb().
			Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
			Field(dmodel.DefineField().Name("uom_id").DataType(dmodel.FieldDataTypeUlid())).
			Field(dmodel.DefineField().Name("org_id").DataType(dmodel.FieldDataTypeUlid())).
			Build(),
	}
}

func uomRef(id string, orgId string) ResourceRef {
	return ResourceRef{ResourceName: ResourceUom, Identifier: NewIdentifier(id, orgId)}
}

func checkerOver(searcher *recordingSearcher) ResourceUsageChecker {
	return NewTableChecker(
		func(string) (RowSearcher, error) { return searcher, nil },
		ReferencingTable{SchemaName: "test_referencing", Field: "uom_id"},
	)
}

// THE REGRESSION THIS FILE EXISTS FOR.
//
// SearchGraph.And REPLACES its node list rather than appending, so building the graph with two
// And calls — one for the reference, one for the org — silently dropped the reference condition.
// The search then matched every row in the table and reported a brand-new, entirely unreferenced
// unit as "in use by inventory, sales". It reached a live server before anything caught it,
// because a fake repository that ignores the graph cannot tell the two apart.
func TestTableChecker_UnreferencedResourceIsNotInUse(t *testing.T) {
	searcher := searcherWith(
		dmodel.DynamicFields{"id": "ROW-1", "uom_id": "OTHER-UOM", "org_id": "ORG-1"},
		dmodel.DynamicFields{"id": "ROW-2", "uom_id": "ANOTHER-UOM", "org_id": "ORG-1"},
	)

	used, _, err := checkerOver(searcher).IsUsed(testCtx(), uomRef("UOM-1", "ORG-1"))

	require.NoError(t, err)
	assert.False(t, used, "no row references this unit, so it is not in use")
}

func TestTableChecker_ReferencedResourceIsInUse(t *testing.T) {
	searcher := searcherWith(
		dmodel.DynamicFields{"id": "ROW-1", "uom_id": "OTHER-UOM", "org_id": "ORG-1"},
		dmodel.DynamicFields{"id": "ROW-2", "uom_id": "UOM-1", "org_id": "ORG-1"},
	)

	used, usedBy, err := checkerOver(searcher).IsUsed(testCtx(), uomRef("UOM-1", "ORG-1"))

	require.NoError(t, err)
	assert.True(t, used)
	assert.Equal(t, "test_referencing", usedBy, "the answer names the table, never the row")
}

// Both conditions must survive into the one graph the repository receives.
func TestTableChecker_SearchesOnReferenceAndOrgTogether(t *testing.T) {
	searcher := searcherWith()

	_, _, err := checkerOver(searcher).IsUsed(testCtx(), uomRef("UOM-1", "ORG-1"))
	require.NoError(t, err)

	fields := map[string]any{}
	for _, node := range searcher.lastGraph.GetAnd() {
		if condition := node.GetCondition(); condition.Field() != "" {
			fields[condition.Field()] = condition.Value()
		}
	}

	assert.Equal(t, map[string]any{"uom_id": "UOM-1", "org_id": "ORG-1"}, fields,
		"dropping either condition changes what the search means")
}

// A reference held by another organization is that organization's business: counting it would
// block the wrong delete and disclose that the row exists.
func TestTableChecker_IgnoresAnotherOrgsReference(t *testing.T) {
	searcher := searcherWith(
		dmodel.DynamicFields{"id": "ROW-1", "uom_id": "UOM-1", "org_id": "ORG-2"},
	)

	used, _, err := checkerOver(searcher).IsUsed(testCtx(), uomRef("UOM-1", "ORG-1"))

	require.NoError(t, err)
	assert.False(t, used)
}

// A failed lookup must surface as an error, never as "not used": the caller blocks on an error
// and permits the delete on a false.
func TestTableChecker_SearchFailureIsAnError(t *testing.T) {
	searcher := searcherWith()
	searcher.searchErr = errors.New("connection lost")

	used, _, err := checkerOver(searcher).IsUsed(testCtx(), uomRef("UOM-1", "ORG-1"))

	require.Error(t, err)
	assert.False(t, used)
}

// An unresolvable repository is an error for the same reason.
func TestTableChecker_UnresolvableRepositoryIsAnError(t *testing.T) {
	checker := NewTableChecker(
		func(string) (RowSearcher, error) { return nil, errors.New("engine not built") },
		ReferencingTable{SchemaName: "test_referencing", Field: "uom_id"},
	)

	_, _, err := checker.IsUsed(testCtx(), uomRef("UOM-1", "ORG-1"))

	assert.Error(t, err)
}

// The first table that answers "yes" ends the check: the rest cannot change the outcome.
func TestTableChecker_StopsAtTheFirstMatch(t *testing.T) {
	first := searcherWith(dmodel.DynamicFields{"id": "R", "uom_id": "UOM-1", "org_id": "ORG-1"})
	second := searcherWith()

	checker := NewTableChecker(
		func(schemaName string) (RowSearcher, error) {
			if schemaName == "first" {
				return first, nil
			}
			return second, nil
		},
		ReferencingTable{SchemaName: "first", Field: "uom_id"},
		ReferencingTable{SchemaName: "second", Field: "uom_id"},
	)

	used, usedBy, err := checker.IsUsed(testCtx(), uomRef("UOM-1", "ORG-1"))

	require.NoError(t, err)
	assert.True(t, used)
	assert.Equal(t, "first", usedBy)
	assert.Zero(t, second.searchCount, "the second table is not queried once the answer is settled")
}

// An empty identifier means there is nothing to look for, and no query is worth making.
func TestTableChecker_EmptyIdIsNotInUse(t *testing.T) {
	searcher := searcherWith(dmodel.DynamicFields{"id": "R", "uom_id": "UOM-1"})

	used, _, err := checkerOver(searcher).IsUsed(testCtx(), uomRef("", "ORG-1"))

	require.NoError(t, err)
	assert.False(t, used)
	assert.Zero(t, searcher.searchCount)
}
