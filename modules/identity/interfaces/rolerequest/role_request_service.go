package role_request

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type RoleRequestService interface {
	CreateRoleRequest(ctx corectx.Context, cmd CreateRoleRequestCommand) (*CreateRoleRequestResult, error)
	DeleteRoleRequest(ctx corectx.Context, cmd DeleteRoleRequestCommand) (*DeleteRoleRequestResult, error)
	GetRoleRequest(ctx corectx.Context, query GetRoleRequestQuery) (*GetRoleRequestResult, error)
	RoleRequestExists(ctx corectx.Context, query RoleRequestExistsQuery) (*RoleRequestExistsResult, error)
	SearchRoleRequests(ctx corectx.Context, query SearchRoleRequestsQuery) (*SearchRoleRequestsResult, error)
	UpdateRoleRequest(ctx corectx.Context, cmd UpdateRoleRequestCommand) (*UpdateRoleRequestResult, error)
}
