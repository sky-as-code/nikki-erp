package role_request

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	domain "github.com/sky-as-code/nikki-erp/modules/identity/domain/models"
)

type RoleRequestDomainService interface {
	CreateRoleRequest(ctx corectx.Context, cmd CreateRoleRequestCommand, opts ...corecrud.ServiceCreateOptions[*domain.RoleRequest]) (*CreateRoleRequestResult, error)
	DeleteRoleRequest(ctx corectx.Context, cmd DeleteRoleRequestCommand, opts ...corecrud.ServiceDeleteOptions) (*DeleteRoleRequestResult, error)
	GetRoleRequest(ctx corectx.Context, query GetRoleRequestQuery) (*dyn.OpResult[domain.RoleRequest], error)
	RoleRequestExists(ctx corectx.Context, query RoleRequestExistsQuery) (*RoleRequestExistsResult, error)
	SearchRoleRequests(ctx corectx.Context, query SearchRoleRequestsQuery, opts ...corecrud.ServiceSearchOptions) (*SearchRoleRequestsResult, error)
	UpdateRoleRequest(ctx corectx.Context, cmd UpdateRoleRequestCommand, opts ...corecrud.ServiceUpdateOptions[*domain.RoleRequest]) (*UpdateRoleRequestResult, error)
}

type RoleRequestAppService interface {
	CreateRoleRequest(ctx corectx.Context, cmd CreateRoleRequestCommand) (*CreateRoleRequestResult, error)
	DeleteRoleRequest(ctx corectx.Context, cmd DeleteRoleRequestCommand) (*DeleteRoleRequestResult, error)
	GetRoleRequest(ctx corectx.Context, query GetRoleRequestQuery) (*GetRoleRequestResult, error)
	RoleRequestExists(ctx corectx.Context, query RoleRequestExistsQuery) (*RoleRequestExistsResult, error)
	SearchRoleRequests(ctx corectx.Context, query SearchRoleRequestsQuery) (*SearchRoleRequestsResult, error)
	UpdateRoleRequest(ctx corectx.Context, cmd UpdateRoleRequestCommand) (*UpdateRoleRequestResult, error)
}
