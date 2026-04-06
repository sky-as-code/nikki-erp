package role

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type RoleService interface {
	CreateRole(ctx corectx.Context, cmd CreateRoleCommand) (*CreateRoleResult, error)
	DeleteRole(ctx corectx.Context, cmd DeleteRoleCommand) (*DeleteRoleResult, error)
	DeletePrivateRole(ctx corectx.Context, cmd DeletePrivateRoleCommand) (*DeleteRoleResult, error)
	GetRole(ctx corectx.Context, query GetRoleQuery) (*GetRoleResult, error)
	ManageRoleEntitlements(ctx corectx.Context, cmd ManageRoleEntitlementsCommand) (
		*ManageRoleEntitlementsResult, error,
	)
	RoleExists(ctx corectx.Context, query RoleExistsQuery) (*RoleExistsResult, error)
	SearchRoles(ctx corectx.Context, query SearchRolesQuery) (*SearchRolesResult, error)
	SetRoleIsArchived(ctx corectx.Context, cmd SetRoleIsArchivedCommand) (*SetRoleIsArchivedResult, error)
	UpdateRole(ctx corectx.Context, cmd UpdateRoleCommand) (*UpdateRoleResult, error)
}
