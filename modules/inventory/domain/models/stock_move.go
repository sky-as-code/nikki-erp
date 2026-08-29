package models

import (
	_ "embed"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	StockMoveSchemaName = "inventory_stock_move"

	StockMoveFieldId                    = basemodel.FieldId
	StockMoveFieldTransferId            = "transfer_id"
	StockMoveFieldSequence              = "sequence"
	StockMoveFieldProductVariantId      = "product_variant_id"
	StockMoveFieldUomId                 = "uom_id"
	StockMoveFieldDemandQuantity        = "demand_quantity"
	StockMoveFieldBaseDemandQuantity    = "base_demand_quantity"
	StockMoveFieldSourceLocationId      = "source_location_id"
	StockMoveFieldDestinationLocationId = "destination_location_id"
	StockMoveFieldFinalLocationId       = "final_location_id"
	StockMoveFieldStatus                = "status"
	StockMoveFieldPriority              = "priority"
	StockMoveFieldScheduledAt           = "scheduled_at"
	StockMoveFieldDeadlineAt            = "deadline_at"
	StockMoveFieldReservationDate       = "reservation_date"
	StockMoveFieldPicked                = "picked"
	StockMoveFieldOriginMoveId          = "origin_move_id"
	StockMoveFieldIsInventoryAdjustment = "is_inventory_adjustment"
	StockMoveFieldScrapId               = "scrap_id"
	StockMoveFieldValuationValue        = "valuation_value"
	StockMoveFieldRemainingQuantity     = "remaining_quantity"
	StockMoveFieldRemainingValue        = "remaining_value"
	StockMoveFieldCurrencyId            = "currency_id"
	StockMoveFieldOrgId                 = "org_id"

	StockMoveEdgeTransfer            = "transfer"
	StockMoveEdgeProductVariant      = "product_variant"
	StockMoveEdgeSourceLocation      = "source_location"
	StockMoveEdgeDestinationLocation = "destination_location"
)

// Stock move lifecycle states. partially_available and assigned have no transfer equivalent: a move
// knows how much of its demand is reserved, a transfer only whether all its moves are ready.
const (
	StockMoveStatusDraft              = "draft"
	StockMoveStatusWaiting            = "waiting"
	StockMoveStatusConfirmed          = "confirmed"
	StockMoveStatusPartiallyAvailable = "partially_available"
	StockMoveStatusAssigned           = "assigned"
	StockMoveStatusDone               = "done"
	StockMoveStatusCancelled          = "cancelled"
)

//go:embed stock_move.json
var stockMoveSchemaJson string

func StockMoveSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(stockMoveSchemaJson)
}

// StockMove is one line of demand. It records what was asked for, never what was delivered; its
// move lines hold that, and keeping them apart lets a partial delivery be reported without
// rewriting the original demand.
type StockMove struct {
	basemodel.DynamicModelBase
}

func NewStockMove() *StockMove {
	return &StockMove{basemodel.NewDynamicModel()}
}

func NewStockMoveFrom(src dmodel.DynamicFields) *StockMove {
	return &StockMove{basemodel.NewDynamicModel(src)}
}

func (this StockMove) GetTransferId() *model.Id {
	return this.GetFieldData().GetModelId(StockMoveFieldTransferId)
}

func (this *StockMove) SetTransferId(v *model.Id) {
	this.GetFieldData().SetModelId(StockMoveFieldTransferId, v)
}

func (this StockMove) GetProductVariantId() *model.Id {
	return this.GetFieldData().GetModelId(StockMoveFieldProductVariantId)
}

func (this *StockMove) SetProductVariantId(v *model.Id) {
	this.GetFieldData().SetModelId(StockMoveFieldProductVariantId, v)
}

func (this StockMove) GetDemandQuantity() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(StockMoveFieldDemandQuantity)
}

func (this *StockMove) SetDemandQuantity(v *decimal.Decimal) {
	this.GetFieldData().SetDecimal(StockMoveFieldDemandQuantity, v)
}

func (this StockMove) GetBaseDemandQuantity() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(StockMoveFieldBaseDemandQuantity)
}

func (this *StockMove) SetBaseDemandQuantity(v *decimal.Decimal) {
	this.GetFieldData().SetDecimal(StockMoveFieldBaseDemandQuantity, v)
}

func (this StockMove) GetSourceLocationId() *model.Id {
	return this.GetFieldData().GetModelId(StockMoveFieldSourceLocationId)
}

func (this *StockMove) SetSourceLocationId(v *model.Id) {
	this.GetFieldData().SetModelId(StockMoveFieldSourceLocationId, v)
}

func (this StockMove) GetDestinationLocationId() *model.Id {
	return this.GetFieldData().GetModelId(StockMoveFieldDestinationLocationId)
}

func (this *StockMove) SetDestinationLocationId(v *model.Id) {
	this.GetFieldData().SetModelId(StockMoveFieldDestinationLocationId, v)
}

func (this StockMove) GetStatus() *string {
	return this.GetFieldData().GetString(StockMoveFieldStatus)
}

func (this *StockMove) SetStatus(v *string) {
	this.GetFieldData().SetString(StockMoveFieldStatus, v)
}

func (this StockMove) GetPicked() *bool {
	return this.GetFieldData().GetBool(StockMoveFieldPicked)
}

func (this *StockMove) SetPicked(v *bool) {
	this.GetFieldData().SetBool(StockMoveFieldPicked, v)
}

func (this StockMove) GetOriginMoveId() *model.Id {
	return this.GetFieldData().GetModelId(StockMoveFieldOriginMoveId)
}

func (this *StockMove) SetOriginMoveId(v *model.Id) {
	this.GetFieldData().SetModelId(StockMoveFieldOriginMoveId, v)
}

func (this StockMove) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(StockMoveFieldOrgId)
}

func (this *StockMove) SetOrgId(v *model.Id) {
	this.GetFieldData().SetModelId(StockMoveFieldOrgId, v)
}
