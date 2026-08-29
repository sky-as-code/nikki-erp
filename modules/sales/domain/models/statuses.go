package models

// Status values of the lifecycle-bearing Sales resources. Declared here rather than inline so the
// schema JSON and the code reading it cannot drift; a typo in a comparison would otherwise be a
// condition silently never true. Stored lower-case, matching the JSON schemas.

// SalesChannelStatus is the business lifecycle of a sales channel, separate from is_archived (the
// system lifecycle): suspended means "not selling right now", archived means "no longer part of the
// catalogue of channels".
type SalesChannelStatus string

const (
	// SalesChannelStatusActive permits new sales points, new orders and integration requests.
	SalesChannelStatusActive = SalesChannelStatus("active")
	// SalesChannelStatusSuspended stops all three, while leaving reads, returns, refunds and
	// fiscal adjustments of existing transactions working.
	SalesChannelStatusSuspended = SalesChannelStatus("suspended")
)

// SalesPointStatus is the business lifecycle of a sales point.
type SalesPointStatus string

const (
	// SalesPointStatusActive permits new sales orders at this point.
	SalesPointStatusActive = SalesPointStatus("active")
	// SalesPointStatusSuspended stops new orders but keeps history, returns and refunds available.
	SalesPointStatusSuspended = SalesPointStatus("suspended")
)

// SalesOrderStatus is the document's own lifecycle. It is one of four independent status fields on
// a sales order, never collapsed into one: an order can be confirmed and fully paid but undelivered,
// or delivered and unpaid, or complete with its VAT invoice rejected.
type SalesOrderStatus string

const (
	// SalesOrderStatusDraft is a document still being built. Lines may be added, changed and
	// removed, and prices are recalculated on every change.
	SalesOrderStatusDraft = SalesOrderStatus("draft")

	// SalesOrderStatusConfirmed is a sale the business has committed to. The snapshot fields on
	// every line become immutable at this moment.
	SalesOrderStatusConfirmed = SalesOrderStatus("confirmed")

	// SalesOrderStatusProcessing is confirmed and part-way through fulfilment. An order needing no
	// fulfilment at all never enters it.
	SalesOrderStatusProcessing = SalesOrderStatus("processing")

	// SalesOrderStatusCompleted is paid and fulfilled. Terminal for this dimension only — returns,
	// refunds and invoicing still happen afterwards.
	SalesOrderStatusCompleted = SalesOrderStatus("completed")

	// SalesOrderStatusCancelled is a sale that will not happen. The record is kept as evidence of
	// what was attempted.
	SalesOrderStatusCancelled = SalesOrderStatus("cancelled")
)

// SalesOrderPaymentStatus is how much of the money has arrived. Derived from the sum of the order's
// payments rather than set directly, so it can never disagree with them.
type SalesOrderPaymentStatus string

const (
	SalesOrderPaymentStatusUnpaid        = SalesOrderPaymentStatus("unpaid")
	SalesOrderPaymentStatusPartiallyPaid = SalesOrderPaymentStatus("partially_paid")
	SalesOrderPaymentStatusPaid          = SalesOrderPaymentStatus("paid")

	// SalesOrderPaymentStatusOverpaid is a real state, not an error: a cash till takes what the
	// customer hands over. Whether change may be given back is the allow_cash_change policy setting.
	SalesOrderPaymentStatusOverpaid = SalesOrderPaymentStatus("overpaid")

	SalesOrderPaymentStatusRefunded          = SalesOrderPaymentStatus("refunded")
	SalesOrderPaymentStatusPartiallyRefunded = SalesOrderPaymentStatus("partially_refunded")
)

// SalesOrderFulfillmentStatus is how much of the goods have moved. Derived from the lines'
// fulfilled_quantity and returned_quantity totals, so it cannot contradict them.
type SalesOrderFulfillmentStatus string

const (
	SalesOrderFulfillmentStatusPending = SalesOrderFulfillmentStatus("pending")

	// SalesOrderFulfillmentStatusNotRequired is an order with nothing to hand over — every line is
	// a service, a fee or a non-stocked item. Fulfilled means goods moved; this means none were ever
	// owed. Both satisfy completion.
	SalesOrderFulfillmentStatusNotRequired = SalesOrderFulfillmentStatus("not_required")

	SalesOrderFulfillmentStatusPartiallyFulfilled = SalesOrderFulfillmentStatus("partially_fulfilled")
	SalesOrderFulfillmentStatusFulfilled          = SalesOrderFulfillmentStatus("fulfilled")
	SalesOrderFulfillmentStatusReturned           = SalesOrderFulfillmentStatus("returned")
	SalesOrderFulfillmentStatusPartiallyReturned  = SalesOrderFulfillmentStatus("partially_returned")
)

