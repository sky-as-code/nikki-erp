package stock

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

// The custom actions of the stock resources, as the bound request.
type (
	EnterCountCommand      = dmodel.DynamicFields
	ResetCountCommand      = dmodel.DynamicFields
	ApplyAdjustmentCommand = dmodel.DynamicFields
	ScheduleCountCommand   = dmodel.DynamicFields
	AssignCounterCommand   = dmodel.DynamicFields

	VariantStockSummaryQuery   = dmodel.DynamicFields
	VariantStockSummariesQuery = dmodel.DynamicFields
	TemplateStockSummaryQuery  = dmodel.DynamicFields
	StockByWarehouseQuery      = dmodel.DynamicFields
	StockByLocationQuery       = dmodel.DynamicFields
	ProductUsageQuery          = dmodel.DynamicFields

	ConfirmTransferCommand        = dmodel.DynamicFields
	CheckAvailabilityQuery        = dmodel.DynamicFields
	ReserveTransferCommand        = dmodel.DynamicFields
	UnreserveTransferCommand      = dmodel.DynamicFields
	ValidateTransferCommand       = dmodel.DynamicFields
	CancelTransferCommand         = dmodel.DynamicFields
	CreateReturnCommand           = dmodel.DynamicFields
	ReserveForSourceCommand       = dmodel.DynamicFields
	ReleaseForSourceCommand       = dmodel.DynamicFields
	ReallocateReservationCommand  = dmodel.DynamicFields
	ApplyFulfillmentResultCommand = dmodel.DynamicFields

	DoScrapCommand = dmodel.DynamicFields
)

type (
	CountMutateResult      = dyn.OpResult[dyn.MutateResultData]
	ProductStockReadResult = dyn.OpResult[any]
	TransferMutateResult   = dyn.OpResult[dyn.MutateResultData]
	TransferReadResult     = dyn.OpResult[any]
	DoScrapResult          = dyn.OpResult[dyn.MutateResultData]
)

// StockQuantActionService is the authorized custom surface of the quant: physical inventory and
// cycle counting, plus the product-facing reads that let the Product UI show stock it does not
// own. Counting writes only count metadata; Apply changes the balance solely by generating a
// movement. The reads create no movement, reserve nothing and change no balance.
type StockQuantActionService interface {
	EnterCount(ctx corectx.Context, cmd EnterCountCommand) (*CountMutateResult, error)
	ResetCount(ctx corectx.Context, cmd ResetCountCommand) (*CountMutateResult, error)
	ApplyAdjustment(ctx corectx.Context, cmd ApplyAdjustmentCommand) (*CountMutateResult, error)
	ScheduleCount(ctx corectx.Context, cmd ScheduleCountCommand) (*CountMutateResult, error)
	AssignCounter(ctx corectx.Context, cmd AssignCounterCommand) (*CountMutateResult, error)

	VariantStockSummary(ctx corectx.Context, query VariantStockSummaryQuery) (*ProductStockReadResult, error)
	VariantStockSummaries(ctx corectx.Context, query VariantStockSummariesQuery) (*ProductStockReadResult, error)
	TemplateStockSummary(ctx corectx.Context, query TemplateStockSummaryQuery) (*ProductStockReadResult, error)
	StockByWarehouse(ctx corectx.Context, query StockByWarehouseQuery) (*ProductStockReadResult, error)
	StockByLocation(ctx corectx.Context, query StockByLocationQuery) (*ProductStockReadResult, error)
	ProductUsage(ctx corectx.Context, query ProductUsageQuery) (*ProductStockReadResult, error)
}

// StockTransferActionService is the authorized custom surface of a transfer: the movement
// operations addressed by transfer id, and the reservation operations addressed by the demand
// for a caller in another module that knows only what it asked for.
type StockTransferActionService interface {
	Confirm(ctx corectx.Context, cmd ConfirmTransferCommand) (*TransferMutateResult, error)
	CheckAvailability(ctx corectx.Context, query CheckAvailabilityQuery) (*TransferReadResult, error)
	Reserve(ctx corectx.Context, cmd ReserveTransferCommand) (*TransferMutateResult, error)
	Unreserve(ctx corectx.Context, cmd UnreserveTransferCommand) (*TransferMutateResult, error)
	Validate(ctx corectx.Context, cmd ValidateTransferCommand) (*TransferMutateResult, error)
	Cancel(ctx corectx.Context, cmd CancelTransferCommand) (*TransferMutateResult, error)
	CreateReturn(ctx corectx.Context, cmd CreateReturnCommand) (*TransferMutateResult, error)

	ReserveForSource(ctx corectx.Context, cmd ReserveForSourceCommand) (*TransferReadResult, error)
	ReleaseForSource(ctx corectx.Context, cmd ReleaseForSourceCommand) (*TransferMutateResult, error)
	ReallocateReservation(ctx corectx.Context, cmd ReallocateReservationCommand) (*TransferReadResult, error)
	ApplyFulfillmentResult(ctx corectx.Context, cmd ApplyFulfillmentResultCommand) (*TransferReadResult, error)
}

// StockScrapActionService executes a scrap document: the event that writes off the goods and
// cannot be undone by editing the record afterwards.
type StockScrapActionService interface {
	DoScrap(ctx corectx.Context, cmd DoScrapCommand) (*DoScrapResult, error)
}
