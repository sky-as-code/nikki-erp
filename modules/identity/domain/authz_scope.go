package domain

// AuthzScopeWidth maps authorization scope strings to a numeric width where a larger value is a
// broader grant (domain widest, private narrowest). Used to compare resource min/max bounds with
// entitlement scope without duplicating ordering rules.
func AuthzScopeWidth(scope ResourceScope) (int, bool) {
	switch scope {
	case ResourceScopeDomain:
		return 4, true
	case ResourceScopeOrg:
		return 3, true
	case ResourceScopeOrgUnit:
		return 2, true
	case ResourceScopePrivate:
		return 1, true
	default:
		return 0, false
	}
}

func IsResourceScopeInBounds(minScope, maxScope, scope ResourceScope) bool {
	wMin, okMin := AuthzScopeWidth(minScope)
	wMax, okMax := AuthzScopeWidth(maxScope)
	wThis, okThis := AuthzScopeWidth(scope)
	if !okMin || !okMax || !okThis {
		return false
	}
	return wMin <= wThis && wThis <= wMax
}
