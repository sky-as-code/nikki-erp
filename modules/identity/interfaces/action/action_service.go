package action

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type ActionService interface {
	CreateAction(ctx corectx.Context, cmd CreateActionCommand) (*CreateActionResult, error)
	DeleteAction(ctx corectx.Context, cmd DeleteActionCommand) (*DeleteActionResult, error)
	ActionExists(ctx corectx.Context, query ActionExistsQuery) (*ActionExistsResult, error)
	GetAction(ctx corectx.Context, query GetActionQuery) (*GetActionResult, error)
	SearchActions(ctx corectx.Context, query SearchActionsQuery) (*SearchActionsResult, error)
	UpdateAction(ctx corectx.Context, cmd UpdateActionCommand) (*UpdateActionResult, error)
}
