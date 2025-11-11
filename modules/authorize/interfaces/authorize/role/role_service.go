package role

import "github.com/sky-as-code/nikki-erp/modules/core/crud"

type RoleService interface {
	AddEntitlements(ctx crud.Context, cmd AddEntitlementsCommand) (*AddEntitlementsResult, error)
	CreateRole(ctx crud.Context, cmd CreateRoleCommand) (*CreateRoleResult, error)
	DeleteRoleHard(ctx crud.Context, cmd DeleteRoleHardCommand) (*DeleteRoleHardResult, error)
	GetRoleById(ctx crud.Context, query GetRoleByIdQuery) (*GetRoleByIdResult, error)
	GetRolesBySubject(ctx crud.Context, query GetRolesBySubjectQuery) (*GetRolesBySubjectResult, error)
	RemoveEntitlements(ctx crud.Context, cmd RemoveEntitlementsCommand) (*RemoveEntitlementsResult, error)
	SearchRoles(ctx crud.Context, query SearchRolesQuery) (*SearchRolesResult, error)
	UpdateRole(ctx crud.Context, cmd UpdateRoleCommand) (*UpdateRoleResult, error)
}
