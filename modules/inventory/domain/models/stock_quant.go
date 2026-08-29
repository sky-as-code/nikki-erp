package models

import (
	_ "embed"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	StockQuantSchemaName = "inventory_stock_quant"

	StockQuantFieldId                = basemodel.FieldId
	StockQuantFieldProductVariantId  = "product_variant_id"
	StockQuantFieldLocationId        = "location_id"
	StockQuantFieldLotRef            = "lot_ref"
	StockQuantFieldPackageRef        = "package_ref"
	StockQuantFieldOwnerRef          = "owner_ref"
	StockQuantFieldBaseUomId         = "base_uom_id"
	StockQuantFieldOnHandQuantity    = "on_hand_quantity"
	StockQuantFieldReservedQuantity  = "reserved_quantity"
	StockQuantFieldAvailableQuantity = "available_quantity"
	StockQuantFieldIncomingDate      = "incoming_date"
	StockQuantFieldCountedQuantity   = "counted_quantity"
	StockQuantFieldCountQuantitySet  = "count_quantity_set"
	StockQuantFieldCountSnapshotQty  = "count_snapshot_quantity"
	StockQuantFieldCountReasonCode   = "count_reason_code"
	StockQuantFieldCountReasonText   = "count_reason_text"
	StockQuantFieldNextCountDate     = "next_count_date"
	StockQuantFieldLastCountDate     = "last_count_date"
	StockQuantFieldCountAssignedUser = "count_assigned_user_id"
	StockQuantFieldOrgId             = "org_id"

	StockQuantEdgeProductVariant = "product_variant"
	StockQuantEdgeLocation       = "location"
)

//go:embed stock_quant.json
var stockQuantSchemaJson string

func StockQuantSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(stockQuantSchemaJson)
}

// StockQuant is the current stock state at one dimension combination: a variant, in a location,
// optionally narrowed by lot, package and owner. Its quantities are the running total of completed
// movements, so no client may write them directly.
type StockQuant struct {
	basemodel.DynamicModelBase
}

func NewStockQuant() *StockQuant {
	return &StockQuant{basemodel.NewDynamicModel()}
}

func NewStockQuantFrom(src dmodel.DynamicFields) *StockQuant {
	return &StockQuant{basemodel.NewDynamicModel(src)}
}

func (this StockQuant) GetProductVariantId() *model.Id {
	return this.GetFieldData().GetModelId(StockQuantFieldProductVariantId)
}

func (this *StockQuant) SetProductVariantId(v *model.Id) {
	this.GetFieldData().SetModelId(StockQuantFieldProductVariantId, v)
}

func (this StockQuant) GetLocationId() *model.Id {
	return this.GetFieldData().GetModelId(StockQuantFieldLocationId)
}

func (this *StockQuant) SetLocationId(v *model.Id) {
	this.GetFieldData().SetModelId(StockQuantFieldLocationId, v)
}

func (this StockQuant) GetLotRef() *string {
	return this.GetFieldData().GetString(StockQuantFieldLotRef)
}

func (this *StockQuant) SetLotRef(v *string) {
	this.GetFieldData().SetString(StockQuantFieldLotRef, v)
}

func (this StockQuant) GetPackageRef() *string {
	return this.GetFieldData().GetString(StockQuantFieldPackageRef)
}

func (this *StockQuant) SetPackageRef(v *string) {
	this.GetFieldData().SetString(StockQuantFieldPackageRef, v)
}

func (this StockQuant) GetOwnerRef() *string {
	return this.GetFieldData().GetString(StockQuantFieldOwnerRef)
}

func (this *StockQuant) SetOwnerRef(v *string) {
	this.GetFieldData().SetString(StockQuantFieldOwnerRef, v)
}

func (this StockQuant) GetOnHandQuantity() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(StockQuantFieldOnHandQuantity)
}

func (this *StockQuant) SetOnHandQuantity(v *decimal.Decimal) {
	this.GetFieldData().SetDecimal(StockQuantFieldOnHandQuantity, v)
}

func (this StockQuant) GetReservedQuantity() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(StockQuantFieldReservedQuantity)
}

func (this *StockQuant) SetReservedQuantity(v *decimal.Decimal) {
	this.GetFieldData().SetDecimal(StockQuantFieldReservedQuantity, v)
}

// GetAvailableQuantity reads a virtual field the quant service fills on read; it is absent on a
// model the caller built itself, where services.AvailableQuantity is needed instead.
func (this StockQuant) GetAvailableQuantity() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(StockQuantFieldAvailableQuantity)
}

func (this *StockQuant) SetAvailableQuantity(v *decimal.Decimal) {
	this.GetFieldData().SetDecimal(StockQuantFieldAvailableQuantity, v)
}

func (this StockQuant) GetCountedQuantity() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(StockQuantFieldCountedQuantity)
}

func (this *StockQuant) SetCountedQuantity(v *decimal.Decimal) {
	this.GetFieldData().SetDecimal(StockQuantFieldCountedQuantity, v)
}

func (this StockQuant) GetCountQuantitySet() *bool {
	return this.GetFieldData().GetBool(StockQuantFieldCountQuantitySet)
}

func (this *StockQuant) SetCountQuantitySet(v *bool) {
	this.GetFieldData().SetBool(StockQuantFieldCountQuantitySet, v)
}

func (this StockQuant) GetCountSnapshotQuantity() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(StockQuantFieldCountSnapshotQty)
}

func (this *StockQuant) SetCountSnapshotQuantity(v *decimal.Decimal) {
	this.GetFieldData().SetDecimal(StockQuantFieldCountSnapshotQty, v)
}

func (this StockQuant) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(StockQuantFieldOrgId)
}

func (this *StockQuant) SetOrgId(v *model.Id) {
	this.GetFieldData().SetModelId(StockQuantFieldOrgId, v)
}
