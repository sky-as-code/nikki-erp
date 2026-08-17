package constants

import (
	reguard "github.com/sky-as-code/nikki-erp/modules/core/requestguard"
)

const IamModuleName = "iam"

type ResourceScope = reguard.ResourceScope

const (
	ResourceScopeDomain  = reguard.ResourceScopeDomain
	ResourceScopeOrg     = reguard.ResourceScopeOrg
	ResourceScopeOrgUnit = reguard.ResourceScopeOrgUnit
	ResourceScopePrivate = reguard.ResourceScopePrivate
)

const (
	ResourceIamResource     = "iam_resource"
	ResourceIamRole         = "iam_role"
	ResourceIamGrantRequest = "iam_grant_request"
	ResourceIamEntitlement  = "iam_entitlement"
	ResourceIamUser         = "iam_user"
	ResourceIamGroup        = "iam_group"
	ResourceIamOrganization = "iam_org"
	ResourceIamOrgUnit      = "iam_orgunit"
)
