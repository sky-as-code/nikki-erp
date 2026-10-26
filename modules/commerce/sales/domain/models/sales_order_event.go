package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	SalesOrderEventSchemaName = "sales_order_event"

	SalesOrderEventFieldId           = "id"
	SalesOrderEventFieldOrgId        = "org_id"
	SalesOrderEventFieldSalesOrderId = "sales_order_id"
	SalesOrderEventFieldEntityType   = "entity_type"
	SalesOrderEventFieldEntityId     = "entity_id"
	SalesOrderEventFieldAction       = "action"
	SalesOrderEventFieldActorId      = "actor_id"
	SalesOrderEventFieldFromStatus   = "from_status"
	SalesOrderEventFieldToStatus     = "to_status"
	SalesOrderEventFieldReason       = "reason"
	SalesOrderEventFieldMetadata     = "metadata"
)

// The recorded actions. Constants rather than free strings at the call sites, so a typo is a compile
// error instead of an event nobody can find by filtering.
const (
	SalesOrderActionCreate         = "create"
	SalesOrderActionConfirm        = "confirm"
	SalesOrderActionCancel         = "cancel"
	SalesOrderActionComplete       = "complete"
	SalesOrderActionAddLine        = "add_line"
	SalesOrderActionUpdateLine     = "update_line"
	SalesOrderActionRemoveLine     = "remove_line"
	SalesOrderActionManualDiscount = "manual_discount"
	SalesOrderActionPriceOverride  = "price_override"
	SalesOrderActionSplitBill      = "split_bill"
	SalesOrderActionMergeBill      = "merge_bill"
	SalesOrderActionRecordPayment  = "record_payment"
	SalesOrderActionRefund         = "refund"
	SalesOrderActionProcessReturn  = "process_return"
	SalesOrderActionRequestInvoice = "request_invoice"
	SalesOrderActionFulfill        = "fulfill"

	// SalesOrderActionConvertQuotation records that an order came from an accepted offer rather than
	// being raised directly, so an investigator can find what it was quoted at.
	SalesOrderActionConvertQuotation = "convert_quotation"

	// SalesOrderActionExpire records a draft that went stale rather than one somebody withdrew. Its
	// own action because sales_orders has no 'expired' status: an expired draft is stored as
	// cancelled, so only this tells the two apart afterwards.
	SalesOrderActionExpire = "expire"
)

//go:embed sales_order_event.json
var salesOrderEventSchemaJson string

func SalesOrderEventSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesOrderEventSchemaJson)
}

// SalesOrderEvent is one thing that happened to a sale: the document trail, not the pricing trail
// (sales_order_adjustment). An adjustment explains a number and is replaced when the order is
// repriced; an event records that something occurred and is never replaced. The schema extends
// auditable_readonly_model with no archivable mixin: an editable audit row would be worthless as
// evidence.
type SalesOrderEvent struct {
	basemodel.DynamicModelBase
}

func NewSalesOrderEvent() *SalesOrderEvent {
	return &SalesOrderEvent{basemodel.NewDynamicModel()}
}

func NewSalesOrderEventFrom(src dmodel.DynamicFields) *SalesOrderEvent {
	return &SalesOrderEvent{basemodel.NewDynamicModel(src)}
}

func (this SalesOrderEvent) GetId() *model.Id {
	return this.GetFieldData().GetModelId(basemodel.FieldId)
}

func (this *SalesOrderEvent) SetId(id *model.Id) {
	this.GetFieldData().SetModelId(basemodel.FieldId, id)
}

func (this SalesOrderEvent) GetSalesOrderId() *model.Id {
	return this.GetFieldData().GetModelId(SalesOrderEventFieldSalesOrderId)
}

func (this *SalesOrderEvent) SetSalesOrderId(id *model.Id) {
	this.GetFieldData().SetModelId(SalesOrderEventFieldSalesOrderId, id)
}

func (this SalesOrderEvent) GetEntityType() *string {
	return this.GetFieldData().GetString(SalesOrderEventFieldEntityType)
}

func (this *SalesOrderEvent) SetEntityType(entityType *string) {
	this.GetFieldData().SetString(SalesOrderEventFieldEntityType, entityType)
}

func (this SalesOrderEvent) GetEntityId() *model.Id {
	return this.GetFieldData().GetModelId(SalesOrderEventFieldEntityId)
}

func (this *SalesOrderEvent) SetEntityId(id *model.Id) {
	this.GetFieldData().SetModelId(SalesOrderEventFieldEntityId, id)
}

func (this SalesOrderEvent) GetAction() *string {
	return this.GetFieldData().GetString(SalesOrderEventFieldAction)
}

func (this *SalesOrderEvent) SetAction(action *string) {
	this.GetFieldData().SetString(SalesOrderEventFieldAction, action)
}

func (this SalesOrderEvent) GetActorId() *model.Id {
	return this.GetFieldData().GetModelId(SalesOrderEventFieldActorId)
}

func (this *SalesOrderEvent) SetActorId(id *model.Id) {
	this.GetFieldData().SetModelId(SalesOrderEventFieldActorId, id)
}

func (this SalesOrderEvent) GetFromStatus() *string {
	return this.GetFieldData().GetString(SalesOrderEventFieldFromStatus)
}

func (this *SalesOrderEvent) SetFromStatus(status *string) {
	this.GetFieldData().SetString(SalesOrderEventFieldFromStatus, status)
}

func (this SalesOrderEvent) GetToStatus() *string {
	return this.GetFieldData().GetString(SalesOrderEventFieldToStatus)
}

func (this *SalesOrderEvent) SetToStatus(status *string) {
	this.GetFieldData().SetString(SalesOrderEventFieldToStatus, status)
}

func (this SalesOrderEvent) GetReason() *string {
	return this.GetFieldData().GetString(SalesOrderEventFieldReason)
}

func (this *SalesOrderEvent) SetReason(reason *string) {
	this.GetFieldData().SetString(SalesOrderEventFieldReason, reason)
}

func (this SalesOrderEvent) GetMetadata() map[string]any {
	value := this.GetFieldData().GetAny(SalesOrderEventFieldMetadata)
	if typed, ok := value.(map[string]any); ok {
		return typed
	}
	return nil
}

func (this *SalesOrderEvent) SetMetadata(metadata map[string]any) {
	this.GetFieldData().SetAny(SalesOrderEventFieldMetadata, metadata)
}
