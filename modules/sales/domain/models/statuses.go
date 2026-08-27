package models

// The status values of the lifecycle-bearing Sales resources, and the enums that go with them.
//
// They are declared here rather than inline at their use sites so that the schema JSON and the code
// reading it cannot drift: a typo in a comparison would otherwise be a condition that is silently
// never true, which no compiler catches.
//
// Stored lower-case, matching the strings the JSON schemas declare.

// SalesChannelStatus is the business lifecycle of a sales channel.
//
// It is deliberately separate from is_archived, which is the system lifecycle. CR §9 and §10 keep
// them apart because they answer different questions: suspended means "not selling right now",
// archived means "no longer part of the catalogue of channels". Collapsing them would lose the
// difference between a channel paused for the season and one retired for good.
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
	// SalesPointStatusSuspended stops new orders but keeps history, returns and refunds available
	// — what a temporarily offline kiosk needs (CR §14, §52).
	SalesPointStatusSuspended = SalesPointStatus("suspended")
)

// SalesOrderStatus is the document's own lifecycle — whether the sale has been committed to,
// finished, or called off.
//
// This is ONE of four independent status fields on a sales order, and the four are never collapsed
// (BR 9). Each answers a different question, and the answers do not imply one another: an order can
// be confirmed and fully paid but undelivered, or delivered and unpaid, or complete with its VAT
// invoice rejected. A single status would have to invent an ordering between those that the
// business does not have.
type SalesOrderStatus string

const (
	// SalesOrderStatusDraft is a document still being built. Lines may be added, changed and
	// removed, and prices are recalculated on every change.
	SalesOrderStatusDraft = SalesOrderStatus("draft")

	// SalesOrderStatusConfirmed is a sale the business has committed to. The snapshot fields on
	// every line become immutable at this moment (BR 11).
	SalesOrderStatusConfirmed = SalesOrderStatus("confirmed")

	// SalesOrderStatusProcessing is confirmed and part-way through fulfilment. It exists so that
	// "committed to but not yet delivered" is distinguishable from "delivery under way", which is
	// the difference between an order a customer may still amend cheaply and one already moving
	// through a warehouse. An order needing no fulfilment at all never enters it (D-14).
	SalesOrderStatusProcessing = SalesOrderStatus("processing")

	// SalesOrderStatusCompleted is paid and fulfilled. Terminal for this dimension only — returns,
	// refunds and invoicing all still happen afterwards.
	SalesOrderStatusCompleted = SalesOrderStatus("completed")

	// SalesOrderStatusCancelled is a sale that will not happen. The record is kept rather than
	// deleted: it is evidence of what was attempted.
	SalesOrderStatusCancelled = SalesOrderStatus("cancelled")
)

// SalesOrderPaymentStatus is how much of the money has arrived.
//
// Derived from the sum of the order's payments rather than set directly, so it can always be
// recomputed from the payment rows and can never disagree with them.
type SalesOrderPaymentStatus string

const (
	SalesOrderPaymentStatusUnpaid        = SalesOrderPaymentStatus("unpaid")
	SalesOrderPaymentStatusPartiallyPaid = SalesOrderPaymentStatus("partially_paid")
	SalesOrderPaymentStatusPaid          = SalesOrderPaymentStatus("paid")

	// SalesOrderPaymentStatusOverpaid is a real state rather than an error condition: a cash till
	// takes what the customer hands over, and whether change may be given back is a policy setting
	// (allow_cash_change), not a schema question.
	SalesOrderPaymentStatusOverpaid = SalesOrderPaymentStatus("overpaid")

	SalesOrderPaymentStatusRefunded          = SalesOrderPaymentStatus("refunded")
	SalesOrderPaymentStatusPartiallyRefunded = SalesOrderPaymentStatus("partially_refunded")
)

// SalesOrderFulfillmentStatus is how much of the goods have moved.
//
// Derived from the lines' fulfilled_quantity and returned_quantity totals, for the same reason the
// payment status is derived from its payments: a status settable independently of the quantities it
// summarises would eventually contradict them.
type SalesOrderFulfillmentStatus string

const (
	SalesOrderFulfillmentStatusPending = SalesOrderFulfillmentStatus("pending")

	// SalesOrderFulfillmentStatusNotRequired is an order with nothing to hand over — every line is
	// a service, a fee or a non-stocked item. It is a distinct value rather than an immediate
	// "fulfilled" because the two answer different questions: fulfilled means goods moved, this
	// means none were ever owed. D-15 accepts both as satisfying completion.
	SalesOrderFulfillmentStatusNotRequired = SalesOrderFulfillmentStatus("not_required")

	SalesOrderFulfillmentStatusPartiallyFulfilled = SalesOrderFulfillmentStatus("partially_fulfilled")
	SalesOrderFulfillmentStatusFulfilled          = SalesOrderFulfillmentStatus("fulfilled")
	SalesOrderFulfillmentStatusReturned           = SalesOrderFulfillmentStatus("returned")
	SalesOrderFulfillmentStatusPartiallyReturned  = SalesOrderFulfillmentStatus("partially_returned")
)

