package permission

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type PermissionService interface {
	// IsAuthorized checks if a user is authorized to perform an action on a resource within a scope.
	// Use this function to quickly get the answer without reasons.
	IsAuthorized(ctx corectx.Context, query IsAuthorizedQuery) (*IsAuthorizedResult, error)

	// CheckPermissions checks if a user is authorized to perform an action on a resource within a scope and returns the reasons.
	// Use this function to investigate the authorization decisions.
	CheckPermissions(ctx corectx.Context, query CheckPermissionsQuery) (*CheckPermissionsResult, error)
}
