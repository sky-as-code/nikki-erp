package models

import c "github.com/sky-as-code/nikki-erp/modules/identity/constants"

// AuthzScopeWidth maps authorization scope strings to a numeric width where a larger value is a
// broader grant (domain widest, private narrowest). Used to compare resource min/max bounds with
// entitlement scope without duplicating ordering rules.
func AuthzScopeWidth(scope c.ResourceScope) int {
	switch scope {
	case c.ResourceScopeDomain:
		return 4
	case c.ResourceScopeOrg:
		return 3
	case c.ResourceScopeOrgUnit:
		return 2
	case c.ResourceScopePrivate:
		return 1
	default:
		return 0
	}
}

func IsResourceScopeInBounds(minScope, maxScope, scope c.ResourceScope) bool {
	wMin := AuthzScopeWidth(minScope)
	wMax := AuthzScopeWidth(maxScope)
	wThis := AuthzScopeWidth(scope)
	return wMin <= wThis && wThis <= wMax
}
