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
	ResourceAuthorizationResource     = "authz_resource"
	ResourceAuthorizationRole         = "authz_role"
	ResourceAuthorizationGrantRequest = "authz_grant_request"
	ResourceAuthorizationEntitlement  = "authz_entitlement"
	ResourceIdentityUser              = "identity_user"
	ResourceIdentityGroup             = "identity_group"
	ResourceIdentityOrganization      = "identity_org"
	ResourceIdentityOrgUnit           = "identity_orgunit"
)
