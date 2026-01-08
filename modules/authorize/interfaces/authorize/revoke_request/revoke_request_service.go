package revoke_request

import "github.com/sky-as-code/nikki-erp/modules/core/crud"

type RevokeRequestService interface {
	Create(ctx crud.Context, cmd CreateRevokeRequestCommand) (*CreateRevokeRequestResult, error)
	CreateBulk(ctx crud.Context, cmd CreateBulkRevokeRequestsCommand) (*CreateBulkRevokeRequestsResult, error)
	GetById(ctx crud.Context, query GetRevokeRequestByIdQuery) (*GetRevokeRequestByIdResult, error)
	Search(ctx crud.Context, query SearchRevokeRequestsQuery) (*SearchRevokeRequestsResult, error)
	Delete(ctx crud.Context, cmd DeleteRevokeRequestCommand) (*DeleteRevokeRequestResult, error)
	TargetIsDeleted(ctx crud.Context, cmd TargetIsDeletedCommand) (*TargetIsDeletedResult, error)
}
