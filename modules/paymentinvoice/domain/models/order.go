package models

import (
	_ "embed"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	OrderSchemaName = "paymentinvoice_order"

	OrderFieldId              = basemodel.FieldId
	OrderFieldOrderId         = "order_id"
	OrderFieldOrderCode       = "order_code"
	OrderFieldSource          = "source"
	OrderFieldStatus          = "status"
	OrderFieldAmount          = "amount"
	OrderFieldRefundAmount    = "refund_amount"
	OrderFieldCurrencyId      = "currency_id"
	OrderFieldPaymentMethodId = "payment_method_id"
	OrderFieldContent         = "content"
	OrderFieldReturnUrl       = "return_url"
	OrderFieldLastSyncStatus  = "last_sync_status"
	OrderFieldSyncLogs        = "sync_logs"
	OrderFieldMetadata        = "metadata"
	OrderFieldOrgId           = "org_id"
)

// Keys an adapter may require inside an order's metadata. Each adapter owns the keys it declares;
// they are listed here only so that the strings are not spelled twice across the module.
const (
	// OrderMetaPosId names the card terminal a prompt is pushed to. Required by the mPOS adapter
	// and meaningless to the others, which is the whole reason these live in a map.
	OrderMetaPosId = "pos_id"

	// OrderMetaCreateResponse and OrderMetaRefundResponse hold the gateway's own replies, kept
	// verbatim as the evidence for what was asked and answered.
	OrderMetaCreateResponse = "create_response"
	OrderMetaRefundResponse = "refund_response"
)

// Order statuses. The gateway callbacks, the refund action and the watchdog all move an order
// between these, and nothing else may.
const (
	OrderStatusPending        = "pending"
	OrderStatusProcessing     = "processing"
	OrderStatusPaymentSuccess = "payment_success"
	OrderStatusPaymentFailed  = "payment_failed"
	OrderStatusCanceled       = "canceled"
	OrderStatusRefundSuccess  = "refund_success"
	OrderStatusRefundFailed   = "refund_failed"
	OrderStatusExpired        = "expired"
)

// How the most recent attempt to notify the ordering system ended.
const (
	SyncStatusSuccess = "success"
	SyncStatusFailure = "failure"
)

//go:embed order.json
var orderSchemaJson string

func OrderSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(orderSchemaJson)
}

// Order is a request to collect money through a payment gateway, and the record of what became
// of it. Table: paymentinvoice_orders.
//
// Two of its columns are contracts with parties outside this codebase and may not be reshaped to
// suit us: order_code is the key every gateway callback arrives under, and order_id is the
// identifier quoted to the ordering system and to support. Both are generated rather than
// supplied, so that their parts always agree with the record they describe.
//
// amount and refund_amount are numeric, never an integer count of minor units: how many minor
// units a currency has is a property of the currency — VND has none, USD two — so that
// representation would push a per-currency scale into everything that reads one.
//
// currency_id and payment_method_id are plain ids. currency_id has no foreign key because the
// currency belongs to Essential, and a foreign key across a module boundary would make this
// module's schema depend on another module's table.
//
// metadata carries whatever the paying method needs — a terminal id for a card reader, nothing
// for a wallet — alongside each gateway's own replies, kept verbatim as the evidence when a
// payment is disputed. A column per gateway would be dead weight on every order paid another way,
// and adding a gateway would mean a migration.
type Order struct {
	basemodel.DynamicModelBase
}

func NewOrder() *Order {
	return &Order{basemodel.NewDynamicModel()}
}

func NewOrderFrom(src dmodel.DynamicFields) *Order {
	return &Order{basemodel.NewDynamicModel(src)}
}

func (this Order) GetOrderId() *string {
	return this.GetFieldData().GetString(OrderFieldOrderId)
}

func (this *Order) SetOrderId(v *string) {
	this.GetFieldData().SetString(OrderFieldOrderId, v)
}

func (this Order) GetOrderCode() *string {
	return this.GetFieldData().GetString(OrderFieldOrderCode)
}

func (this *Order) SetOrderCode(v *string) {
	this.GetFieldData().SetString(OrderFieldOrderCode, v)
}

func (this Order) GetSource() *string {
	return this.GetFieldData().GetString(OrderFieldSource)
}

func (this *Order) SetSource(v *string) {
	this.GetFieldData().SetString(OrderFieldSource, v)
}

func (this Order) GetStatus() *string {
	return this.GetFieldData().GetString(OrderFieldStatus)
}

func (this *Order) SetStatus(v *string) {
	this.GetFieldData().SetString(OrderFieldStatus, v)
}

func (this Order) GetAmount() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(OrderFieldAmount)
}

func (this *Order) SetAmount(v *decimal.Decimal) {
	this.GetFieldData().SetDecimal(OrderFieldAmount, v)
}

func (this Order) GetRefundAmount() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(OrderFieldRefundAmount)
}

func (this *Order) SetRefundAmount(v *decimal.Decimal) {
	this.GetFieldData().SetDecimal(OrderFieldRefundAmount, v)
}

func (this Order) GetCurrencyId() *model.Id {
	return this.GetFieldData().GetModelId(OrderFieldCurrencyId)
}

func (this *Order) SetCurrencyId(v *model.Id) {
	this.GetFieldData().SetModelId(OrderFieldCurrencyId, v)
}

func (this Order) GetPaymentMethodId() *model.Id {
	return this.GetFieldData().GetModelId(OrderFieldPaymentMethodId)
}

func (this *Order) SetPaymentMethodId(v *model.Id) {
	this.GetFieldData().SetModelId(OrderFieldPaymentMethodId, v)
}

func (this Order) GetContent() *string {
	return this.GetFieldData().GetString(OrderFieldContent)
}

func (this *Order) SetContent(v *string) {
	this.GetFieldData().SetString(OrderFieldContent, v)
}

// GetMetadata returns the method-specific map described on OrderFieldMetadata. It is reached
// through GetAny because a jsonmap has no typed accessor; only the owning adapter interprets it.
func (this Order) GetMetadata() map[string]any {
	raw := this.GetFieldData().GetAny(OrderFieldMetadata)
	if raw == nil {
		return nil
	}
	metadata, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	return metadata
}

func (this *Order) SetMetadata(v map[string]any) {
	this.GetFieldData().SetAny(OrderFieldMetadata, v)
}

func (this Order) GetReturnUrl() *string {
	return this.GetFieldData().GetString(OrderFieldReturnUrl)
}

func (this *Order) SetReturnUrl(v *string) {
	this.GetFieldData().SetString(OrderFieldReturnUrl, v)
}

func (this Order) GetLastSyncStatus() *string {
	return this.GetFieldData().GetString(OrderFieldLastSyncStatus)
}

func (this *Order) SetLastSyncStatus(v *string) {
	this.GetFieldData().SetString(OrderFieldLastSyncStatus, v)
}

// GetSyncLogs returns the record of attempts to notify the ordering system of this order's
// outcome. Like the metadata it is a jsonmap, so it is reached through GetAny; the entries live
// under a key inside it because the dynamic-model system has no array-typed field.
func (this Order) GetSyncLogs() map[string]any {
	raw := this.GetFieldData().GetAny(OrderFieldSyncLogs)
	if raw == nil {
		return nil
	}
	logs, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	return logs
}

func (this *Order) SetSyncLogs(v map[string]any) {
	this.GetFieldData().SetAny(OrderFieldSyncLogs, v)
}

func (this Order) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(OrderFieldOrgId)
}

func (this *Order) SetOrgId(v *model.Id) {
	this.GetFieldData().SetModelId(OrderFieldOrgId, v)
}
