package constants

import (
	reguard "github.com/sky-as-code/nikki-erp/modules/core/requestguard"
)

const SalesModuleName = "sales"

// Authorization scopes, re-exported so Sales code need not import requestguard everywhere.
type ResourceScope = reguard.ResourceScope

const (
	ResourceScopeDomain  = reguard.ResourceScopeDomain
	ResourceScopeOrg     = reguard.ResourceScopeOrg
	ResourceScopeOrgUnit = reguard.ResourceScopeOrgUnit
)

// SalesRouteV1 is the REST route group every Sales resource engine hangs off. It must match the
// schema prefix: the engine derives the path segment and the IAM resource code from the schema
// name, so the URL, the schema and the resource code are one string in three places.
const SalesRouteV1 = "/v1/sales"
