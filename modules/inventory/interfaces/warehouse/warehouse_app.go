// Package warehouse is the contract for the Warehouse operations that span more than one
// resource.
//
// Plain warehouse CRUD does not appear here: the resource engine serves it and the domain service
// holds its rules. What is here are the operations that write a warehouse and its locations
// together, and the flow reads the Stock movement engine plans against.
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

// ResolvedLeg is one hop of a movement plan, with the location ids the engine will move between.
//
// A leg touching the outside world has an empty id on that side: which vendor or customer location
// applies depends on the transaction, and choosing it is the movement engine's business rather
// than the warehouse's.
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

// WarehouseAppService orchestrates the warehouse operations that touch locations too.
//
// Every one of them is configuration. None creates a stock move, changes a quantity, or alters a
// transfer already under way — a flow change applies to transactions made after it.
type WarehouseAppService interface {
	CreateWarehouse(ctx corectx.Context, cmd CreateWarehouseCommand) (*CreateWarehouseResult, error)

	ConfigureIncomingFlow(ctx corectx.Context, cmd ConfigureFlowCommand) (*ConfigureFlowResult, error)
	ConfigureOutgoingFlow(ctx corectx.Context, cmd ConfigureFlowCommand) (*ConfigureFlowResult, error)

	ResolveIncomingFlow(ctx corectx.Context, query ResolveFlowQuery) (*ResolveFlowResult, error)
	ResolveOutgoingFlow(ctx corectx.Context, query ResolveFlowQuery) (*ResolveFlowResult, error)
}
