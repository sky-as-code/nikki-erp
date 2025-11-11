package grant_request

import (
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type GrantRequestService interface {
	CreateGrantRequest(ctx crud.Context, cmd CreateGrantRequestCommand) (*CreateGrantRequestResult, error)
	CancelGrantRequest(ctx crud.Context, cmd CancelGrantRequestCommand) (*CancelGrantRequestResult, error)
	DeleteGrantRequest(ctx crud.Context, cmd DeleteGrantRequestCommand) (*DeleteGrantRequestResult, error)
	GetGrantRequestById(ctx crud.Context, query GetGrantRequestByIdQuery) (*GetGrantRequestByIdResult, error)
	SearchGrantRequests(ctx crud.Context, query SearchGrantRequestsQuery) (*SearchGrantRequestsResult, error)
	RespondToGrantRequest(ctx crud.Context, cmd RespondToGrantRequestCommand) (*RespondToGrantRequestResult, error)
	TargetIsDeleted(ctx crud.Context, cmd TargetIsDeletedCommand) (*TargetIsDeletedResult, error)
}
