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

// The Commerce org settings that name an operation type by id. They are declared in the Commerce
// org settings schema (sales/domain/settings/org_settings.json); Inventory seeds them for a new org
// from the types it creates, and a stock correction records its movements under the correction type.
const (
	OrgSettingCorrectionOperationTypeId = "correction_operation_type_id"
	OrgSettingOutgoingOperationTypeId   = "outgoing_operation_type_id"
	OrgSettingIncomingOperationTypeId   = "incoming_operation_type_id"
)

const (
	StockOperationTypeSchemaName = "inventory_stock_operation_type"

	StockOperationTypeFieldId                           = basemodel.FieldId
	StockOperationTypeFieldName                         = "name"
	StockOperationTypeFieldOperationCode                = "operation_code"
	StockOperationTypeFieldReservationMethod            = "reservation_method"
	StockOperationTypeFieldReserveBeforeDays            = "reserve_before_days"
	StockOperationTypeFieldBackorderPolicy              = "backorder_policy"
	StockOperationTypeFieldShippingPolicy               = "shipping_policy"
	StockOperationTypeFieldDefaultSourceLocationId      = "default_source_location_id"
	StockOperationTypeFieldDefaultDestinationLocationId = "default_destination_location_id"
	StockOperationTypeFieldOrgId                        = "org_id"
	StockOperationTypeFieldSequencePrefix               = "sequence_prefix"
	StockOperationTypeFieldWarehouseId                  = "warehouse_id"

	StockOperationTypeEdgeDefaultSourceLocation      = "default_source_location"
	StockOperationTypeEdgeDefaultDestinationLocation = "default_destination_location"
	StockOperationTypeEdgeWarehouse                  = "warehouse"
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

func (this StockOperationType) GetSequencePrefix() *string {
	return this.GetFieldData().GetString(StockOperationTypeFieldSequencePrefix)
}

func (this StockOperationType) GetWarehouseId() *model.Id {
	return this.GetFieldData().GetModelId(StockOperationTypeFieldWarehouseId)
}
