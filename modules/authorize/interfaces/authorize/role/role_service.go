package role

import "context"

type RoleService interface {
	CreateRole(ctx context.Context, cmd CreateRoleCommand) (*CreateRoleResult, error)
	UpdateRole(ctx context.Context, cmd UpdateRoleCommand) (*UpdateRoleResult, error)
	DeleteRoleHard(ctx context.Context, cmd DeleteRoleHardCommand) (*DeleteRoleHardResult, error)
	GetRoleById(ctx context.Context, query GetRoleByIdQuery) (*GetRoleByIdResult, error)
	SearchRoles(ctx context.Context, query SearchRolesQuery) (*SearchRolesResult, error)
	GetRolesBySubject(ctx context.Context, query GetRolesBySubjectQuery) (*GetRolesBySubjectResult, error)
}