// SalesOrderInvoiceStatus is where the VAT invoice has got to. Separate from the other three
// because a tax authority can reject an invoice for a sale that is paid, delivered and complete,
// and that failure must be visible without making the sale itself look broken.
type SalesOrderInvoiceStatus string

const (
	SalesOrderInvoiceStatusNotRequested = SalesOrderInvoiceStatus("not_requested")
	SalesOrderInvoiceStatusRequested    = SalesOrderInvoiceStatus("requested")
	SalesOrderInvoiceStatusIssued       = SalesOrderInvoiceStatus("issued")
	SalesOrderInvoiceStatusFailed       = SalesOrderInvoiceStatus("failed")
	SalesOrderInvoiceStatusCancelled    = SalesOrderInvoiceStatus("cancelled")
)

// SalesOrderLineType is what kind of thing a line is.
type SalesOrderLineType string

const (
	// SalesOrderLineTypeProduct sells one variant.
	SalesOrderLineTypeProduct = SalesOrderLineType("product")

	// SalesOrderLineTypeCombo is the virtual parent of a bundle. Its real variants live in
	// sales_order_line_components, because Inventory fulfils real variants and never a virtual combo.
	SalesOrderLineTypeCombo = SalesOrderLineType("combo")

	// SalesOrderLineTypePromotionReward is a free item given by a promotion. A real line rather than
	// an adjustment, because Inventory must physically fulfil it and its VAT treatment is line-level.
	SalesOrderLineTypePromotionReward = SalesOrderLineType("promotion_reward")
)

// SalesOrderPricingSource is where a line's price came from. It lets the price-explanation API
// answer "why does this cost this" without replaying the whole engine.
type SalesOrderPricingSource string

const (
	SalesOrderPricingSourceCatalogue       = SalesOrderPricingSource("catalogue")
	SalesOrderPricingSourcePricelist       = SalesOrderPricingSource("pricelist")
	SalesOrderPricingSourceCombo           = SalesOrderPricingSource("combo")
	SalesOrderPricingSourcePromotionReward = SalesOrderPricingSource("promotion_reward")
	SalesOrderPricingSourceManualOverride  = SalesOrderPricingSource("manual_override")
)

// VoucherCodeStatus is whether a code may currently be applied. Expiry is deliberately absent: it
// is a function of valid_until and the current time, so making it a status would leave a code
// wrongly usable until some job ran.
type VoucherCodeStatus string

const (
	VoucherCodeStatusActive = VoucherCodeStatus("active")

	// VoucherCodeStatusDisabled is an operator's decision, and reversible; archiving is not.
	VoucherCodeStatusDisabled = VoucherCodeStatus("disabled")

	// VoucherCodeStatusExhausted is derived, not chosen: usage_count reached usage_limit. Set by the
	// redemption path, and a return that restores a use moves it back.
	VoucherCodeStatusExhausted = VoucherCodeStatus("exhausted")
)

// VoucherRedemptionStatus tracks one code's use on one order. A reservation taken on a draft is
// settled as 'redeemed' if the order confirms or 'released' if it never does; a return then undoes a
// redemption as 'reversed'. Released and reversed stay apart because a campaign report counts them
// differently - never really used versus used and given back.
type VoucherRedemptionStatus string

const (
	// VoucherRedemptionStatusReserved holds a use while an order is still a draft, stopping a second
	// customer taking the last use of a voucher already in someone's basket. A usage counter alone
	// could not, since a draft has not incremented it yet.
	VoucherRedemptionStatusReserved = VoucherRedemptionStatus("reserved")

	VoucherRedemptionStatusRedeemed = VoucherRedemptionStatus("redeemed")

	// VoucherRedemptionStatusReleased gives the hold back without a sale: the draft was cancelled
	// or expired.
	VoucherRedemptionStatusReleased = VoucherRedemptionStatus("released")

	// VoucherRedemptionStatusReversed gives a completed use back after a return. Whether a return
	// restores at all is the program's decision, not the redemption's.
	VoucherRedemptionStatusReversed = VoucherRedemptionStatus("reversed")
)

