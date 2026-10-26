package constants

import (
	commconstants "github.com/sky-as-code/nikki-erp/modules/commerce/constants"
	reguard "github.com/sky-as-code/nikki-erp/modules/core/requestguard"
)

// SalesModuleName names the Sales part of Commerce: its schema prefix and translation namespace.
// The module itself, and so the settings module key, is commerce/constants.CommerceModuleName.
const SalesModuleName = "sales"

// SettingsModuleKey is the module key Sales' settings are registered, read and written under. A
// setting belongs to a module, and Sales is part of Commerce.
const SettingsModuleKey = commconstants.CommerceModuleName

// Authorization scopes, re-exported so Sales code need not import requestguard everywhere.
type ResourceScope = reguard.ResourceScope

const (
	ResourceScopeTenant  = reguard.ResourceScopeTenant
	ResourceScopeOrg     = reguard.ResourceScopeOrg
	ResourceScopeOrgUnit = reguard.ResourceScopeOrgUnit
)

// SalesRouteV1 is the REST route group every Sales resource engine hangs off. It must match the
// schema prefix: the engine derives the path segment and the IAM resource code from the schema
// name, so the URL, the schema and the resource code are one string in three places.
const SalesRouteV1 = "/v1/sales"
