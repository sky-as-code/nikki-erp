package baserepo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
)

// Coverage for edge selection and the select/filter validation, neither of which had any test
// in this package before. buildNestedSelectPlan decides which columns come from the main table
// and which are fetched per edge, so a mistake here is either a missing field in the response or
// a query that never had to run.

func TestEdgePlan_BareEdgeExpandsToDestinationFields(t *testing.T) {
	repo := virtualRepo(t)

	plan, cErrs := repo.buildNestedSelectPlan([]string{"id", "peer"})

	require.Zero(t, cErrs.Count(), "%v", cErrs)
	assert.Contains(t, plan.MainColumns, "id")
	assert.Contains(t, plan.MainColumns, "peer_id",
		"the foreign key must be selected or the edge cannot be resolved")
	assert.ElementsMatch(t, []string{"id", "name"}, plan.EdgeLeafColumns["peer"],
		"a bare edge expands to every readable field of the destination")
}

func TestEdgePlan_DottedPathNarrowsToNamedLeaf(t *testing.T) {
	repo := virtualRepo(t)

	plan, cErrs := repo.buildNestedSelectPlan([]string{"id", "peer.name"})

	require.Zero(t, cErrs.Count(), "%v", cErrs)
	assert.Equal(t, []string{"name"}, plan.EdgeLeafColumns["peer"])
	assert.Contains(t, plan.MainColumns, "peer_id")
}

// A virtual scalar has no column, so it must land in neither the main projection nor an edge --
// the service fills it after the read.
func TestEdgePlan_VirtualScalarStaysOutOfEdgeFetch(t *testing.T) {
	repo := virtualRepo(t)

	plan, cErrs := repo.buildNestedSelectPlan([]string{"id", "peer_name", "peer.name"})

	require.Zero(t, cErrs.Count(), "%v", cErrs)
	assert.NotContains(t, plan.EdgeLeafColumns["peer"], "peer_name")
}

func TestEdgePlan_UnknownEdgeRejected(t *testing.T) {
	repo := virtualRepo(t)

	_, cErrs := repo.buildNestedSelectPlan([]string{"id", "nosuch.name"})

	require.NotZero(t, cErrs.Count())
	assert.Contains(t, string(cErrs[0].Key), "err_unknown_schema_field")
}

func TestEdgePlan_UnknownLeafOnEdgeRejected(t *testing.T) {
	repo := virtualRepo(t)

	_, cErrs := repo.buildNestedSelectPlan([]string{"id", "peer.nosuch"})

	require.NotZero(t, cErrs.Count())
	assert.Contains(t, string(cErrs[0].Key), "err_unknown_schema_field")
}

// Selection is capped at one dot while filtering allows five. The limits differ by design, so
// the select-side cap is pinned here.
func TestEdgePlan_NestedPathBeyondOneDotRejected(t *testing.T) {
	repo := virtualRepo(t)

	_, cErrs := repo.buildNestedSelectPlan([]string{"peer.owner.name"})

	require.NotZero(t, cErrs.Count())
	assert.Contains(t, string(cErrs[0].Key), "err_graph_field_path_too_deep")
	assert.Equal(t, 1, orm.MaxSelectGraphColumnDots, "the select-side limit this test pins")
}

// F4: GetOne validated its field list up front while Search did not, so the two accepted
// different projections and reported different errors for the same mistake. They now share one
// validator -- these cases pin that they agree.
func TestSelectValidation_GetOneAndSearchAgree(t *testing.T) {
	repo := virtualRepo(t)

	testCases := []struct {
		name    string
		columns []string
		wantErr bool
	}{
		{"plain columns", []string{"id", "sku"}, false},
		{"virtual scalar", []string{"id", "peer_name"}, false},
		{"bare edge", []string{"id", "peer"}, false},
		{"dotted edge field", []string{"id", "peer.name"}, false},
		{"unknown field", []string{"id", "nosuch"}, true},
		{"unknown edge", []string{"id", "nosuch.name"}, true},
		{"unknown leaf", []string{"id", "peer.nosuch"}, true},
		{"too deep", []string{"peer.owner.name"}, true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			vErr := repo.validateSelectColumns(testCase.columns)

			if testCase.wantErr {
				require.NotNil(t, vErr, "Search and GetOne must both reject this")
				return
			}
			require.Nil(t, vErr, "Search and GetOne must both accept this")
		})
	}
}