// SalesBillStatus is where a settlement unit stands. A bill is never a VAT invoice: a bill is how
// the money is collected, an invoice is the legal document, and one sale can need several of the
// first and one of the second - or the other way round.
type SalesBillStatus string

const (
	SalesBillStatusOpen = SalesBillStatus("open")

	// SalesBillStatusSettled means the money is fully in. The line allocations freeze at this point,
	// because they are what the payment was measured against.
	SalesBillStatusSettled = SalesBillStatus("settled")

	// SalesBillStatusCancelled marks a bill superseded by a split or a merge. The row stays because
	// the lineage relations point at it.
	SalesBillStatusCancelled = SalesBillStatus("cancelled")
)

// SalesBillRelationType says which operation produced a lineage row. Both types read
// source -> target, so only the type distinguishes a split from a merge.
type SalesBillRelationType string

const (
	SalesBillRelationSplitInto  = SalesBillRelationType("split_into")
	SalesBillRelationMergedInto = SalesBillRelationType("merged_into")
)

// SalesPaymentStatus is where one payment stands with its provider. Only `captured` counts toward
// settling a bill: an authorization is a hold the provider may still release, and treating it as
// money in would settle a bill against funds that never arrived.
type SalesPaymentStatus string

const (
	SalesPaymentStatusPending    = SalesPaymentStatus("pending")
	SalesPaymentStatusAuthorized = SalesPaymentStatus("authorized")
	SalesPaymentStatusCaptured   = SalesPaymentStatus("captured")
	SalesPaymentStatusFailed     = SalesPaymentStatus("failed")
	SalesPaymentStatusCancelled  = SalesPaymentStatus("cancelled")
)

// SalesFulfillmentRequestType is what Sales is asking Inventory to do: intent, never instruction.
// Inventory decides availability, warehouse, location and the movements. Sales never touches stock.
type SalesFulfillmentRequestType string

const (
	// SalesFulfillmentTypeReservation holds stock without moving it, at confirmation.
	SalesFulfillmentTypeReservation = SalesFulfillmentRequestType("reservation")

	// SalesFulfillmentTypeGoodsIssue is the movement that takes goods out.
	SalesFulfillmentTypeGoodsIssue = SalesFulfillmentRequestType("goods_issue")

	SalesFulfillmentTypeReturnReceipt = SalesFulfillmentRequestType("return_receipt")

	// SalesFulfillmentTypeCancellation releases a reservation a cancelled sale no longer needs.
	SalesFulfillmentTypeCancellation = SalesFulfillmentRequestType("cancellation")
)

// SalesFulfillmentRequestStatus is how far Inventory has got.
type SalesFulfillmentRequestStatus string

const (
	SalesFulfillmentStatusPending = SalesFulfillmentRequestStatus("pending")

	// SalesFulfillmentStatusAccepted means Inventory took the request and reserved stock. The goods
	// have not moved yet: money captured but goods not dispensed lives exactly between accepted and
	// completed.
	SalesFulfillmentStatusAccepted = SalesFulfillmentRequestStatus("accepted")

	SalesFulfillmentStatusCompleted = SalesFulfillmentRequestStatus("completed")
	SalesFulfillmentStatusRejected  = SalesFulfillmentRequestStatus("rejected")
	SalesFulfillmentStatusCancelled = SalesFulfillmentRequestStatus("cancelled")
)

// SalesFiscalIntent is what commercially happened, never what document to produce. Sales reports the
// event; the provider decides whether it needs an invoice, a credit note or an adjustment
// declaration. The absence of a document type in this enum is the point of it.
type SalesFiscalIntent string

const (
	SalesFiscalIntentIssueOriginal          = SalesFiscalIntent("ISSUE_ORIGINAL")
	SalesFiscalIntentAdjustForFullReturn    = SalesFiscalIntent("ADJUST_FOR_FULL_RETURN")
	SalesFiscalIntentAdjustForPartialReturn = SalesFiscalIntent("ADJUST_FOR_PARTIAL_RETURN")
	SalesFiscalIntentAdjustPrice            = SalesFiscalIntent("ADJUST_PRICE")
)

// SalesFiscalRequestStatus is how far the eInvoice provider has got.
type SalesFiscalRequestStatus string

