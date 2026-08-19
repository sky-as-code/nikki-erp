package requestguard

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	ds "github.com/sky-as-code/nikki-erp/common/datastructure"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// The permission system has three executors: the in-memory guard used by the
// middleware, the SQL matcher used by cross-module `isAuthorized` calls, and the
// self-service permission probe. All three answer by asking whether the user holds
// any CandidateExpressions row, so a scenario table run against the candidate set
// is the parity contract they share - if this table passes, they cannot disagree.
//
// The SQL matcher's own execution is exercised by the API suite; what is proven
// here is that the rule set it consumes is the same one the guard consumes.

type parityScenario struct {
	name string
	// Expressions the user holds in the cache.
	grants []string
	// The question being asked.
	required Perm
	evalCtx  EvalContext
	isOwner  bool
	expected bool
}

func orgId() *model.Id  { return idOf("ORG1") }
func unitId() *model.Id { return idOf("OU1") }

func parityScenarios() []parityScenario {
	return []parityScenario{
		{
			name:     "owner is allowed without holding anything",
			isOwner:  true,
			required: PermFor("create", "iam_user", ResourceScopeDomain),
			expected: true,
		},
		{
			name:     "no grants denies",
			required: PermFor("create", "iam_user", ResourceScopeDomain),
			expected: false,
		},
		{
			name:     "exact domain grant",
			grants:   []string{"create:iam_user:domain"},
			required: PermFor("create", "iam_user", ResourceScopeDomain),
			expected: true,
		},
		{
			name:     "omnipotent grant answers everything",
			grants:   []string{"*:*:*"},
			required: PermFor("delete", "inventory_product", ResourceScopeOrgUnit).InOrgUnit(unitId()),
			expected: true,
		},
		{
			// D4 regression: the SQL matcher used to omit this candidate entirely,
			// so the guard allowed and the matcher denied the same question.
			name:     "action wildcard at domain answers an org question",
			grants:   []string{"create:*:domain"},
			required: PermFor("create", "iam_user", ResourceScopeOrg).InOrg(orgId()),
			expected: true,
		},
		{
			name:     "resource wildcard at domain answers an org question",
			grants:   []string{"*:iam_user:domain"},
			required: PermFor("create", "iam_user", ResourceScopeOrg).InOrg(orgId()),
			expected: true,
		},
		{
			name:     "exact org grant answers its own org",
			grants:   []string{"create:iam_user:org/ORG1"},
			required: PermFor("create", "iam_user", ResourceScopeOrg).InOrg(orgId()),
			expected: true,
		},
		{
			name:     "org grant does not answer a different org",
			grants:   []string{"create:iam_user:org/ORG2"},
			required: PermFor("create", "iam_user", ResourceScopeOrg).InOrg(orgId()),
			expected: false,
		},
		{
			name:     "bare org grant answers only for a member",
			grants:   []string{"create:iam_user:org"},
			required: PermFor("create", "iam_user", ResourceScopeOrg).InOrg(orgId()),
			evalCtx:  EvalContext{UserOrgIds: []model.Id{*orgId()}},
			expected: true,
		},
		{
			name:     "bare org grant denies a non-member",
			grants:   []string{"create:iam_user:org"},
			required: PermFor("create", "iam_user", ResourceScopeOrg).InOrg(orgId()),
			evalCtx:  EvalContext{UserOrgIds: []model.Id{model.Id("ORG_OTHER")}},
			expected: false,
		},
		{
			name:     "exact unit grant answers its own unit",
			grants:   []string{"create:iam_user:orgunit/OU1"},
			required: PermFor("create", "iam_user", ResourceScopeOrgUnit).InOrgUnit(unitId()),
			expected: true,
		},
		{
			name:     "unit grant does not answer a different unit",
			grants:   []string{"create:iam_user:orgunit/OU_OTHER"},
			required: PermFor("create", "iam_user", ResourceScopeOrgUnit).InOrgUnit(unitId()),
			expected: false,
		},
		{
			// D4 regression: the matcher used to build `...:org/{orgUnitId}` here -
			// an org expression carrying a unit id, which matches nothing real.
			name:     "org grant for the unit's org answers a unit question",
			grants:   []string{"create:iam_user:org/ORG1"},
			required: PermFor("create", "iam_user", ResourceScopeOrgUnit).InOrgUnit(unitId()).InOrg(orgId()),
			expected: true,
		},
		{
			name:     "bare unit grant answers when the caller is in that unit",
			grants:   []string{"create:iam_user:orgunit"},
			required: PermFor("create", "iam_user", ResourceScopeOrgUnit).InOrgUnit(unitId()),
			evalCtx:  EvalContext{OrgUnitId: unitId()},
			expected: true,
		},
		{
			name:     "bare unit grant denies a caller in a different unit",
			grants:   []string{"create:iam_user:orgunit"},
			required: PermFor("create", "iam_user", ResourceScopeOrgUnit).InOrgUnit(unitId()),
			evalCtx:  EvalContext{OrgUnitId: idOf("OU_OTHER")},
			expected: false,
		},
		{
			name:     "a parent unit grant does not reach a child unit",
			grants:   []string{"create:iam_user:orgunit/OU_PARENT"},
			required: PermFor("create", "iam_user", ResourceScopeOrgUnit).InOrgUnit(idOf("OU_CHILD")),
			expected: false,
		},
		{
			name:     "private grant answers for the caller's own record",
			grants:   []string{"view:iam_user:private"},
			required: PermFor("view", "iam_user", ResourceScopePrivate).OwnedByCaller(true),
			expected: true,
		},
		{
			name:     "private grant denies someone else's record",
			grants:   []string{"view:iam_user:private"},
			required: PermFor("view", "iam_user", ResourceScopePrivate).OwnedByCaller(false),
			expected: false,
		},
		{
			name:     "a narrower grant does not satisfy a wider requirement",
			grants:   []string{"create:iam_user:org/ORG1"},
			required: PermFor("create", "iam_user", ResourceScopeDomain),
			expected: false,
		},
		{
			name:     "the wrong action denies",
			grants:   []string{"view:iam_user:domain"},
			required: PermFor("delete", "iam_user", ResourceScopeDomain),
			expected: false,
		},
		{
			name:     "the wrong resource denies",
			grants:   []string{"create:iam_group:domain"},
			required: PermFor("create", "iam_user", ResourceScopeDomain),
			expected: false,
		},
	}
}