// SalesOrderInvoiceStatus is where the VAT invoice has got to.
//
// Separate from the other three because issuing a fiscal document is an EXTERNAL act with its own
// failure mode: a tax authority can reject an invoice for a sale that is paid, delivered and
// complete, and that failure must be visible without making the sale itself look broken.
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
	// sales_order_line_components, because Inventory fulfils real variants and never a virtual
	// combo (BR 17).
	SalesOrderLineTypeCombo = SalesOrderLineType("combo")

	// SalesOrderLineTypePromotionReward is a free item given by a promotion. A real line rather
	// than an adjustment, because Inventory must physically fulfil it and its VAT treatment is a
	// line-level question (D-11).
	SalesOrderLineTypePromotionReward = SalesOrderLineType("promotion_reward")
)

// SalesOrderPricingSource is where a line's price came from.
//
// It lets the price-explanation API answer "why does this cost this" without replaying the whole
// engine, and makes a manual override visible as a deliberate act rather than an unexplained number.
type SalesOrderPricingSource string

const (
	SalesOrderPricingSourceCatalogue       = SalesOrderPricingSource("catalogue")
	SalesOrderPricingSourcePricelist       = SalesOrderPricingSource("pricelist")
	SalesOrderPricingSourceCombo           = SalesOrderPricingSource("combo")
	SalesOrderPricingSourcePromotionReward = SalesOrderPricingSource("promotion_reward")
	SalesOrderPricingSourceManualOverride  = SalesOrderPricingSource("manual_override")
)

// VoucherCodeStatus is whether a code may currently be applied.
//
// Expiry is deliberately absent. A code past its valid_until is unusable, but that is a function of
// the window and the current time rather than a stored value - making it a status would require a
// job to run for a code to become expired, and a code would be wrongly usable until that job did.
type VoucherCodeStatus string

const (
	VoucherCodeStatusActive = VoucherCodeStatus("active")

	// VoucherCodeStatusDisabled is an operator's decision, and reversible. A campaign found to be
	// mispriced is stopped this way rather than by archiving, so it can be turned back on.
	VoucherCodeStatusDisabled = VoucherCodeStatus("disabled")

	// VoucherCodeStatusExhausted is derived, not chosen: usage_count reached usage_limit. It is set
	// by the redemption path rather than by hand, and a return that restores a use moves it back.
	VoucherCodeStatusExhausted = VoucherCodeStatus("exhausted")
)

// VoucherRedemptionStatus tracks one code's use on one order.
//
// Four values in two pairs. A reservation is taken when a code is applied to a draft and is settled
// one of two ways: 'redeemed' if the order confirms, 'released' if it never does. A redemption can
// then be undone by a return, which is 'reversed'. Keeping released and reversed apart matters
// because they mean different things to a campaign report - one says the code was never really
// used, the other says it was used and given back.
type VoucherRedemptionStatus string

const (
	// VoucherRedemptionStatusReserved holds a use while an order is still a draft. This is what
	// stops a second customer taking the last use of a voucher that is already in someone's basket
	// (BR 82) - a usage counter alone could not, since a draft has not incremented it yet.
	VoucherRedemptionStatusReserved = VoucherRedemptionStatus("reserved")

	VoucherRedemptionStatusRedeemed = VoucherRedemptionStatus("redeemed")

	// VoucherRedemptionStatusReleased gives the hold back without a sale: the draft was cancelled
	// or expired.
	VoucherRedemptionStatusReleased = VoucherRedemptionStatus("released")

	// VoucherRedemptionStatusReversed gives a completed use back after a return (BR 32). Whether a
	// return restores at all is the program's decision, not the redemption's.
	VoucherRedemptionStatusReversed = VoucherRedemptionStatus("reversed")
)

// SalesBillStatus is where a settlement unit stands.
//
// A bill is NEVER a VAT invoice (BR 33). The separation is explicit in the requirement and is why
// BR 34 exists: a bill is how the money is collected, an invoice is the legal document, and one sale
// can need several of the first and one of the second - or the other way round.
type SalesBillStatus string

