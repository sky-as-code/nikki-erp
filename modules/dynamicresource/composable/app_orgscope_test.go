package composable

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

func TestAssertActionRejectsMissingOrgId(t *testing.T) {
	domSvc, appSvc := newAppService(newOrgSchema())

	result, err := appSvc.Create(memberContext(), dmodel.DynamicFields{"name": "x"})

	require.NoError(t, err, "a missing org is the caller's mistake, not a server fault")
	assert.Positive(t, result.ClientErrors.Count())
	assert.Zero(t, domSvc.calls, "the domain service must not run without an org")
}

func TestAssertActionRejectsAnOrgTheCallerDoesNotBelongTo(t *testing.T) {
	domSvc, appSvc := newAppService(newOrgSchema())

	result, err := appSvc.Create(memberContext(), dmodel.DynamicFields{basemodel.FieldOrgId: "org_someone_else"})

	require.NoError(t, err)
	assert.Positive(t, result.ClientErrors.Count())
	assert.Zero(t, domSvc.calls)
}

func TestAssertActionAcceptsTheCallersOwnOrg(t *testing.T) {
	domSvc, appSvc := newAppService(newOrgSchema())

	result, err := appSvc.Create(memberContext(), dmodel.DynamicFields{basemodel.FieldOrgId: "org_mine"})

	require.NoError(t, err)
	assert.Zero(t, result.ClientErrors.Count())
	assert.Equal(t, "org_mine", domSvc.seen[basemodel.FieldOrgId])
}

// Without the service branch this fails: a service has an empty UserOrgIds, so membership alone
// would refuse every org-scoped action it ever attempts.
func TestAssertActionAcceptsAServicesOwnOrg(t *testing.T) {
	domSvc, appSvc := newAppService(newOrgSchema())

	result, err := appSvc.Create(serviceContext(util.ToPtr(model.Id("org_mine"))),
		dmodel.DynamicFields{basemodel.FieldOrgId: "org_mine"})

	require.NoError(t, err)
	assert.Zero(t, result.ClientErrors.Count())
	assert.Equal(t, 1, domSvc.calls)
}

func TestAssertActionRejectsAnOrgTheServiceWasNotMintedFor(t *testing.T) {
	domSvc, appSvc := newAppService(newOrgSchema())

	result, err := appSvc.Create(serviceContext(util.ToPtr(model.Id("org_mine"))),
		dmodel.DynamicFields{basemodel.FieldOrgId: "org_someone_else"})

	require.NoError(t, err)
	assert.Positive(t, result.ClientErrors.Count())
	assert.Zero(t, domSvc.calls)
}

func TestAssertActionRejectsAServiceWithNoOrg(t *testing.T) {
	domSvc, appSvc := newAppService(newOrgSchema())

	result, err := appSvc.Create(serviceContext(nil), dmodel.DynamicFields{basemodel.FieldOrgId: "org_mine"})

	require.NoError(t, err)
	assert.Positive(t, result.ClientErrors.Count())
	assert.Zero(t, domSvc.calls)
}

func TestOrgScopingIsSkippedWhenWithdrawn(t *testing.T) {
	domSvc, appSvc := newAppService(newOrgSchema(), NewAppServiceParam{IsOrgScoped: util.ToPtr(false)})

	result, err := appSvc.Create(memberContext(), dmodel.DynamicFields{"name": "global"})

	require.NoError(t, err)
	assert.Zero(t, result.ClientErrors.Count())
	assert.Equal(t, 1, domSvc.calls, "a withdrawn resource runs without an org")
}

func TestOrgScopingIsSkippedWhenSchemaHasNoOrgColumn(t *testing.T) {
	// A resource that cannot be org-filtered must keep working without one; this is what spares
	// settings and reference tables.
	domSvc, appSvc := newAppService(newPlainSchema())

	result, err := appSvc.Create(memberContext(), dmodel.DynamicFields{"name": "x"})

	require.NoError(t, err)
	assert.Zero(t, result.ClientErrors.Count())
	assert.Equal(t, 1, domSvc.calls)
}

func TestSearchGetsTheOrgConditionAppliedToItsGraph(t *testing.T) {
	domSvc, appSvc := newAppService(newOrgSchema())

	_, err := appSvc.Search(memberContext(), dmodel.DynamicFields{basemodel.FieldOrgId: "org_mine"})

	require.NoError(t, err)
	graph, ok := domSvc.seen[fieldNameGraph].(*dmodel.SearchGraph)
	require.True(t, ok, "search must reach the domain service carrying an org-constrained graph")
	assert.Len(t, graph.GetAnd(), 1)
}

