package revoke_request

import "github.com/sky-as-code/nikki-erp/modules/core/crud"

type RevokeRequestService interface {
	Create(ctx crud.Context, cmd CreateRevokeRequestCommand) (*CreateRevokeRequestResult, error)
	TargetIsDeleted(ctx crud.Context, cmd TargetIsDeletedCommand) (*TargetIsDeletedResult, error)
}