const (
	// SalesFiscalStatusPending means asked and not yet answered. A request never moves optimistically
	// to issued: the far side is a third-party call, and reporting early would tell a customer they
	// hold a VAT invoice that does not exist.
	SalesFiscalStatusPending = SalesFiscalRequestStatus("pending")

	// SalesFiscalStatusIssued means the provider confirmed the document exists. Only this status
	// carries a provider_reference, and only this one moves the order to invoice_status 'issued'.
	SalesFiscalStatusIssued = SalesFiscalRequestStatus("issued")

	// SalesFiscalStatusFailed is normal operation, not a fault: a provider is unreachable,
	// rate-limited, or refuses incomplete buyer information. It is the only status besides cancelled
	// that leaves the bill free to ask again.
	SalesFiscalStatusFailed = SalesFiscalRequestStatus("failed")

	SalesFiscalStatusCancelled = SalesFiscalRequestStatus("cancelled")
)

// SalesQuotationStatus is where an offer stands. `expired` and `cancelled` are distinct on purpose:
// one lapsed on its own terms, the other was withdrawn, and they are served differently.
type SalesQuotationStatus string

const (
	SalesQuotationStatusDraft = SalesQuotationStatus("draft")

	// SalesQuotationStatusSent means the customer has seen it: editing one is changing something
	// already promised, which the audit trail records.
	SalesQuotationStatusSent = SalesQuotationStatus("sent")

	SalesQuotationStatusAccepted  = SalesQuotationStatus("accepted")
	SalesQuotationStatusExpired   = SalesQuotationStatus("expired")
	SalesQuotationStatusCancelled = SalesQuotationStatus("cancelled")
)

// SalesReturnStatus is where a return stands commercially. It reaches `completed` once the goods
// are back and the money is refunded, not when the tax paperwork succeeds: a failed fiscal
// adjustment must not roll back a completed inventory return or refund.
type SalesReturnStatus string

const (
	SalesReturnStatusDraft = SalesReturnStatus("draft")

	// SalesReturnStatusApproved is the last point at which cancelling costs nothing.
	SalesReturnStatusApproved = SalesReturnStatus("approved")

	// SalesReturnStatusProcessing means at least one of the three side effects is in flight; the
	// return can no longer be cancelled.
	SalesReturnStatusProcessing = SalesReturnStatus("processing")

	// SalesReturnStatusCompleted means commercially complete - goods back, money refunded. It says
	// nothing about the tax correction, which carries its own status and may still be failed.
	SalesReturnStatusCompleted = SalesReturnStatus("completed")

	SalesReturnStatusCancelled = SalesReturnStatus("cancelled")
)

// SalesReturnStepStatus is where one of a return's three side effects stands. The same values serve
// inventory, refund and fiscal. `not_required` is what makes a return of services completable: the
// inventory step is satisfied by not applying rather than by succeeding.
type SalesReturnStepStatus string

const (
	// SalesReturnStepNotRequired means this step does not apply, and counts as done when deciding
	// whether the return is complete.
	SalesReturnStepNotRequired = SalesReturnStepStatus("not_required")

	SalesReturnStepPending    = SalesReturnStepStatus("pending")
	SalesReturnStepProcessing = SalesReturnStepStatus("processing")
	SalesReturnStepCompleted  = SalesReturnStepStatus("completed")

	// SalesReturnStepFailed is normal operation rather than a fault. On the fiscal step it must not
	// block the return: it is a retryable to-do, and the customer is already whole.
	SalesReturnStepFailed = SalesReturnStepStatus("failed")
)

// SalesReturnDisposition is what the business intends should happen to returned goods. Intent only:
// Inventory decides the actual location and movement, so no warehouse is named here.
type SalesReturnDisposition string

const (
	SalesReturnDispositionRestock    = SalesReturnDisposition("restock")
	SalesReturnDispositionScrap      = SalesReturnDisposition("scrap")
	SalesReturnDispositionQuarantine = SalesReturnDisposition("quarantine")
)

// SalesRefundPaymentStatus is where one leg of a refund stands with its provider. Only `completed`
// counts as money actually returned - the mirror of a payment's `captured`. A pending refund treated
// as done reports a customer repaid who is still waiting.
type SalesRefundPaymentStatus string

const (
	SalesRefundPaymentStatusPending    = SalesRefundPaymentStatus("pending")
	SalesRefundPaymentStatusProcessing = SalesRefundPaymentStatus("processing")
	SalesRefundPaymentStatusCompleted  = SalesRefundPaymentStatus("completed")
	SalesRefundPaymentStatusFailed     = SalesRefundPaymentStatus("failed")
)
