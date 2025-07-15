package role

import "context"

type RoleService interface {
	CreateRole(ctx context.Context, cmd CreateRoleCommand) (*CreateRoleResult, error)
	GetRoleById(ctx context.Context, query GetRoleByIdQuery) (*GetRoleByIdResult, error)
	SearchRoles(ctx context.Context, query SearchRolesQuery) (*SearchRolesResult, error)
}
