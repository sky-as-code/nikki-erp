package grant_request

import (
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type GrantRequestService interface {
	CreateGrantRequest(ctx crud.Context, cmd CreateGrantRequestCommand) (*CreateGrantRequestResult, error)
	CancelGrantRequest(ctx crud.Context, cmd CancelGrantRequestCommand) (*CancelGrantRequestResult, error)
	RespondToGrantRequest(ctx crud.Context, cmd RespondToGrantRequestCommand) (*RespondToGrantRequestResult, error)
}
