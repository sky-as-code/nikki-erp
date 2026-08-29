package constants

import (
	reguard "github.com/sky-as-code/nikki-erp/modules/core/requestguard"
)

const AccountingModuleName = "accounting"

// Authorization scopes, re-exported so Accounting code avoids importing requestguard everywhere.
type ResourceScope = reguard.ResourceScope

const (
	ResourceScopeDomain  = reguard.ResourceScopeDomain
	ResourceScopeOrg     = reguard.ResourceScopeOrg
	ResourceScopeOrgUnit = reguard.ResourceScopeOrgUnit
)

// AccountingRouteV1 is the REST route group every Accounting resource engine hangs off. It must
// match the schema prefix: the engine derives a resource's path segment from its schema name, so
// the URL, the schema name and the IAM resource code are one string in three places.
const AccountingRouteV1 = "/v1/accounting"