const (
	SalesBillStatusOpen = SalesBillStatus("open")

	// SalesBillStatusSettled means the money is fully in. The line allocations freeze at this point
	// (BR 76), because they are what the payment was measured against.
	SalesBillStatusSettled = SalesBillStatus("settled")

	// SalesBillStatusCancelled marks a bill superseded by a split or a merge. The row stays (BR 83)
	// because the lineage relations point at it, and an auditor tracing a payment needs to arrive
	// somewhere.
	SalesBillStatusCancelled = SalesBillStatus("cancelled")
)

// SalesBillRelationType says which operation produced a lineage row.
//
// Both types read source -> target, so the type is what distinguishes "this bill became those three"
// from "those three became this one". The row shape alone cannot.
type SalesBillRelationType string

const (
	SalesBillRelationSplitInto  = SalesBillRelationType("split_into")
	SalesBillRelationMergedInto = SalesBillRelationType("merged_into")
)

// SalesPaymentStatus is where one payment stands with its provider.
//
// Only `captured` counts toward settling a bill. An authorization is a hold the provider may still
// release, and treating it as money in would settle a bill against funds that never arrived - the
// single most expensive confusion available in this enum.
type SalesPaymentStatus string

const (
	SalesPaymentStatusPending    = SalesPaymentStatus("pending")
	SalesPaymentStatusAuthorized = SalesPaymentStatus("authorized")
	SalesPaymentStatusCaptured   = SalesPaymentStatus("captured")
	SalesPaymentStatusFailed     = SalesPaymentStatus("failed")
	SalesPaymentStatusCancelled  = SalesPaymentStatus("cancelled")
)

// SalesFulfillmentRequestType is what Sales is asking Inventory to do.
//
// INTENT, never instruction (BR 44). Sales says goods were sold and must leave; Inventory decides
// availability, warehouse, location and the movements that follow. Sales never touches stock.
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
	// have NOT moved yet, and the distinction is load-bearing: BR 7.3's failure - money captured but
	// goods not dispensed - lives exactly between accepted and completed.
	SalesFulfillmentStatusAccepted = SalesFulfillmentRequestStatus("accepted")

	SalesFulfillmentStatusCompleted = SalesFulfillmentRequestStatus("completed")
	SalesFulfillmentStatusRejected  = SalesFulfillmentRequestStatus("rejected")
	SalesFulfillmentStatusCancelled = SalesFulfillmentRequestStatus("cancelled")
)

// SalesFiscalIntent is what commercially happened, never what document to produce.
//
// BUSINESS INTENT, never a vendor command (BR 48, BR 49). Sales reports the event; the provider
// decides whether it needs an invoice, a credit note or an adjustment declaration (BR 50). The
// absence of a document type in this enum is the point of it.
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
	// SalesFiscalStatusPending means asked and not yet answered. A request stays here rather than
	// moving optimistically to issued (BR 77): the far side is a third-party network call, and
	// reporting a document as issued before confirmation would tell a customer they hold a VAT
	// invoice that does not exist.
	SalesFiscalStatusPending = SalesFiscalRequestStatus("pending")

	// SalesFiscalStatusIssued means the provider confirmed the document exists. Only this status
	// carries a provider_reference, and only this one moves the order to invoice_status 'issued'.
	SalesFiscalStatusIssued = SalesFiscalRequestStatus("issued")

	// SalesFiscalStatusFailed is NORMAL OPERATION, not a fault: a provider is unreachable,
	// rate-limited, or refuses a request whose buyer information is incomplete. It is the only
	// status besides cancelled that leaves the bill free to ask again.
	SalesFiscalStatusFailed = SalesFiscalRequestStatus("failed")

	SalesFiscalStatusCancelled = SalesFiscalRequestStatus("cancelled")
)

// SalesQuotationStatus is where an offer stands (BR 87.1).
//
// `expired` and `cancelled` are distinct on purpose: one lapsed on its own terms and the other was
// withdrawn. A customer arriving with a lapsed quote is served differently from one whose quote was
// pulled — the first is re-quoted, the second is a conversation — and a single "closed" state would
// lose exactly that difference.
type SalesQuotationStatus string

const (
	SalesQuotationStatusDraft = SalesQuotationStatus("draft")

	// SalesQuotationStatusSent means the customer has seen it. Distinct from draft because the offer
	// is out: editing one is changing something already promised, which the audit trail records.
	SalesQuotationStatusSent = SalesQuotationStatus("sent")

	SalesQuotationStatusAccepted  = SalesQuotationStatus("accepted")
	SalesQuotationStatusExpired   = SalesQuotationStatus("expired")
	SalesQuotationStatusCancelled = SalesQuotationStatus("cancelled")
)
