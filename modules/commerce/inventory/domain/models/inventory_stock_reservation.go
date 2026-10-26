package models

import (
	_ "embed"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	StockReservationSchemaName = "inventory_stock_reservation"

	StockReservationFieldId                        = basemodel.FieldId
	StockReservationFieldWarehouseId               = "warehouse_id"
	StockReservationFieldProductVariantId          = "product_variant_id"
	StockReservationFieldBaseUomId                 = "base_uom_id"
	StockReservationFieldQuantity                  = "quantity"
	StockReservationFieldConsumedQuantity          = "consumed_quantity"
	StockReservationFieldReleasedQuantity          = "released_quantity"
	StockReservationFieldRemainingQuantity         = "remaining_quantity"
	StockReservationFieldReservedUntil             = "reserved_until"
	StockReservationFieldStatus                    = "status"
	StockReservationFieldEffectiveStatus           = "effective_status"
	StockReservationFieldEffectiveReservedQuantity = "effective_reserved_quantity"
	StockReservationFieldSourceModule              = "source_module"
	StockReservationFieldSourceType                = "source_type"
	StockReservationFieldSourceId                  = "source_id"
	StockReservationFieldSourceLineId              = "source_line_id"
	StockReservationFieldSourceRevision            = "source_revision"
	StockReservationFieldIdempotencyKey            = "idempotency_key"
	StockReservationFieldRequestFingerprint        = "request_fingerprint"
	StockReservationFieldReleaseReason             = "release_reason"
	StockReservationFieldReleasedAt                = "released_at"
	StockReservationFieldExpiryRecordedAt          = "expiry_recorded_at"
	StockReservationFieldOrgId                     = "org_id"

	StockReservationEdgeWarehouse      = "warehouse"
	StockReservationEdgeProductVariant = "product_variant"
	StockReservationEdgeBaseUom        = "base_uom"
)

// Stored lifecycle of a reservation. Expiry is deliberately absent: it is a condition of the clock,
// derived on read as StockReservationEffectiveStatusExpired, because a stored flag could only ever
// lag the deadline and a lagging flag over-commits stock.
const (
	StockReservationStatusActive   = "active"
	StockReservationStatusConsumed = "consumed"
	StockReservationStatusReleased = "released"

	StockReservationEffectiveStatusExpired = "expired"
)

// StockReservationReleaseReasonExpired is the release_reason written when expiry is materialized,
// as opposed to a reason the caller supplied.
const StockReservationReleaseReasonExpired = "expired"

//go:embed inventory_stock_reservation.json
var stockReservationSchemaJson string

func StockReservationSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(stockReservationSchemaJson)
}

// StockReservation commits a quantity of one variant at one warehouse to one demand line. It names
// no location and touches no quant: the quant's reserved figure stays for concrete allocations, and
// warehouse availability subtracts this row's effective remainder on top of that.
type StockReservation struct {
	basemodel.DynamicModelBase
}

func NewStockReservation() *StockReservation {
	return &StockReservation{basemodel.NewDynamicModel()}
}

func NewStockReservationFrom(src dmodel.DynamicFields) *StockReservation {
	return &StockReservation{basemodel.NewDynamicModel(src)}
}

func (this StockReservation) GetId() *model.Id {
	return this.GetFieldData().GetModelId(StockReservationFieldId)
}

func (this StockReservation) GetWarehouseId() *model.Id {
	return this.GetFieldData().GetModelId(StockReservationFieldWarehouseId)
}

func (this *StockReservation) SetWarehouseId(v *model.Id) {
	this.GetFieldData().SetModelId(StockReservationFieldWarehouseId, v)
}

func (this StockReservation) GetProductVariantId() *model.Id {
	return this.GetFieldData().GetModelId(StockReservationFieldProductVariantId)
}

func (this *StockReservation) SetProductVariantId(v *model.Id) {
	this.GetFieldData().SetModelId(StockReservationFieldProductVariantId, v)
}

func (this StockReservation) GetBaseUomId() *model.Id {
	return this.GetFieldData().GetModelId(StockReservationFieldBaseUomId)
}

func (this *StockReservation) SetBaseUomId(v *model.Id) {
	this.GetFieldData().SetModelId(StockReservationFieldBaseUomId, v)
}

func (this StockReservation) GetQuantity() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(StockReservationFieldQuantity)
}

func (this *StockReservation) SetQuantity(v *decimal.Decimal) {
	this.GetFieldData().SetDecimal(StockReservationFieldQuantity, v)
}

