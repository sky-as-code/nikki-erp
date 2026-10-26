// Package warehouse is the contract for Warehouse operations spanning more than one resource:
// writes that touch a warehouse and its locations together, plus the flow reads the Stock
// movement engine plans against. Plain warehouse CRUD is served by the resource engine instead.
package warehouse

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

type CreateWarehouseCommand struct {
	Fields dmodel.DynamicFields
}

type CreateWarehouseResultData struct {
	Warehouse dmodel.DynamicFields
}

type CreateWarehouseResult = dyn.OpResult[CreateWarehouseResultData]

type ConfigureFlowCommand struct {
	WarehouseId string
	Flow        string
}

type ConfigureFlowResultData struct {
	AffectedCount int
}

type ConfigureFlowResult = dyn.OpResult[ConfigureFlowResultData]

type ResolveFlowQuery struct {
	WarehouseId string
}

// ResolvedLeg is one hop of a movement plan. A leg touching the outside world has an empty id on
// that side: which vendor or customer location applies depends on the transaction, and the
// movement engine chooses it.
type ResolvedLeg struct {
	FromLocationId string
	ToLocationId   string
	FromCode       string
	ToCode         string
}

type ResolveFlowResultData struct {
	Legs []ResolvedLeg
}

type ResolveFlowResult = dyn.OpResult[ResolveFlowResultData]

// WarehouseAppService orchestrates the warehouse operations that touch locations too. All are
// configuration: none creates a stock move, changes a quantity, or alters a transfer already
// under way — a flow change applies only to later transactions.
type WarehouseAppService interface {
	CreateWarehouse(ctx corectx.Context, cmd CreateWarehouseCommand) (*CreateWarehouseResult, error)

	ConfigureIncomingFlow(ctx corectx.Context, cmd ConfigureFlowCommand) (*ConfigureFlowResult, error)
	ConfigureOutgoingFlow(ctx corectx.Context, cmd ConfigureFlowCommand) (*ConfigureFlowResult, error)

	ResolveIncomingFlow(ctx corectx.Context, query ResolveFlowQuery) (*ResolveFlowResult, error)
	ResolveOutgoingFlow(ctx corectx.Context, query ResolveFlowQuery) (*ResolveFlowResult, error)
}
