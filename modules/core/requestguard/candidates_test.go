package requestguard

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sky-as-code/nikki-erp/common/model"
)

func permFor(scope ResourceScope) Perm {
	return Perm{ActionCode: "create", ResourceCode: "iam_user", Scope: scope}
}

// The candidate set is the whole evaluation semantics. These cases enumerate the
// matrix the plan codifies: 4 scopes x wildcard combinations x membership.
func TestCandidateExpressions_Matrix(t *testing.T) {
	orgId := model.Id("ORG1")
	otherOrgId := model.Id("ORG2")
	unitId := model.Id("OU1")

	tests := []struct {
		name        string
		required    Perm
		evalCtx     EvalContext
		mustContain []string
		mustNotHave []string
	}{
		{
			name:     "domain scope accepts only omnipotent and domain grants",
			required: permFor(ResourceScopeDomain),
			mustContain: []string{
				"*:*:*",
				"create:iam_user:domain", "*:iam_user:domain", "create:*:domain", "*:*:domain",
			},
			mustNotHave: []string{"create:iam_user:org", "create:iam_user:orgunit"},
		},
		{
			name:     "org scope with id, caller not a member",
			required: Perm{ActionCode: "create", ResourceCode: "iam_user", Scope: ResourceScopeOrg, OrgId: &orgId},
			mustContain: []string{
				"*:*:*",
				"create:iam_user:domain",
				"create:iam_user:org/ORG1", "*:iam_user:org/ORG1", "create:*:org/ORG1", "*:*:org/ORG1",
			},
			// A bare org grant must NOT answer for a caller who is not in that org.
			mustNotHave: []string{"create:iam_user:org", "*:*:org"},
		},
		{
			name:     "org scope with id, caller is a member",
			required: Perm{ActionCode: "create", ResourceCode: "iam_user", Scope: ResourceScopeOrg, OrgId: &orgId},
			evalCtx:  EvalContext{UserOrgIds: []model.Id{orgId}},
			mustContain: []string{
				"create:iam_user:org/ORG1",
				"create:iam_user:org", "*:iam_user:org", "create:*:org", "*:*:org",
			},
		},
		{
			name:     "org scope, membership in a different org does not help",
			required: Perm{ActionCode: "create", ResourceCode: "iam_user", Scope: ResourceScopeOrg, OrgId: &orgId},
			evalCtx:  EvalContext{UserOrgIds: []model.Id{otherOrgId}},
			mustContain: []string{"create:iam_user:org/ORG1"},
			mustNotHave: []string{"create:iam_user:org", "create:iam_user:org/ORG2"},
		},
		{
			name: "orgunit scope with id, caller belongs to the unit",
			required: Perm{
				ActionCode: "create", ResourceCode: "iam_user",
				Scope: ResourceScopeOrgUnit, OrgUnitId: &unitId, OrgId: &orgId,
			},
			evalCtx: EvalContext{OrgUnitId: &unitId, UserOrgIds: []model.Id{orgId}},
			mustContain: []string{
				"create:iam_user:orgunit/OU1", "*:*:orgunit/OU1",
				"create:iam_user:orgunit", // bare unit grant, caller is in the unit
				"create:iam_user:domain",
				"create:iam_user:org/ORG1", // fallback to the unit's org
				"create:iam_user:org",      // caller is a member of that org
			},
		},
		{
			name: "orgunit scope, caller in a different unit gets no bare unit grant",
			required: Perm{
				ActionCode: "create", ResourceCode: "iam_user",
				Scope: ResourceScopeOrgUnit, OrgUnitId: &unitId,
			},
			evalCtx:     EvalContext{OrgUnitId: idOf("OU_OTHER")},
			mustContain: []string{"create:iam_user:orgunit/OU1"},
			mustNotHave: []string{"create:iam_user:orgunit", "*:*:orgunit"},
		},
		{
			name:        "private scope",
			required:    permFor(ResourceScopePrivate),
			mustContain: []string{"*:*:*", "create:iam_user:domain", "create:iam_user:private", "*:*:private"},
			mustNotHave: []string{"create:iam_user:org", "create:iam_user:orgunit"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			candidates := CandidateExpressions(test.required, test.evalCtx)
			for _, expected := range test.mustContain {
				assert.Contains(t, candidates, expected)
			}
			for _, forbidden := range test.mustNotHave {
				assert.NotContains(t, candidates, forbidden)
			}
		})
	}
}

// A parent unit's grant must not reach a child unit: entitlements on a unit apply
// to that unit alone, which is what makes them auditable when the tree is
// reorganised.
func TestCandidateExpressions_NoOrgUnitInheritance(t *testing.T) {
	childId := model.Id("OU_CHILD")
	required := Perm{
		ActionCode: "create", ResourceCode: "iam_user",
		Scope: ResourceScopeOrgUnit, OrgUnitId: &childId,
	}
	candidates := CandidateExpressions(required, EvalContext{OrgUnitId: idOf("OU_PARENT")})

	assert.NotContains(t, candidates, "create:iam_user:orgunit/OU_PARENT")
	assert.NotContains(t, candidates, "create:iam_user:orgunit")
}

func TestCandidateExpressions_NoDuplicates(t *testing.T) {
	orgId := model.Id("ORG1")
	unitId := model.Id("OU1")
	candidates := CandidateExpressions(
		Perm{ActionCode: "create", ResourceCode: "iam_user", Scope: ResourceScopeOrgUnit, OrgUnitId: &unitId, OrgId: &orgId},
		EvalContext{OrgUnitId: &unitId, UserOrgIds: []model.Id{orgId}},
	)

	seen := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		_, duplicate := seen[candidate]
		assert.False(t, duplicate, "duplicate candidate %q", candidate)
		seen[candidate] = struct{}{}
	}
}

// Every candidate must itself be a well-formed expression - the SQL matcher puts
// them straight into an IN-list and the cache stores exactly this shape.
func TestCandidateExpressions_AllParseable(t *testing.T) {
	orgId := model.Id("ORG1")
	unitId := model.Id("OU1")
	scopes := []ResourceScope{ResourceScopeDomain, ResourceScopeOrg, ResourceScopeOrgUnit, ResourceScopePrivate}

	for _, scope := range scopes {
		required := Perm{
			ActionCode: "create", ResourceCode: "iam_user", Scope: scope,
			OrgId: &orgId, OrgUnitId: &unitId,
		}
		for _, candidate := range CandidateExpressions(required, EvalContext{}) {
			parsed, err := ParseExpression(candidate)
			assert.NoError(t, err, "candidate %q must parse", candidate)
			if err == nil {
				assert.Equal(t, candidate, parsed.String())
			}
		}
	}
}