func (this StockReservation) GetConsumedQuantity() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(StockReservationFieldConsumedQuantity)
}

func (this *StockReservation) SetConsumedQuantity(v *decimal.Decimal) {
	this.GetFieldData().SetDecimal(StockReservationFieldConsumedQuantity, v)
}

func (this StockReservation) GetReleasedQuantity() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(StockReservationFieldReleasedQuantity)
}

func (this *StockReservation) SetReleasedQuantity(v *decimal.Decimal) {
	this.GetFieldData().SetDecimal(StockReservationFieldReleasedQuantity, v)
}

// RemainingQuantity computes quantity - consumed - released from the stored columns rather than
// reading the projected field, so it is right on a row that was loaded without computed fields.
func (this StockReservation) RemainingQuantity() decimal.Decimal {
	return decimalOrZero(this.GetQuantity()).
		Sub(decimalOrZero(this.GetConsumedQuantity())).
		Sub(decimalOrZero(this.GetReleasedQuantity()))
}

func (this StockReservation) GetReservedUntil() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(StockReservationFieldReservedUntil)
}

func (this *StockReservation) SetReservedUntil(v *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(StockReservationFieldReservedUntil, v)
}

func (this StockReservation) GetStatus() *string {
	return this.GetFieldData().GetString(StockReservationFieldStatus)
}

func (this *StockReservation) SetStatus(v *string) {
	this.GetFieldData().SetString(StockReservationFieldStatus, v)
}

func (this StockReservation) GetSourceModule() *string {
	return this.GetFieldData().GetString(StockReservationFieldSourceModule)
}

func (this *StockReservation) SetSourceModule(v *string) {
	this.GetFieldData().SetString(StockReservationFieldSourceModule, v)
}

func (this StockReservation) GetSourceType() *string {
	return this.GetFieldData().GetString(StockReservationFieldSourceType)
}

func (this *StockReservation) SetSourceType(v *string) {
	this.GetFieldData().SetString(StockReservationFieldSourceType, v)
}

func (this StockReservation) GetSourceId() *string {
	return this.GetFieldData().GetString(StockReservationFieldSourceId)
}

func (this *StockReservation) SetSourceId(v *string) {
	this.GetFieldData().SetString(StockReservationFieldSourceId, v)
}

func (this StockReservation) GetSourceLineId() *string {
	return this.GetFieldData().GetString(StockReservationFieldSourceLineId)
}

func (this *StockReservation) SetSourceLineId(v *string) {
	this.GetFieldData().SetString(StockReservationFieldSourceLineId, v)
}

func (this StockReservation) GetSourceRevision() *int32 {
	return this.GetFieldData().GetInt32(StockReservationFieldSourceRevision)
}

func (this *StockReservation) SetSourceRevision(v *int32) {
	this.GetFieldData().SetInt32(StockReservationFieldSourceRevision, v)
}

func (this StockReservation) GetIdempotencyKey() *string {
	return this.GetFieldData().GetString(StockReservationFieldIdempotencyKey)
}

func (this *StockReservation) SetIdempotencyKey(v *string) {
	this.GetFieldData().SetString(StockReservationFieldIdempotencyKey, v)
}

func (this StockReservation) GetRequestFingerprint() *string {
	return this.GetFieldData().GetString(StockReservationFieldRequestFingerprint)
}

func (this *StockReservation) SetRequestFingerprint(v *string) {
	this.GetFieldData().SetString(StockReservationFieldRequestFingerprint, v)
}

func (this StockReservation) GetReleaseReason() *string {
	return this.GetFieldData().GetString(StockReservationFieldReleaseReason)
}

func (this *StockReservation) SetReleaseReason(v *string) {
	this.GetFieldData().SetString(StockReservationFieldReleaseReason, v)
}

func (this StockReservation) GetReleasedAt() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(StockReservationFieldReleasedAt)
}

func (this *StockReservation) SetReleasedAt(v *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(StockReservationFieldReleasedAt, v)
}

func (this StockReservation) GetExpiryRecordedAt() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(StockReservationFieldExpiryRecordedAt)
}

func (this *StockReservation) SetExpiryRecordedAt(v *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(StockReservationFieldExpiryRecordedAt, v)
}

func (this StockReservation) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(StockReservationFieldOrgId)
}

func (this *StockReservation) SetOrgId(v *model.Id) {
	this.GetFieldData().SetModelId(StockReservationFieldOrgId, v)
}

func decimalOrZero(v *decimal.Decimal) decimal.Decimal {
	if v == nil {
		return decimal.Zero
	}
	return *v
}
