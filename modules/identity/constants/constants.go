package constants

import (
	reguard "github.com/sky-as-code/nikki-erp/modules/core/requestguard"
)

const IdentityModuleName = "identity"

type ResourceScope = reguard.ResourceScope

const (
	ResourceScopeDomain  = reguard.ResourceScopeDomain
	ResourceScopeOrg     = reguard.ResourceScopeOrg
	ResourceScopeOrgUnit = reguard.ResourceScopeOrgUnit
	ResourceScopePrivate = reguard.ResourceScopePrivate
)

const (
	ResourceAuthorizationResource     = "authz.resource"
	ResourceAuthorizationRole         = "authz.role"
	ResourceAuthorizationGrantRequest = "authz.grant_request"
	ResourceAuthorizationEntitlement  = "authz.entitlement"
	ResourceIdentityUser              = "identity.user"
	ResourceIdentityGroup             = "identity.group"
	ResourceIdentityOrganization      = "identity.org"
	ResourceIdentityOrgUnit           = "identity.orgunit"
)