// Executor 1: the in-memory guard the middleware runs.
func guardAllows(scenario parityScenario) bool {
	ctx := corectx.NewRequestContext(context.Background())
	grants := ds.NewSet[string]()
	grants.AddMany(scenario.grants...)
	orgs := ds.NewSet[model.Id]()
	orgs.AddMany(scenario.evalCtx.UserOrgIds...)

	ctx.SetPermissions(corectx.ContextPermissions{
		IsOwner:      scenario.isOwner,
		Entitlements: grants,
		UserOrgIds:   orgs,
		OrgUnitId:    scenario.evalCtx.OrgUnitId,
		OrgUnitOrgId: scenario.evalCtx.OrgUnitOrgId,
	})
	return AssertPermission(ctx, scenario.required) == nil
}

// Executor 2: the candidate set the SQL matcher turns into its IN-list, and the
// permission probe range-scans with. Set intersection here is exactly what the
// `ent_expression IN (...)` predicate does in the database.
func candidateSetAllows(scenario parityScenario) bool {
	if scenario.isOwner {
		return true
	}
	if scenario.required.Scope == ResourceScopePrivate && !scenario.required.IsRecordOwnedByCaller {
		return false
	}
	held := make(map[string]struct{}, len(scenario.grants))
	for _, grant := range scenario.grants {
		held[grant] = struct{}{}
	}
	for _, candidate := range CandidateExpressions(scenario.required, scenario.evalCtx) {
		if _, ok := held[candidate]; ok {
			return true
		}
	}
	return false
}

func TestEvaluationParity(t *testing.T) {
	for _, scenario := range parityScenarios() {
		t.Run(scenario.name, func(t *testing.T) {
			byGuard := guardAllows(scenario)
			byCandidates := candidateSetAllows(scenario)

			assert.Equal(t, scenario.expected, byGuard, "in-memory guard disagreed with the expected answer")
			assert.Equal(t, scenario.expected, byCandidates, "candidate set disagreed with the expected answer")
			assert.Equal(t, byGuard, byCandidates, "the two executors disagreed with each other")
		})
	}
}
