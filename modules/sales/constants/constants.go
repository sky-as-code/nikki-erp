package constants

import (
	reguard "github.com/sky-as-code/nikki-erp/modules/core/requestguard"
)

const SalesModuleName = "sales"

// The authorization scopes, re-exported so that Sales code names them without importing
// requestguard everywhere.
type ResourceScope = reguard.ResourceScope

const (
	ResourceScopeDomain  = reguard.ResourceScopeDomain
	ResourceScopeOrg     = reguard.ResourceScopeOrg
	ResourceScopeOrgUnit = reguard.ResourceScopeOrgUnit
)

// SalesRouteV1 is the REST route group every Sales resource engine hangs off.
//
// It matches the schema prefix rather than merely resembling it: the engine derives a resource's
// path segment from its schema name, so "/v1/sales" plus schema "sales_channels" is the URL a
// client calls and "sales_channels" is the IAM resource code asserted against it. Those three
// strings are one string in three places.
const SalesRouteV1 = "/v1/sales"
