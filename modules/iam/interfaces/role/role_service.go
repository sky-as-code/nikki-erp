package role

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
)

type RoleDomainService interface {
	CreateRole(ctx corectx.Context, cmd CreateRoleCommand, opts ...corecrud.ServiceCreateOptions[*domain.Role]) (*CreateRoleResult, error)
	CreatePrivateRole(ctx corectx.Context, cmd CreatePrivateRoleCommand, opts ...corecrud.ServiceCreateOptions[*domain.Role]) (*CreateRoleResult, error)
	DeleteRole(ctx corectx.Context, cmd DeleteRoleCommand, opts ...corecrud.ServiceDeleteOptions) (*DeleteRoleResult, error)
	DeletePrivateRole(ctx corectx.Context, cmd DeletePrivateRoleCommand, opts ...corecrud.ServiceDeleteOptions) (*DeleteRoleResult, error)
	GetRole(ctx corectx.Context, query GetRoleQuery) (*dyn.OpResult[domain.Role], error)
	ManageRoleEntitlements(ctx corectx.Context, cmd ManageRoleEntitlementsCommand) (
		*ManageRoleEntitlementsResult, error,
	)
	RoleExists(ctx corectx.Context, query RoleExistsQuery) (*RoleExistsResult, error)
	SearchRoles(ctx corectx.Context, query SearchRolesQuery, opts ...corecrud.ServiceSearchOptions) (*SearchRolesResult, error)
	SetRoleIsArchived(ctx corectx.Context, cmd SetRoleIsArchivedCommand) (*SetRoleIsArchivedResult, error)
	UpdateRole(ctx corectx.Context, cmd UpdateRoleCommand, opts ...corecrud.ServiceUpdateOptions[*domain.Role]) (*UpdateRoleResult, error)
}

type RoleAppService interface {
	CreateRole(ctx corectx.Context, cmd CreateRoleCommand) (*CreateRoleResult, error)
	DeleteRole(ctx corectx.Context, cmd DeleteRoleCommand) (*DeleteRoleResult, error)
	GetRole(ctx corectx.Context, query GetRoleQuery) (*GetRoleResult, error)
	ManageRoleEntitlements(ctx corectx.Context, cmd ManageRoleEntitlementsCommand) (
		*ManageRoleEntitlementsResult, error,
	)
	RoleExists(ctx corectx.Context, query RoleExistsQuery) (*RoleExistsResult, error)
	SearchRoles(ctx corectx.Context, query SearchRolesQuery) (*SearchRolesResult, error)
	SetRoleIsArchived(ctx corectx.Context, cmd SetRoleIsArchivedCommand) (*SetRoleIsArchivedResult, error)
	UpdateRole(ctx corectx.Context, cmd UpdateRoleCommand) (*UpdateRoleResult, error)
}
