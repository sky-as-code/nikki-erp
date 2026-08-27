package constants

import (
	reguard "github.com/sky-as-code/nikki-erp/modules/core/requestguard"
)

const AccountingModuleName = "accounting"

// The authorization scopes, re-exported so that Accounting code names them without importing
// requestguard everywhere.
type ResourceScope = reguard.ResourceScope

const (
	ResourceScopeDomain  = reguard.ResourceScopeDomain
	ResourceScopeOrg     = reguard.ResourceScopeOrg
	ResourceScopeOrgUnit = reguard.ResourceScopeOrgUnit
)

// AccountingRouteV1 is the REST route group every Accounting resource engine hangs off.
//
// It matches the schema prefix rather than merely resembling it: the engine derives a resource's
// path segment from its schema name, so "/v1/accounting" plus schema "accounting_tax" is the URL a
// client calls and "accounting_tax" is the IAM resource code asserted against it. Those three
// strings are one string in three places.
const AccountingRouteV1 = "/v1/accounting"
