package domain

// AuthzScopeWidth maps authorization scope strings to a numeric width where a larger value is a
// broader grant (domain widest, private narrowest). Used to compare resource min/max bounds with
// entitlement scope without duplicating ordering rules.
func AuthzScopeWidth(scope ResourceScope) int {
	switch scope {
	case ResourceScopeDomain:
		return 4
	case ResourceScopeOrg:
		return 3
	case ResourceScopeOrgUnit:
		return 2
	case ResourceScopePrivate:
		return 1
	default:
		return 0
	}
}

func IsResourceScopeInBounds(minScope, maxScope, scope ResourceScope) bool {
	wMin := AuthzScopeWidth(minScope)
	wMax := AuthzScopeWidth(maxScope)
	wThis := AuthzScopeWidth(scope)
	return wMin <= wThis && wThis <= wMax
}
