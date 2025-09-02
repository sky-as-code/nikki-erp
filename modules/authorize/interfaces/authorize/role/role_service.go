package role

import "github.com/sky-as-code/nikki-erp/modules/core/crud"

type RoleService interface {
	CreateRole(ctx crud.Context, cmd CreateRoleCommand) (*CreateRoleResult, error)
	UpdateRole(ctx crud.Context, cmd UpdateRoleCommand) (*UpdateRoleResult, error)
	DeleteRoleHard(ctx crud.Context, cmd DeleteRoleHardCommand) (*DeleteRoleHardResult, error)
	GetRoleById(ctx crud.Context, query GetRoleByIdQuery) (*GetRoleByIdResult, error)
	SearchRoles(ctx crud.Context, query SearchRolesQuery) (*SearchRolesResult, error)
	GetRolesBySubject(ctx crud.Context, query GetRolesBySubjectQuery) (*GetRolesBySubjectResult, error)
}
