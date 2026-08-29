package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// The direction of the movements a transfer of this type performs.
const (
	StockOperationCodeIncoming = "incoming"
	StockOperationCodeOutgoing = "outgoing"
	StockOperationCodeInternal = "internal"
)

// When stock is reserved for a transfer of this type.
const (
	StockReservationMethodAtConfirmation     = "at_confirmation"
	StockReservationMethodManual             = "manual"
	StockReservationMethodBeforeScheduledDay = "before_scheduled_date"
)

// What happens to the quantity a partially-processed transfer did not deliver.
const (
	StockBackorderPolicyAsk    = "ask"
	StockBackorderPolicyAlways = "always"
	StockBackorderPolicyNever  = "never"
)

// Whether a transfer of this type may ship what it has, or must wait for all of it.
const (
	StockShippingPolicyPartial   = "partial"
	StockShippingPolicyAllAtOnce = "all_at_once"
)

// StockCorrectionOperationTypeCode is the `code` of the operation type adjustments and scraps
// generate their movements through. Seeded once per org; corrections resolve it by this code rather
// than take one from the caller. The seed and this constant must agree exactly, or every correction
// fails with "no internal operation type", which reads as a data problem rather than a typo.
const StockCorrectionOperationTypeCode = "INV_CORRECTION"

const (
	StockOperationTypeSchemaName = "inventory_stock_operation_type"

	StockOperationTypeFieldId                           = basemodel.FieldId
	StockOperationTypeFieldCode                         = "code"
	StockOperationTypeFieldName                         = "name"
	StockOperationTypeFieldOperationCode                = "operation_code"
	StockOperationTypeFieldReservationMethod            = "reservation_method"
	StockOperationTypeFieldReserveBeforeDays            = "reserve_before_days"
	StockOperationTypeFieldBackorderPolicy              = "backorder_policy"
	StockOperationTypeFieldShippingPolicy               = "shipping_policy"
	StockOperationTypeFieldDefaultSourceLocationId      = "default_source_location_id"
	StockOperationTypeFieldDefaultDestinationLocationId = "default_destination_location_id"
	StockOperationTypeFieldOrgId                        = "org_id"

	StockOperationTypeEdgeDefaultSourceLocation      = "default_source_location"
	StockOperationTypeEdgeDefaultDestinationLocation = "default_destination_location"
)

//go:embed stock_operation_type.json
var stockOperationTypeSchemaJson string

func StockOperationTypeSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(stockOperationTypeSchemaJson)
}

type StockOperationType struct {
	basemodel.DynamicModelBase
}

func NewStockOperationType() *StockOperationType {
	return &StockOperationType{basemodel.NewDynamicModel()}
}

func NewStockOperationTypeFrom(src dmodel.DynamicFields) *StockOperationType {
	return &StockOperationType{basemodel.NewDynamicModel(src)}
}

func (this StockOperationType) GetCode() *string {
	return this.GetFieldData().GetString(StockOperationTypeFieldCode)
}

func (this *StockOperationType) SetCode(v *string) {
	this.GetFieldData().SetString(StockOperationTypeFieldCode, v)
}

func (this StockOperationType) GetName() *model.LangJson {
	return this.GetFieldData().GetLangJson(StockOperationTypeFieldName)
}

func (this *StockOperationType) SetName(v *model.LangJson) {
	this.GetFieldData().SetLangJson(StockOperationTypeFieldName, v)
}

func (this StockOperationType) GetOperationCode() *string {
	return this.GetFieldData().GetString(StockOperationTypeFieldOperationCode)
}

func (this *StockOperationType) SetOperationCode(v *string) {
	this.GetFieldData().SetString(StockOperationTypeFieldOperationCode, v)
}

func (this StockOperationType) GetReservationMethod() *string {
	return this.GetFieldData().GetString(StockOperationTypeFieldReservationMethod)
}

func (this *StockOperationType) SetReservationMethod(v *string) {
	this.GetFieldData().SetString(StockOperationTypeFieldReservationMethod, v)
}

func (this StockOperationType) GetBackorderPolicy() *string {
	return this.GetFieldData().GetString(StockOperationTypeFieldBackorderPolicy)
}

func (this *StockOperationType) SetBackorderPolicy(v *string) {
	this.GetFieldData().SetString(StockOperationTypeFieldBackorderPolicy, v)
}

func (this StockOperationType) GetShippingPolicy() *string {
	return this.GetFieldData().GetString(StockOperationTypeFieldShippingPolicy)
}

func (this *StockOperationType) SetShippingPolicy(v *string) {
	this.GetFieldData().SetString(StockOperationTypeFieldShippingPolicy, v)
}

func (this StockOperationType) GetDefaultSourceLocationId() *model.Id {
	return this.GetFieldData().GetModelId(StockOperationTypeFieldDefaultSourceLocationId)
}

func (this *StockOperationType) SetDefaultSourceLocationId(v *model.Id) {
	this.GetFieldData().SetModelId(StockOperationTypeFieldDefaultSourceLocationId, v)
}

func (this StockOperationType) GetDefaultDestinationLocationId() *model.Id {
	return this.GetFieldData().GetModelId(StockOperationTypeFieldDefaultDestinationLocationId)
}

func (this *StockOperationType) SetDefaultDestinationLocationId(v *model.Id) {
	this.GetFieldData().SetModelId(StockOperationTypeFieldDefaultDestinationLocationId, v)
}

func (this StockOperationType) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(StockOperationTypeFieldOrgId)
}

func (this *StockOperationType) SetOrgId(v *model.Id) {
	this.GetFieldData().SetModelId(StockOperationTypeFieldOrgId, v)
}
