package permission

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type PermissionDomainService interface {
	// IsAuthorized checks if a user is authorized to perform an action on a resource within a scope.
	// Use this function to quickly get the answer without reasons.
	IsAuthorized(ctx corectx.Context, query IsAuthorizedQuery) (*IsAuthorizedResult, error)
	ListAllUserPermissions(ctx corectx.Context, query ListAllUserPermissionsQuery) (*ListAllUserPermissionsResult, error)
}

type PermissionAppService interface {
	IsAuthorized(ctx corectx.Context, query IsAuthorizedQuery) (*IsAuthorizedResult, error)
	GetUserEntitlements(ctx corectx.Context, query GetUserEntitlementsQuery) (*GetUserEntitlementsResult, error)
}