func TestSearchNarrowsTheCallersGraphRatherThanReplacingIt(t *testing.T) {
	domSvc, appSvc := newAppService(newOrgSchema())

	_, err := appSvc.Search(memberContext(), dmodel.DynamicFields{
		basemodel.FieldOrgId: "org_mine",
		fieldNameGraph: map[string]any{
			"and": []any{map[string]any{"field": "name", "op": "=", "value": "x"}},
		},
	})

	require.NoError(t, err)
	graph, ok := domSvc.seen[fieldNameGraph].(*dmodel.SearchGraph)
	require.True(t, ok)
	assert.Len(t, graph.GetAnd(), 2, "the caller's node and the org node are ANDed together")
}

// A caller with no permission at all must be denied, whether or not they sent an org. The org
// step runs first (it produces the org the permission check needs), so this guards against the
// org error becoming a way to probe a resource you cannot access.
func TestOrgScopeDoesNotBypassThePermissionCheck(t *testing.T) {
	domSvc, appSvc := newAppService(newOrgSchema())

	result, err := appSvc.Create(deniedContext(), dmodel.DynamicFields{basemodel.FieldOrgId: "org_mine"})

	require.NoError(t, err)
	assert.Positive(t, result.ClientErrors.Count(), "a denied caller is still denied")
	assert.Zero(t, domSvc.calls)
}

// A row in another org answers "not found" rather than "forbidden": the caller is not entitled
// to learn that the id exists at all.
func TestSingleRowActionRefusesARecordOfAnotherOrg(t *testing.T) {
	domSvc, appSvc := newAppService(newOrgSchema())
	repo := domSvc.repo.(*stubRepository)
	repo.found = false

	result, err := appSvc.Update(memberContext(), dmodel.DynamicFields{
		"id": "rec_1", basemodel.FieldOrgId: "org_mine", "name": "renamed",
	})

	require.NoError(t, err)
	assert.Positive(t, result.ClientErrors.Count())
	assert.Zero(t, domSvc.calls, "the update must not run against a row of another org")
	assert.Equal(t, "org_mine", repo.fetchedKeys[basemodel.FieldOrgId], "the org travels with the id in the key set")
	assert.Equal(t, "rec_1", repo.fetchedKeys["id"])
}

func TestSingleRowActionRunsWhenTheRecordIsInTheOrg(t *testing.T) {
	domSvc, appSvc := newAppService(newOrgSchema())
	repo := domSvc.repo.(*stubRepository)
	repo.found = true
	repo.record = dmodel.DynamicFields{"id": "rec_1", basemodel.FieldOrgId: "org_mine"}

	for name, call := range map[string]func() (int, error){
		"update": func() (int, error) {
			r, err := appSvc.Update(memberContext(), dmodel.DynamicFields{"id": "rec_1", basemodel.FieldOrgId: "org_mine"})
			return r.ClientErrors.Count(), err
		},
		"delete": func() (int, error) {
			r, err := appSvc.Delete(memberContext(), dmodel.DynamicFields{"id": "rec_1", basemodel.FieldOrgId: "org_mine"})
			return r.ClientErrors.Count(), err
		},
		"set_archived": func() (int, error) {
			r, err := appSvc.SetArchived(memberContext(), dmodel.DynamicFields{"id": "rec_1", basemodel.FieldOrgId: "org_mine"})
			return r.ClientErrors.Count(), err
		},
		"get_by_id": func() (int, error) {
			r, err := appSvc.GetById(memberContext(), dmodel.DynamicFields{"id": "rec_1", basemodel.FieldOrgId: "org_mine"})
			return r.ClientErrors.Count(), err
		},
	} {
		t.Run(name, func(t *testing.T) {
			before := domSvc.calls
			errCount, err := call()
			require.NoError(t, err)
			assert.Zero(t, errCount)
			assert.Equal(t, before+1, domSvc.calls)
		})
	}
}

// FetchByKeys is what a custom action uses to load its record; a missing row is a client error
// reported through vErrs, never a Go error.
func TestFetchByKeysReportsAMissingRecordAsClientError(t *testing.T) {
	domSvc, appSvc := newAppService(newOrgSchema())
	domSvc.repo.(*stubRepository).found = false

	vErrs := newClientErrors()
	found, err := appSvc.FetchByKeys(memberContext(), dmodel.DynamicFields{"id": "missing"}, vErrs)

	require.NoError(t, err)
	assert.Nil(t, found)
	assert.Positive(t, vErrs.Count())
}

// AssertRecordInOrg leaves params without an id to the action's own validation.
func TestAssertRecordInOrgIgnoresParamsWithoutId(t *testing.T) {
	_, appSvc := newAppService(newOrgSchema())

	cErrs, err := appSvc.AssertRecordInOrg(memberContext(), dmodel.DynamicFields{}, "org_mine")

	require.NoError(t, err)
	assert.Nil(t, cErrs)
}
