package resource

import (
	"context"
)

type ActionService interface {
	CreateAction(ctx context.Context, cmd CreateActionCommand) (*CreateActionResult, error)
	UpdateAction(ctx context.Context, cmd UpdateActionCommand) (*UpdateActionResult, error)
	DeleteActionHard(ctx context.Context, cmd DeleteActionHardByIdQuery) (*DeleteActionHardByIdResult, error)
	GetActionById(ctx context.Context, query GetActionByIdQuery) (*GetActionByIdResult, error)
	SearchActions(ctx context.Context, query SearchActionsCommand) (*SearchActionsResult, error)
}
