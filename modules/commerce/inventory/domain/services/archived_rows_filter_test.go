package services

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

func useGraphCapturingRepository(t *testing.T) *graphCapturingRepository {
	t.Helper()
	repo := &graphCapturingRepository{}
	original := repoFor
	t.Cleanup(func() { repoFor = original })
	repoFor = func(string) (composable.CrudRepository, error) { return repo, nil }
	return repo
}

func assertSearchedUnarchivedOnly(t *testing.T, repo *graphCapturingRepository) {
	t.Helper()
	require.Len(t, repo.graphs, 1)
	encoded, err := json.Marshal(repo.graphs[0])
	require.NoError(t, err)
	assert.Contains(t, string(encoded), `"is_archived"`)
}

type graphCapturingRepository struct {
	composable.CrudRepository
	graphs []*dmodel.SearchGraph
}

func (this *graphCapturingRepository) Search(
	_ corectx.Context, param dyn.RepoSearchParam,
) (*dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]], error) {
	this.graphs = append(this.graphs, param.Graph)
	return &dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]]{}, nil
}

// An archived putaway rule is retired: it must not be read as a candidate, or the suggestion
// keeps sending goods somewhere an operator has deliberately stopped using.
func TestCandidatePutawayRulesExcludeArchivedOnes(t *testing.T) {
	repo := useGraphCapturingRepository(t)

	_, err := findCandidatePutawayRules(callerContext(), PutawayContext{ArrivalLocationId: "01ARRIVAL"})

	require.NoError(t, err)
	assertSearchedUnarchivedOnly(t, repo)
}

// An archived child no longer sits under its parent in any working sense, so it must not block
// archiving the parent; counting it made a cleaned-up tree impossible to retire.
func TestChildCountsIgnoreArchivedChildren(t *testing.T) {
	repo := useGraphCapturingRepository(t)
	_, err := (&WarehouseDomainServiceImpl{}).countUnarchivedChildren(callerContext(), "01PARENT")
	require.NoError(t, err)
	assertSearchedUnarchivedOnly(t, repo)

	repo = useGraphCapturingRepository(t)
	_, err = (&InventoryLocationDomainServiceImpl{}).countUnarchivedChildren(callerContext(), "01PARENT")
	require.NoError(t, err)
	assertSearchedUnarchivedOnly(t, repo)
}
