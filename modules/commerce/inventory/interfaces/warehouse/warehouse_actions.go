package warehouse

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

// The custom actions of the warehouse resources. Each command is the bound request: the record
// id from the path plus whatever the body carried.
type (
	SuspendWarehouseCommand       = dmodel.DynamicFields
	ResumeWarehouseCommand        = dmodel.DynamicFields
	ConfigureWarehouseFlowCommand = dmodel.DynamicFields

	SuspendInventoryLocationCommand = dmodel.DynamicFields
	ResumeInventoryLocationCommand  = dmodel.DynamicFields
	MoveInventoryLocationCommand    = dmodel.DynamicFields

	SuggestPutawayLocationQuery = dmodel.DynamicFields
)

type (
	SuspendWarehouseResult       = dyn.OpResult[dyn.MutateResultData]
	ResumeWarehouseResult        = dyn.OpResult[dyn.MutateResultData]
	ConfigureWarehouseFlowResult = dyn.OpResult[dyn.MutateResultData]

	SuspendInventoryLocationResult = dyn.OpResult[dyn.MutateResultData]
	ResumeInventoryLocationResult  = dyn.OpResult[dyn.MutateResultData]
	MoveInventoryLocationResult    = dyn.OpResult[dyn.MutateResultData]

	SuggestPutawayLocationResult = dyn.OpResult[any]
)

// WarehouseActionService is the authorized custom surface of the warehouse: the lifecycle
// operations and the flow reconfigurations, each asserting its own permission.
type WarehouseActionService interface {
	Suspend(ctx corectx.Context, cmd SuspendWarehouseCommand) (*SuspendWarehouseResult, error)
	Resume(ctx corectx.Context, cmd ResumeWarehouseCommand) (*ResumeWarehouseResult, error)
	ConfigureIncomingFlowAction(ctx corectx.Context, cmd ConfigureWarehouseFlowCommand) (*ConfigureWarehouseFlowResult, error)
	ConfigureOutgoingFlowAction(ctx corectx.Context, cmd ConfigureWarehouseFlowCommand) (*ConfigureWarehouseFlowResult, error)
}

// InventoryLocationActionService is the authorized custom surface of a location. Suspend is
// allowed while the location still holds stock, whereas archiving one that does is refused by
// the built-in set_archived; the asymmetry is enforced in the domain service.
type InventoryLocationActionService interface {
	Suspend(ctx corectx.Context, cmd SuspendInventoryLocationCommand) (*SuspendInventoryLocationResult, error)
	Resume(ctx corectx.Context, cmd ResumeInventoryLocationCommand) (*ResumeInventoryLocationResult, error)
	Move(ctx corectx.Context, cmd MoveInventoryLocationCommand) (*MoveInventoryLocationResult, error)
}

// PutawayRuleActionService answers where arriving goods should be put. It asks a question of
// the whole rule set rather than acting on one rule, and changes nothing.
type PutawayRuleActionService interface {
	SuggestLocation(ctx corectx.Context, query SuggestPutawayLocationQuery) (*SuggestPutawayLocationResult, error)
}
