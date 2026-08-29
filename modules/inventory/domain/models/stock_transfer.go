package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	StockTransferSchemaName = "inventory_stock_transfer"

	StockTransferFieldId                    = basemodel.FieldId
	StockTransferFieldTransferNumber        = "transfer_number"
	StockTransferFieldOperationTypeId       = "operation_type_id"
	StockTransferFieldOperationCode         = "operation_code"
	StockTransferFieldOriginReference       = "origin_reference"
	StockTransferFieldSourceLocationId      = "source_location_id"
	StockTransferFieldDestinationLocationId = "destination_location_id"
	StockTransferFieldStatus                = "status"
	StockTransferFieldPriority              = "priority"
	StockTransferFieldReservationMethod     = "reservation_method"
	StockTransferFieldBackorderPolicy       = "backorder_policy"
	StockTransferFieldShippingPolicy        = "shipping_policy"
	StockTransferFieldScheduledAt           = "scheduled_at"
	StockTransferFieldDeadlineAt            = "deadline_at"
	StockTransferFieldCompletedAt           = "completed_at"
	StockTransferFieldBackorderOfId         = "backorder_of_id"
	StockTransferFieldReturnOfId            = "return_of_id"
	StockTransferFieldChainGroupId          = "chain_group_id"
	StockTransferFieldIdempotencyKey        = "idempotency_key"
	StockTransferFieldNote                  = "note"
	StockTransferFieldOrgId                 = "org_id"

	StockTransferEdgeOperationType       = "operation_type"
	StockTransferEdgeSourceLocation      = "source_location"
	StockTransferEdgeDestinationLocation = "destination_location"
	StockTransferEdgeBackorderOf         = "backorder_of"
)

// Stock transfer lifecycle states, derived from the transfer's moves rather than set by a client.
// The transition table is in domain/services/stock_transfer_states.go.
const (
	StockTransferStatusDraft     = "draft"
	StockTransferStatusWaiting   = "waiting"
	StockTransferStatusConfirmed = "confirmed"
	StockTransferStatusReady     = "ready"
	StockTransferStatusDone      = "done"
	StockTransferStatusCancelled = "cancelled"
)

// operation_code, reservation_method, backorder_policy and shipping_policy are snapshots of the
// operation type's fields and share its constants in stock_operation_type.go; a second set declared
// here would drift apart silently.

//go:embed stock_transfer.json
var stockTransferSchemaJson string

func StockTransferSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(stockTransferSchemaJson)
}

// StockTransfer is the header of a stock transaction. It owns no quantities, which live on its
// moves, and carries its own copies of the operation type's policies so reconfiguring the type
// cannot reinterpret a transfer already created.
type StockTransfer struct {
	basemodel.DynamicModelBase
}

func NewStockTransfer() *StockTransfer {
	return &StockTransfer{basemodel.NewDynamicModel()}
}

func NewStockTransferFrom(src dmodel.DynamicFields) *StockTransfer {
	return &StockTransfer{basemodel.NewDynamicModel(src)}
}

func (this StockTransfer) GetTransferNumber() *string {
	return this.GetFieldData().GetString(StockTransferFieldTransferNumber)
}

func (this *StockTransfer) SetTransferNumber(v *string) {
	this.GetFieldData().SetString(StockTransferFieldTransferNumber, v)
}

func (this StockTransfer) GetOperationTypeId() *model.Id {
	return this.GetFieldData().GetModelId(StockTransferFieldOperationTypeId)
}

func (this *StockTransfer) SetOperationTypeId(v *model.Id) {
	this.GetFieldData().SetModelId(StockTransferFieldOperationTypeId, v)
}

func (this StockTransfer) GetOperationCode() *string {
	return this.GetFieldData().GetString(StockTransferFieldOperationCode)
}

func (this *StockTransfer) SetOperationCode(v *string) {
	this.GetFieldData().SetString(StockTransferFieldOperationCode, v)
}

func (this StockTransfer) GetSourceLocationId() *model.Id {
	return this.GetFieldData().GetModelId(StockTransferFieldSourceLocationId)
}

func (this *StockTransfer) SetSourceLocationId(v *model.Id) {
	this.GetFieldData().SetModelId(StockTransferFieldSourceLocationId, v)
}

func (this StockTransfer) GetDestinationLocationId() *model.Id {
	return this.GetFieldData().GetModelId(StockTransferFieldDestinationLocationId)
}

func (this *StockTransfer) SetDestinationLocationId(v *model.Id) {
	this.GetFieldData().SetModelId(StockTransferFieldDestinationLocationId, v)
}

func (this StockTransfer) GetStatus() *string {
	return this.GetFieldData().GetString(StockTransferFieldStatus)
}

func (this *StockTransfer) SetStatus(v *string) {
	this.GetFieldData().SetString(StockTransferFieldStatus, v)
}

func (this StockTransfer) GetReservationMethod() *string {
	return this.GetFieldData().GetString(StockTransferFieldReservationMethod)
}

func (this *StockTransfer) SetReservationMethod(v *string) {
	this.GetFieldData().SetString(StockTransferFieldReservationMethod, v)
}

func (this StockTransfer) GetBackorderPolicy() *string {
	return this.GetFieldData().GetString(StockTransferFieldBackorderPolicy)
}

func (this *StockTransfer) SetBackorderPolicy(v *string) {
	this.GetFieldData().SetString(StockTransferFieldBackorderPolicy, v)
}

func (this StockTransfer) GetShippingPolicy() *string {
	return this.GetFieldData().GetString(StockTransferFieldShippingPolicy)
}

func (this *StockTransfer) SetShippingPolicy(v *string) {
	this.GetFieldData().SetString(StockTransferFieldShippingPolicy, v)
}

func (this StockTransfer) GetBackorderOfId() *model.Id {
	return this.GetFieldData().GetModelId(StockTransferFieldBackorderOfId)
}

func (this *StockTransfer) SetBackorderOfId(v *model.Id) {
	this.GetFieldData().SetModelId(StockTransferFieldBackorderOfId, v)
}

func (this StockTransfer) GetIdempotencyKey() *string {
	return this.GetFieldData().GetString(StockTransferFieldIdempotencyKey)
}

func (this *StockTransfer) SetIdempotencyKey(v *string) {
	this.GetFieldData().SetString(StockTransferFieldIdempotencyKey, v)
}

func (this StockTransfer) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(StockTransferFieldOrgId)
}

func (this *StockTransfer) SetOrgId(v *model.Id) {
	this.GetFieldData().SetModelId(StockTransferFieldOrgId, v)
}
