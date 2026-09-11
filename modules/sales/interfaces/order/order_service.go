// Package order is Sales' port for raising and settling an order from another module in-process.
//
// It exists because a kiosk cannot use the REST surface: it is already inside the same binary, and
// going out through HTTP to come back in would lose the request's transaction, its org scoping and
// its authenticated principal. What it can do is call this.
//
// The port is deliberately NARROW. It offers the three steps a selling device actually takes —
// create the order, confirm it, cancel it — and nothing else. Pricing, fulfillment resolution, stock
// reservation and refund policy all stay inside Sales, decided by the same code that decides them
// for every other channel. A caller that could reach further would eventually reach differently.
package order

import (
	"github.com/shopspring/decimal"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// SalesOrderExtService is what an in-process seller calls to put a sale on Sales' books.
type SalesOrderExtService interface {
	// CreateOrder raises a draft order. Idempotent by (channel, idempotency key): a caller retrying
	// after a timeout is handed the order it created the first time rather than a second one, which
	// is what makes it safe to call before taking money.
	CreateOrder(
		ctx corectx.Context, command CreateSalesOrderCommand,
	) (*CreateSalesOrderResult, error)

	// ConfirmOrder commits the sale: prices it, redeems any vouchers, and — for a kiosk method —
	// creates the fulfillment and reserves the stock BEFORE the order is frozen. A reservation that
	// cannot be met refuses here, with the order still a draft and nothing charged.
	ConfirmOrder(
		ctx corectx.Context, command SalesOrderCommand,
	) (*ConfirmSalesOrderResult, error)

	// CancelOrder calls the sale off and releases whatever stock it was holding. Refused for an
	// order that has been paid — that needs a refund, which is a different operation with different
	// consequences.
	CancelOrder(
		ctx corectx.Context, command CancelSalesOrderCommand,
	) (*CancelSalesOrderResult, error)

	// CreateAttempt records that an executor is about to be asked to hand goods over, and returns
	// the correlation id it must echo back.
	//
	// Called BEFORE the device is told anything. A machine that takes a command and then goes silent
	// then leaves a pending attempt naming what it was asked for and when, which an operator can
	// find; telling it first and recording afterwards would lose exactly those cases.
	CreateAttempt(
		ctx corectx.Context, command CreateAttemptCommand,
	) (*CreateAttemptResult, error)

	// ReportAttemptResult records what the executor actually did.
	//
	// Safe to call more than once with the same event id: the same id and the same content replays
	// the stored answer rather than crediting the goods twice, while the same id with different
	// content is refused as a contradiction. That is what lets a result arrive over an at-most-once
	// broker AND a REST backstop without double-counting.
	ReportAttemptResult(
		ctx corectx.Context, command ReportAttemptResultCommand,
	) (*ReportAttemptResultResult, error)

	// ViewFulfillment reports what a delivery still owes, per item.
	//
	// A caller about to command a dispense needs this rather than its own record of the sale: Sales
	// knows what may be attempted NOW — still owed, minus anything with a refund in flight — and the
	// caller does not. Asking for the original quantities would re-attempt goods already handed over
	// on an earlier try, or goods being refunded underneath this one.
	ViewFulfillment(
		ctx corectx.Context, command ViewFulfillmentCommand,
	) (*ViewFulfillmentResult, error)
}

type ViewFulfillmentCommand struct {
	FulfillmentId string `json:"fulfillment_id"`
}

// FulfillmentItemView is one product of a delivery and the three quantities that matter to a caller
// deciding what to attempt.
type FulfillmentItemView struct {
	FulfillmentItemId string          `json:"fulfillment_item_id"`
	ProductVariantId  string          `json:"product_variant_id"`
	RemainingQty      decimal.Decimal `json:"remaining_qty"`

	// PendingRefundQty is owed but NOT attemptable: the money may be about to go back, and
	// dispensing against it would hand goods over while paying for them in reverse.
	PendingRefundQty decimal.Decimal `json:"pending_refund_qty"`

	// FulfillableQty is what an attempt may ask for. This is the number to use, not RemainingQty.
	FulfillableQty decimal.Decimal `json:"fulfillable_qty"`

	// SourceLocationId is the exact place this item's goods are held — a vending slot, where the
	// target has addressable positions. Empty means the fulfillment's single target location holds
	// them, which is every target that does not.
	SourceLocationId string `json:"source_location_id,omitempty"`

	// InventorySourceId is the key the hold for THIS item is found by in Inventory. An executor
	// reporting a physical result quotes it so the right hold is consumed or released: with stock
	// held at several slots there is no longer one hold per fulfillment to guess at.
	InventorySourceId string `json:"inventory_source_id,omitempty"`
}

type FulfillmentView struct {
	FulfillmentId     string                `json:"fulfillment_id"`
	FulfillmentStatus string                `json:"fulfillment_status"`
	TargetOutletId    string                `json:"target_outlet_id"`
	Items             []FulfillmentItemView `json:"items"`
}

type ViewFulfillmentResult struct {
	ClientErrors ft.ClientErrors `json:"client_errors,omitempty"`
	Data         FulfillmentView `json:"data"`
	HasData      bool            `json:"has_data"`
}

// CreateAttemptCommand asks for one try at a fulfillment.
type CreateAttemptCommand struct {
	FulfillmentId string `json:"fulfillment_id"`

	// ExecutorOutletId is the sales point being asked. Named explicitly rather than read from the
	// fulfillment so a request aimed at the wrong machine is REFUSED rather than silently
	// redirected to the right one.
	ExecutorOutletId string `json:"executor_outlet_id"`

	Items []AttemptItemCommand `json:"items"`
}

// AttemptItemCommand is one product and how much of it to try.
type AttemptItemCommand struct {
	FulfillmentItemId string          `json:"fulfillment_item_id"`
	Quantity          decimal.Decimal `json:"quantity"`
}

type AttemptData struct {
	AttemptId string `json:"attempt_id"`
	AttemptNo int32  `json:"attempt_no"`

	// ExternalCorrelationId is what the executor must echo back, so a reply can be matched to the
	// try that caused it. A device that only knows an order code carries this in its payload.
	ExternalCorrelationId string `json:"external_correlation_id"`
}

type CreateAttemptResult struct {
	ClientErrors ft.ClientErrors `json:"client_errors,omitempty"`
	Data         AttemptData     `json:"data"`
	HasData      bool            `json:"has_data"`
}

// ReportAttemptResultCommand is one executor's report about one try.
type ReportAttemptResultCommand struct {
	AttemptId string `json:"attempt_id"`

	// ResultEventId is the REPORTER's id for this event, and the idempotency key. Required: without
	// one a redelivery is indistinguishable from a second dispense.
	ResultEventId string `json:"result_event_id"`

	ExternalCorrelationId string `json:"external_correlation_id"`

	// InventoryResultRef proves Inventory already applied the physical consequence. Sales records a
	// dispense only after the stock ledger has, or it would report goods gone that Inventory still
	// believes are on the shelf.
	InventoryResultRef string `json:"inventory_result_ref"`

	ExecutorOutletId string                     `json:"executor_outlet_id"`
	Items            []AttemptResultItemCommand `json:"items"`
}

// AttemptResultItemCommand is what happened to one product. Dispensed plus failed must equal what
// was attempted, or the report is describing something other than the try it names.
type AttemptResultItemCommand struct {
	FulfillmentItemId string          `json:"fulfillment_item_id"`
	DispensedQty      decimal.Decimal `json:"dispensed_qty"`
	FailedQty         decimal.Decimal `json:"failed_qty"`
	FailureCode       string          `json:"failure_code"`
	FailureMessage    string          `json:"failure_message"`
}

type AttemptResultData struct {
	AttemptId         string `json:"attempt_id"`
	AttemptStatus     string `json:"attempt_status"`
	FulfillmentId     string `json:"fulfillment_id"`
	FulfillmentStatus string `json:"fulfillment_status"`

	// AlreadyApplied marks a replay: the answer is the stored one and nothing was written again.
	AlreadyApplied bool `json:"already_applied"`

	// RefundId names a refund the failure policy raised, if any. Creation is not success — the
	// quantity is not refunded until settlement says so, which RefundStatus carries.
	RefundId     string `json:"refund_id,omitempty"`
	RefundStatus string `json:"refund_status,omitempty"`
}

type ReportAttemptResultResult struct {
	ClientErrors ft.ClientErrors   `json:"client_errors,omitempty"`
	Data         AttemptResultData `json:"data"`
	HasData      bool              `json:"has_data"`
}

// CreateSalesOrderCommand names a sale in the caller's own terms.
//
// It carries no prices and no channel: the sales point decides the channel, and Sales' own pricelists
// decide the price. A caller that could send either would be able to sell at a price the business
// never set, on a channel it does not belong to.
type CreateSalesOrderCommand struct {
	// SalesPointId is where the sale happens, and the only thing that decides which channel it lands
	// on. For a kiosk this is the sales point the machine registered as.
	SalesPointId string `json:"sales_point_id"`

	CurrencyCode string `json:"currency_code"`

	// IdempotencyKey is the caller's own id for this sale — a kiosk transaction code. Unique per
	// channel; absent means the caller accepts that a retry creates a second order.
	IdempotencyKey string `json:"idempotency_key"`

	// ExternalReference ties the order back to the caller's own record for an operator tracing it.
	ExternalReference string `json:"external_reference"`

	Lines []CreateSalesOrderLine `json:"lines"`

	// TargetOutletId names where the goods should be handed over, when that differs from where the
	// sale was made. Empty for a kiosk selling to the customer standing in front of it, where the
	// method resolves the target to the selling point itself.
	TargetOutletId string `json:"target_outlet_id"`

	// EstimatedTotalPrice is what the caller's own calculation came to. It does not breach the
	// no-prices rule above: it is recorded for reconciliation and never charged, so a caller sending
	// it cannot sell at a price the business did not set.
	EstimatedTotalPrice *decimal.Decimal `json:"estimated_total_price,omitempty"`
}

// CreateSalesOrderLine is one product and how much of it. No authoritative price: see the command
// above.
type CreateSalesOrderLine struct {
	ProductVariantId string          `json:"product_variant_id"`
	UomId            string          `json:"uom_id"`
	Quantity         decimal.Decimal `json:"quantity"`

	// EstimatedPrice is the unit price the caller displayed to the customer. Recorded so a device
	// showing a stale price can be found; Sales prices the line regardless of what it says.
	EstimatedPrice *decimal.Decimal `json:"estimated_price,omitempty"`
}

type SalesOrderData struct {
	SalesOrderId   string `json:"sales_order_id"`
	OrderNumber    string `json:"order_number"`
	SalesChannelId string `json:"sales_channel_id"`

	// AlreadyExisted tells a caller its retry was recognised rather than a new order created.
	AlreadyExisted bool `json:"already_existed"`
}

type CreateSalesOrderResult struct {
	ClientErrors ft.ClientErrors `json:"client_errors,omitempty"`
	Data         SalesOrderData  `json:"data"`
	HasData      bool            `json:"has_data"`
}

type SalesOrderCommand struct {
	SalesOrderId string `json:"sales_order_id"`
}

type CancelSalesOrderCommand struct {
	SalesOrderId string `json:"sales_order_id"`
	Reason       string `json:"reason"`
}

// ConfirmedOrderData reports what confirming produced, including the delivery half of a kiosk sale.
type ConfirmedOrderData struct {
	SalesOrderId string `json:"sales_order_id"`
	Status       string `json:"status"`

	// FulfillmentId is the delivery this sale created, for a method that dispenses. Empty for a sale
	// that hands nothing over — a service, or an order shipped by another channel entirely.
	FulfillmentId string `json:"fulfillment_id"`

	// FulfillmentStatus says whether the stock is actually held. A caller must read this rather than
	// assume: an order can confirm with its reservation still pending when no inventory port is
	// bound, and telling a customer to expect goods on that basis would be wrong.
	FulfillmentStatus string `json:"fulfillment_status"`

	// InitialBillId is the bill the confirmation raised, and what a payment settles against. Always
	// set on success - a confirmed order without one is a data-integrity fault, not a case to handle.
	InitialBillId string `json:"initial_bill_id"`

	// AlreadyConfirmed says the order was already confirmed and this call changed nothing. The bill
	// is the original one, which is what makes a retry after a lost response safe.
	AlreadyConfirmed bool `json:"already_confirmed"`

	// Order and InitialBill are the whole records, for a caller whose next act is to ask for money
	// and would otherwise need a second round trip to learn how much. Nil when the read-back failed;
	// the ids above are the contract.
	Order       map[string]any `json:"order,omitempty"`
	InitialBill map[string]any `json:"initial_bill,omitempty"`

	// Pending names the steps confirm did not complete, so a caller does not read a success as
	// "everything is done".
	Pending []string `json:"pending"`
}

type ConfirmSalesOrderResult struct {
	ClientErrors ft.ClientErrors    `json:"client_errors,omitempty"`
	Data         ConfirmedOrderData `json:"data"`
	HasData      bool               `json:"has_data"`
}

type CancelledOrderData struct {
	SalesOrderId string `json:"sales_order_id"`
	Status       string `json:"status"`

	// ReleasedFulfillmentIds are the deliveries whose held stock went back to the sellable pool.
	ReleasedFulfillmentIds []string `json:"released_fulfillment_ids"`
}

type CancelSalesOrderResult struct {
	ClientErrors ft.ClientErrors    `json:"client_errors,omitempty"`
	Data         CancelledOrderData `json:"data"`
	HasData      bool               `json:"has_data"`
}
