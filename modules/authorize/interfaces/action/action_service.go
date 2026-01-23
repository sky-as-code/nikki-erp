package action

import "github.com/sky-as-code/nikki-erp/modules/core/crud"

type ActionService interface {
	CreateAction(ctx crud.Context, cmd CreateActionCommand) (*CreateActionResult, error)
	UpdateAction(ctx crud.Context, cmd UpdateActionCommand) (*UpdateActionResult, error)
	DeleteActionHard(ctx crud.Context, cmd DeleteActionHardByIdCommand) (*DeleteActionHardByIdResult, error)
	GetActionById(ctx crud.Context, query GetActionByIdQuery) (*GetActionByIdResult, error)
	SearchActions(ctx crud.Context, query SearchActionsQuery) (*SearchActionsResult, error)
}
