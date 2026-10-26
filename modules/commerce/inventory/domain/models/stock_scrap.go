package models

import (
	_ "embed"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	StockScrapSchemaName = "inventory_stock_scrap"

	StockScrapFieldId               = basemodel.FieldId
	StockScrapFieldScrapNumber      = "scrap_number"
	StockScrapFieldOriginReference  = "origin_reference"
	StockScrapFieldTransferId       = "transfer_id"
	StockScrapFieldProductVariantId = "product_variant_id"
	StockScrapFieldBaseUomId        = "base_uom_id"
	StockScrapFieldLotRef           = "lot_ref"
	StockScrapFieldPackageRef       = "package_ref"
	StockScrapFieldOwnerRef         = "owner_ref"
	StockScrapFieldSourceLocationId = "source_location_id"
	StockScrapFieldScrapLocationId  = "scrap_location_id"
	StockScrapFieldQuantity         = "quantity"
	StockScrapFieldReasonCode       = "reason_code"
	StockScrapFieldReason           = "reason"
	StockScrapFieldStatus           = "status"
	StockScrapFieldMoveId           = "move_id"
	StockScrapFieldCompletedAt      = "completed_at"
	StockScrapFieldNote             = "note"
	StockScrapFieldOrgId            = "org_id"

	StockScrapEdgeTransfer       = "transfer"
	StockScrapEdgeProductVariant = "product_variant"
	StockScrapEdgeSourceLocation = "source_location"
	StockScrapEdgeScrapLocation  = "scrap_location"
)

// Stock scrap states. There is deliberately no cancelled: an unwanted draft is deleted, and a done
// scrap is corrected by a reverse movement rather than by reopening it.
const (
	StockScrapStatusDraft = "draft"
	StockScrapStatusDone  = "done"
)

//go:embed stock_scrap.json
var stockScrapSchemaJson string

func StockScrapSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(stockScrapSchemaJson)
}

// StockScrap removes goods from usable stock by moving them to a scrap location. It owns no
// balance: a draft changes nothing, and completing it generates a movement, the only thing that
// touches a quant.
type StockScrap struct {
	basemodel.DynamicModelBase
}

func NewStockScrap() *StockScrap {
	return &StockScrap{basemodel.NewDynamicModel()}
}

func NewStockScrapFrom(src dmodel.DynamicFields) *StockScrap {
	return &StockScrap{basemodel.NewDynamicModel(src)}
}

func (this StockScrap) GetScrapNumber() *string {
	return this.GetFieldData().GetString(StockScrapFieldScrapNumber)
}

func (this *StockScrap) SetScrapNumber(v *string) {
	this.GetFieldData().SetString(StockScrapFieldScrapNumber, v)
}

func (this StockScrap) GetTransferId() *model.Id {
	return this.GetFieldData().GetModelId(StockScrapFieldTransferId)
}

func (this *StockScrap) SetTransferId(v *model.Id) {
	this.GetFieldData().SetModelId(StockScrapFieldTransferId, v)
}

func (this StockScrap) GetProductVariantId() *model.Id {
	return this.GetFieldData().GetModelId(StockScrapFieldProductVariantId)
}

func (this *StockScrap) SetProductVariantId(v *model.Id) {
	this.GetFieldData().SetModelId(StockScrapFieldProductVariantId, v)
}

func (this StockScrap) GetBaseUomId() *model.Id {
	return this.GetFieldData().GetModelId(StockScrapFieldBaseUomId)
}

func (this *StockScrap) SetBaseUomId(v *model.Id) {
	this.GetFieldData().SetModelId(StockScrapFieldBaseUomId, v)
}

func (this StockScrap) GetLotRef() *string {
	return this.GetFieldData().GetString(StockScrapFieldLotRef)
}

func (this *StockScrap) SetLotRef(v *string) {
	this.GetFieldData().SetString(StockScrapFieldLotRef, v)
}

func (this StockScrap) GetPackageRef() *string {
	return this.GetFieldData().GetString(StockScrapFieldPackageRef)
}

func (this *StockScrap) SetPackageRef(v *string) {
	this.GetFieldData().SetString(StockScrapFieldPackageRef, v)
}

func (this StockScrap) GetOwnerRef() *string {
	return this.GetFieldData().GetString(StockScrapFieldOwnerRef)
}

func (this *StockScrap) SetOwnerRef(v *string) {
	this.GetFieldData().SetString(StockScrapFieldOwnerRef, v)
}

func (this StockScrap) GetSourceLocationId() *model.Id {
	return this.GetFieldData().GetModelId(StockScrapFieldSourceLocationId)
}

func (this *StockScrap) SetSourceLocationId(v *model.Id) {
	this.GetFieldData().SetModelId(StockScrapFieldSourceLocationId, v)
}

func (this StockScrap) GetScrapLocationId() *model.Id {
	return this.GetFieldData().GetModelId(StockScrapFieldScrapLocationId)
}

func (this *StockScrap) SetScrapLocationId(v *model.Id) {
	this.GetFieldData().SetModelId(StockScrapFieldScrapLocationId, v)
}

func (this StockScrap) GetQuantity() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(StockScrapFieldQuantity)
}

func (this *StockScrap) SetQuantity(v *decimal.Decimal) {
	this.GetFieldData().SetDecimal(StockScrapFieldQuantity, v)
}

func (this StockScrap) GetStatus() *string {
	return this.GetFieldData().GetString(StockScrapFieldStatus)
}

func (this *StockScrap) SetStatus(v *string) {
	this.GetFieldData().SetString(StockScrapFieldStatus, v)
}

func (this StockScrap) GetReasonCode() *string {
	return this.GetFieldData().GetString(StockScrapFieldReasonCode)
}

func (this *StockScrap) SetReasonCode(v *string) {
	this.GetFieldData().SetString(StockScrapFieldReasonCode, v)
}

func (this StockScrap) GetMoveId() *model.Id {
	return this.GetFieldData().GetModelId(StockScrapFieldMoveId)
}

func (this *StockScrap) SetMoveId(v *model.Id) {
	this.GetFieldData().SetModelId(StockScrapFieldMoveId, v)
}

func (this StockScrap) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(StockScrapFieldOrgId)
}

func (this *StockScrap) SetOrgId(v *model.Id) {
	this.GetFieldData().SetModelId(StockScrapFieldOrgId, v)
}
