package requestguard

import (
	"github.com/sky-as-code/nikki-erp/common/model"
)

// EvalContext is what the evaluator knows about the *caller*, as opposed to Perm
// which describes the *record* being reached for. Both sides are needed: a bare
// `org` grant only answers when the caller is a member of the record's org.
type EvalContext struct {
	// Orgs the caller belongs to.
	UserOrgIds []model.Id
	// The org unit the caller belongs to, if any.
	OrgUnitId *model.Id
	// The org that OrgUnitId belongs to. Needed for the orgunit -> org fallback.
	OrgUnitOrgId *model.Id
}

func (this EvalContext) belongsToOrg(orgId *model.Id) bool {
	if orgId == nil {
		return false
	}
	for _, id := range this.UserOrgIds {
		if id == *orgId {
			return true
		}
	}
	return false
}

func (this EvalContext) belongsToOrgUnit(orgUnitId *model.Id) bool {
	return orgUnitId != nil && this.OrgUnitId != nil && *this.OrgUnitId == *orgUnitId
}

// CandidateExpressions returns every stored expression that would satisfy the
// required permission for this caller. Holding ANY of them means "allowed".
//
// This is the single definition of the evaluation semantics. The in-memory guard
// tests set membership against it; the SQL matcher turns it into an IN-list; the
// permission probe range-scans the cache with it. Because all three consume the
// same slice they cannot disagree - which is the whole point, since a permission
// system whose "can I?" and "may I?" answers differ is a security defect, not an
// inconsistency.
//
// Scope widening runs domain > org > orgunit: a wider grant satisfies a narrower
// requirement. There is deliberately NO inheritance between org units - a grant on
// a parent unit does not reach its children, which is the documented entitlement
// semantics and what makes a unit grant auditable.
func CandidateExpressions(required Perm, evalCtx EvalContext) []string {
	candidates := make([]string, 0, 24)
	add := func(exprs ...string) {
		candidates = append(candidates, exprs...)
	}

	// Rank 0: the omnipotent grant answers every question.
	add(OmnipotentExpression())

	// Rank 1: domain grants answer every scope.
	add(scopeVariants(required, ResourceScopeDomain, nil)...)

	switch required.Scope {
	case ResourceScopeDomain:
		// Already covered by the domain variants above.

	case ResourceScopeOrg:
		add(orgCandidates(required, evalCtx, required.OrgId)...)

	case ResourceScopeOrgUnit:
		// Exact unit grants, and bare unit grants when the caller belongs to that unit.
		add(scopeVariants(required, ResourceScopeOrgUnit, required.OrgUnitId)...)
		if evalCtx.belongsToOrgUnit(required.OrgUnitId) {
			add(scopeVariants(required, ResourceScopeOrgUnit, nil)...)
		}
		// Fallback: an org-level grant for the org that owns the unit.
		add(orgCandidates(required, evalCtx, required.OrgId)...)

	case ResourceScopePrivate:
		// Private grants are scoped to the record's owner. Ownership itself is the
		// caller's to assert (see AssertPermission); reaching here means the record
		// is the caller's own, so a bare private grant answers.
		add(scopeVariants(required, ResourceScopePrivate, nil)...)
	}

	return dedupe(candidates)
}

// orgCandidates yields the org-scoped grants that answer for orgId: the ones
// naming it explicitly, plus the bare org grants when the caller is a member.
func orgCandidates(required Perm, evalCtx EvalContext, orgId *model.Id) []string {
	if orgId == nil {
		return nil
	}
	result := scopeVariants(required, ResourceScopeOrg, orgId)
	if evalCtx.belongsToOrg(orgId) {
		result = append(result, scopeVariants(required, ResourceScopeOrg, nil)...)
	}
	return result
}

// scopeVariants returns the four wildcard combinations at one scope, widest last
// so that a reader of the slice sees exact-first ordering.
func scopeVariants(required Perm, scope ResourceScope, scopeId *model.Id) []string {
	return []string{
		BuildExpression(required.ActionCode, required.ResourceCode, scope, scopeId),
		BuildExpression(Wildcard, required.ResourceCode, scope, scopeId),
		BuildExpression(required.ActionCode, Wildcard, scope, scopeId),
		BuildExpression(Wildcard, Wildcard, scope, scopeId),
	}
}

func dedupe(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		if _, exists := seen[item]; exists {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}
