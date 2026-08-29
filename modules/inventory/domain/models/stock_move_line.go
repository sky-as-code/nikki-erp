package models

import (
	_ "embed"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	StockMoveLineSchemaName = "inventory_stock_move_line"

	StockMoveLineFieldId                    = basemodel.FieldId
	StockMoveLineFieldMoveId                = "move_id"
	StockMoveLineFieldTransferId            = "transfer_id"
	StockMoveLineFieldProductVariantId      = "product_variant_id"
	StockMoveLineFieldUomId                 = "uom_id"
	StockMoveLineFieldQuantity              = "quantity"
	StockMoveLineFieldBaseQuantity          = "base_quantity"
	StockMoveLineFieldSourceLocationId      = "source_location_id"
	StockMoveLineFieldDestinationLocationId = "destination_location_id"
	StockMoveLineFieldLotRef                = "lot_ref"
	StockMoveLineFieldPackageRef            = "package_ref"
	StockMoveLineFieldResultPackageRef      = "result_package_ref"
	StockMoveLineFieldOwnerRef              = "owner_ref"
	StockMoveLineFieldPicked                = "picked"
	StockMoveLineFieldOperationAt           = "operation_at"
	StockMoveLineFieldOrgId                 = "org_id"

	StockMoveLineEdgeMove                = "move"
	StockMoveLineEdgeTransfer            = "transfer"
	StockMoveLineEdgeProductVariant      = "product_variant"
	StockMoveLineEdgeSourceLocation      = "source_location"
	StockMoveLineEdgeDestinationLocation = "destination_location"
)

//go:embed stock_move_line.json
var stockMoveLineSchemaJson string

func StockMoveLineSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(stockMoveLineSchemaJson)
}

// StockMoveLine is the execution detail of a move: the exact quantity taken from an exact quant
// dimension. It has no status; it is a reservation until operation_at is stamped, a recorded
// movement afterwards.
type StockMoveLine struct {
	basemodel.DynamicModelBase
}

func NewStockMoveLine() *StockMoveLine {
	return &StockMoveLine{basemodel.NewDynamicModel()}
}

func NewStockMoveLineFrom(src dmodel.DynamicFields) *StockMoveLine {
	return &StockMoveLine{basemodel.NewDynamicModel(src)}
}

func (this StockMoveLine) GetMoveId() *model.Id {
	return this.GetFieldData().GetModelId(StockMoveLineFieldMoveId)
}

func (this *StockMoveLine) SetMoveId(v *model.Id) {
	this.GetFieldData().SetModelId(StockMoveLineFieldMoveId, v)
}

func (this StockMoveLine) GetTransferId() *model.Id {
	return this.GetFieldData().GetModelId(StockMoveLineFieldTransferId)
}

func (this *StockMoveLine) SetTransferId(v *model.Id) {
	this.GetFieldData().SetModelId(StockMoveLineFieldTransferId, v)
}

func (this StockMoveLine) GetProductVariantId() *model.Id {
	return this.GetFieldData().GetModelId(StockMoveLineFieldProductVariantId)
}

func (this *StockMoveLine) SetProductVariantId(v *model.Id) {
	this.GetFieldData().SetModelId(StockMoveLineFieldProductVariantId, v)
}

func (this StockMoveLine) GetQuantity() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(StockMoveLineFieldQuantity)
}

func (this *StockMoveLine) SetQuantity(v *decimal.Decimal) {
	this.GetFieldData().SetDecimal(StockMoveLineFieldQuantity, v)
}

func (this StockMoveLine) GetBaseQuantity() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(StockMoveLineFieldBaseQuantity)
}

func (this *StockMoveLine) SetBaseQuantity(v *decimal.Decimal) {
	this.GetFieldData().SetDecimal(StockMoveLineFieldBaseQuantity, v)
}

func (this StockMoveLine) GetSourceLocationId() *model.Id {
	return this.GetFieldData().GetModelId(StockMoveLineFieldSourceLocationId)
}

func (this *StockMoveLine) SetSourceLocationId(v *model.Id) {
	this.GetFieldData().SetModelId(StockMoveLineFieldSourceLocationId, v)
}

func (this StockMoveLine) GetDestinationLocationId() *model.Id {
	return this.GetFieldData().GetModelId(StockMoveLineFieldDestinationLocationId)
}

func (this *StockMoveLine) SetDestinationLocationId(v *model.Id) {
	this.GetFieldData().SetModelId(StockMoveLineFieldDestinationLocationId, v)
}

func (this StockMoveLine) GetLotRef() *string {
	return this.GetFieldData().GetString(StockMoveLineFieldLotRef)
}

func (this *StockMoveLine) SetLotRef(v *string) {
	this.GetFieldData().SetString(StockMoveLineFieldLotRef, v)
}

func (this StockMoveLine) GetPackageRef() *string {
	return this.GetFieldData().GetString(StockMoveLineFieldPackageRef)
}

func (this *StockMoveLine) SetPackageRef(v *string) {
	this.GetFieldData().SetString(StockMoveLineFieldPackageRef, v)
}

func (this StockMoveLine) GetOwnerRef() *string {
	return this.GetFieldData().GetString(StockMoveLineFieldOwnerRef)
}

func (this *StockMoveLine) SetOwnerRef(v *string) {
	this.GetFieldData().SetString(StockMoveLineFieldOwnerRef, v)
}

func (this StockMoveLine) GetPicked() *bool {
	return this.GetFieldData().GetBool(StockMoveLineFieldPicked)
}

func (this *StockMoveLine) SetPicked(v *bool) {
	this.GetFieldData().SetBool(StockMoveLineFieldPicked, v)
}

func (this StockMoveLine) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(StockMoveLineFieldOrgId)
}

func (this *StockMoveLine) SetOrgId(v *model.Id) {
	this.GetFieldData().SetModelId(StockMoveLineFieldOrgId, v)
}
