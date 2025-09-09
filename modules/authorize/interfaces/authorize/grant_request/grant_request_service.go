package grant_request

import (
	"context"
)

type GrantRequestService interface {
	CreateGrantRequest(ctx context.Context, cmd CreateGrantRequestCommand) (*CreateGrantRequestResult, error)
	RespondToGrantRequest(ctx context.Context, cmd RespondToGrantRequestCommand) (*RespondToGrantRequestResult, error)
}
