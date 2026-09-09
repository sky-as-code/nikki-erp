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

// FulfillmentType is how goods reach the customer, and the field that chooses which execution
// workflow runs. Only kiosk dispense is implemented; the rest are declared so that adding one later
// is code rather than a schema migration, and every execution path checks rather than assumes.
type FulfillmentType string

const (
	// FulfillmentTypeKioskDispense is a vending machine handing the goods over itself. The only
	// type with a workflow in this module.
	FulfillmentTypeKioskDispense = FulfillmentType("kiosk_dispense")

	// The four below reserve their names and nothing else. A method carrying one may be configured
	// and read, but no fulfillment of that type can be executed yet.
	FulfillmentTypeCarrierShipping     = FulfillmentType("carrier_shipping")
	FulfillmentTypeInternalDelivery    = FulfillmentType("internal_delivery")
	FulfillmentTypeStorePickup         = FulfillmentType("store_pickup")
	FulfillmentTypeExternalFulfillment = FulfillmentType("external_fulfillment")
)

// InitialTargetSelection decides where the FIRST fulfillment attempt is aimed. It is a strategy for
// picking a target, never a target itself: "this kiosk" is how the target is chosen once, after
// which the target lives on the fulfillment and may move, while the method stays the same.
type InitialTargetSelection string

const (
	// InitialTargetSelectionCurrentSalesOutlet aims at the sales point the order was raised at —
	// the machine the customer is standing in front of.
	InitialTargetSelectionCurrentSalesOutlet = InitialTargetSelection("current_sales_outlet")

	// InitialTargetSelectionCustomerSelectedOutlet requires the client to name a point before the
	// order may be confirmed. An order without one is refused rather than defaulted: guessing which
	// kiosk a customer meant to collect from would send the goods to the wrong town.
	InitialTargetSelectionCustomerSelectedOutlet = InitialTargetSelection("customer_selected_outlet")
)

// FulfillmentFailureAction is what happens to a quantity that is terminally undeliverable. It
// selects a policy, and is never itself a status: the refund it may raise has its own lifecycle,
// which succeeds or fails independently of the fulfillment that asked for it.
type FulfillmentFailureAction string

const (
	// FulfillmentFailureActionAutoRefund refunds exactly the failed quantity with nobody asked.
	// What an anonymous walk-up sale needs, because there is no customer to come back to.
	FulfillmentFailureActionAutoRefund = FulfillmentFailureAction("auto_refund")

	// FulfillmentFailureActionCustomerActionRequired refunds nothing and waits. The customer keeps
	// the entitlement and chooses: another kiosk, another attempt, or their money back.
	FulfillmentFailureActionCustomerActionRequired = FulfillmentFailureAction("customer_action_required")

	// FulfillmentFailureActionManualResolution parks it for an operator to decide.
	FulfillmentFailureActionManualResolution = FulfillmentFailureAction("manual_resolution")
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

// SalesBillingInstructionStatus is where a sale's billing arrangement stands.
//
// The line that matters is between `ready` and everything before it: only `ready` is picked up by
// the issuance job, so marking ready is the buyer's consent to be billed. An instruction still being
// filled in must never produce a document, and one already claimed must not change under the worker
// building from it.
type SalesBillingInstructionStatus string

const (
	// SalesBillingInstructionStatusDraft: being filled in. Not yet anyone's consent to anything.
	SalesBillingInstructionStatusDraft = SalesBillingInstructionStatus("draft")

	// SalesBillingInstructionStatusReady: the buyer confirmed these details. The only status the
	// issuance job will claim.
	SalesBillingInstructionStatusReady = SalesBillingInstructionStatus("ready")

	// SalesBillingInstructionStatusProcessing: a worker holds it and the snapshot is frozen. An
	// instruction stuck here after a lost reply is the case reconciliation exists for - it is NOT
	// swept back to ready, because the document may already exist.
	SalesBillingInstructionStatusProcessing = SalesBillingInstructionStatus("processing")

	// SalesBillingInstructionStatusIssued: the document exists. Terminal, and frozen: correcting an
	// issued document is the provider's own regulated workflow, never an edit here.
	SalesBillingInstructionStatusIssued = SalesBillingInstructionStatus("issued")

	// SalesBillingInstructionStatusFailed: the provider definitely did not issue. Editable, so the
	// details can be corrected and marked ready again. Reached only when the answer was a refusal -
	// a lost reply leaves the instruction `processing` instead.
	SalesBillingInstructionStatusFailed = SalesBillingInstructionStatus("failed")

	// SalesBillingInstructionStatusCancelled: withdrawn before issuance. Terminal and kept for
	// audit; a buyer who changes their mind gets a new instruction, never this one reopened.
	SalesBillingInstructionStatusCancelled = SalesBillingInstructionStatus("cancelled")
)

// SalesBillingInstructionSource is who created the instruction. Kept because the three carry
// different assurance about the buyer's identity, which is what a dispute over a wrong tax code
// turns on.
type SalesBillingInstructionSource string

const (
	SalesBillingInstructionSourceBackOffice = SalesBillingInstructionSource("back_office")
	SalesBillingInstructionSourcePublic     = SalesBillingInstructionSource("public")
	SalesBillingInstructionSourceImport     = SalesBillingInstructionSource("import")
)

// SalesBillingAttemptStatus is how one issuance attempt ended.
//
// `unknown` is the whole reason attempts are recorded separately from the instruction: the request
// reached the provider and the reply did not come back, so nobody can say whether a document was
// created. It is not a failure and must not be retried as one.
type SalesBillingAttemptStatus string

const (
	SalesBillingAttemptStatusProcessing = SalesBillingAttemptStatus("processing")
	SalesBillingAttemptStatusSucceeded  = SalesBillingAttemptStatus("succeeded")
	SalesBillingAttemptStatusFailed     = SalesBillingAttemptStatus("failed")
	SalesBillingAttemptStatusUnknown    = SalesBillingAttemptStatus("unknown")
)

// FulfillmentStatus is how far ONE fulfillment has got in delivering its own goods. It is
// per-fulfillment and deliberately distinct from SalesOrderFulfillmentStatus, which rolls the whole
// order up for reporting: an order may hold two fulfillments at different stages, and collapsing
// them into one value would lose the only information a customer service agent needs.
//
// Refund state is NOT in this enum. A refund has its own lifecycle that succeeds or fails
// independently, and folding it in here would make "the goods are still owed" and "the money came
// back" the same fact — they are not, and a failed refund would otherwise silently reopen a
// fulfillment that was never going to deliver anything again.
type FulfillmentStatus string

const (
	// FulfillmentStatusPendingReservation is created but holding no stock yet. Nothing is
	// promised to the customer in this state.
	FulfillmentStatusPendingReservation = FulfillmentStatus("pending_reservation")

	// FulfillmentStatusReserved holds the stock but is not yet clear to hand it over, because
	// the money has not settled. Reserving before payment is deliberate: taking payment for goods
	// that were sold out between the price quote and the tap is the failure this ordering avoids.
	FulfillmentStatusReserved = FulfillmentStatus("reserved")

	// FulfillmentStatusReady is paid for and reserved — an executor may now be asked to hand
	// the goods over.
	FulfillmentStatusReady = FulfillmentStatus("ready")

	// FulfillmentStatusInProgress has an attempt outstanding. No second attempt may start
	// while a fulfillment sits here, because two machines dispensing the same reservation would
	// hand over goods that were only paid for once.
	FulfillmentStatusInProgress = FulfillmentStatus("in_progress")

	// FulfillmentStatusPartiallyFulfilled handed some quantity over and still owes the rest.
	// A resting state, not a terminal one: the outstanding quantity may still be attempted.
	FulfillmentStatusPartiallyFulfilled = FulfillmentStatus("partially_fulfilled")

	// FulfillmentStatusWaitingCustomerAction owes goods it will not attempt again on its own.
	// The customer chooses what happens next — another kiosk, another attempt, or a refund — which
	// is why this is not a failure: the entitlement is intact and only the initiative has moved.
	FulfillmentStatusWaitingCustomerAction = FulfillmentStatus("waiting_customer_action")

	// FulfillmentStatusExpired let its reservation lapse; the stock went back to the sellable
	// pool. Expiry never refunds by itself — the customer keeps the entitlement and may reselect a
	// target or ask for the money back.
	FulfillmentStatusExpired = FulfillmentStatus("expired")

	// FulfillmentStatusCompleted owes nothing: every ordered quantity was either handed over
	// or successfully refunded. Reached on remaining quantity alone, whatever the refunds did.
	FulfillmentStatusCompleted = FulfillmentStatus("completed")

	// FulfillmentStatusCancelled was called off before delivering, with its reservation
	// released.
	FulfillmentStatusCancelled = FulfillmentStatus("cancelled")
)

// FulfillmentItemStatus is the same question asked of one item. It carries no
// `waiting_customer_action`: waiting is a decision about the whole fulfillment, taken once by the
// failure policy, and recording it per item would let two items disagree about whose turn it is.
type FulfillmentItemStatus string

const (
	FulfillmentItemStatusPending            = FulfillmentItemStatus("pending")
	FulfillmentItemStatusReserved           = FulfillmentItemStatus("reserved")
	FulfillmentItemStatusPartiallyFulfilled = FulfillmentItemStatus("partially_fulfilled")
	FulfillmentItemStatusFulfilled          = FulfillmentItemStatus("fulfilled")
	FulfillmentItemStatusCancelled          = FulfillmentItemStatus("cancelled")
)

// CustomerIdentityMode records whether the buyer's identity was verified server-side when the order
// was raised. It is a SNAPSHOT the server calculates from the request's own authentication, never
// an input: a client that could assert its own identity mode would be choosing which fulfillment
// policies it qualifies for, including whether a failure refunds automatically or waits for a
// customer who cannot be identified.
type CustomerIdentityMode string

const (
	// CustomerIdentityModeAnonymous is a walk-up sale. Nobody can be contacted afterwards, which is
	// why anonymous orders need policies that resolve themselves.
	CustomerIdentityModeAnonymous = CustomerIdentityMode("anonymous")

	// CustomerIdentityModeAuthenticated carries a verified customer principal, so an entitlement
	// may be left open for them to come back to.
	CustomerIdentityModeAuthenticated = CustomerIdentityMode("authenticated")
)

// FulfillmentAttemptStatus is how one try at handing goods over ended. It is DERIVED from the
// attempt's items, never asserted by whoever reports the result: a reporter that could claim success
// while reporting a shortfall would settle a sale that still owes goods.
//
// `pending` is the state that blocks a second attempt. Two executors acting on one reservation would
// hand over goods that were paid for once, so a fulfillment with an outstanding attempt accepts no
// other until that one reports.
type FulfillmentAttemptStatus string

const (
	FulfillmentAttemptStatusPending            = FulfillmentAttemptStatus("pending")
	FulfillmentAttemptStatusSucceeded          = FulfillmentAttemptStatus("succeeded")
	FulfillmentAttemptStatusPartiallySucceeded = FulfillmentAttemptStatus("partially_succeeded")
	FulfillmentAttemptStatusFailed             = FulfillmentAttemptStatus("failed")
	FulfillmentAttemptStatusCancelled          = FulfillmentAttemptStatus("cancelled")
)

// FulfillmentAttemptItemResult is the same question asked of one product within a try, and is
// likewise derived from the dispensed and failed quantities rather than taken on trust.
type FulfillmentAttemptItemResult string

const (
	FulfillmentAttemptItemResultPending = FulfillmentAttemptItemResult("pending")
	FulfillmentAttemptItemResultSuccess = FulfillmentAttemptItemResult("success")
	FulfillmentAttemptItemResultPartial = FulfillmentAttemptItemResult("partial")
	FulfillmentAttemptItemResultFailure = FulfillmentAttemptItemResult("failure")
)

// TargetChangeReason is why a fulfillment's target moved, in business terms rather than technical
// ones. The same physical movement means different things — a customer choosing another kiosk is a
// service event, an operator moving stock off a failing machine is an operational one — and a
// support agent reading the history needs to tell them apart.
type TargetChangeReason string

const (
	// TargetChangeReasonCustomerRequested is the customer choosing somewhere else to collect.
	TargetChangeReasonCustomerRequested = TargetChangeReason("customer_requested")

	// TargetChangeReasonStockUnavailable is the original target no longer being able to supply.
	TargetChangeReasonStockUnavailable = TargetChangeReason("stock_unavailable")

	// TargetChangeReasonOperational covers an operator moving a delivery for reasons of their own: a
	// machine taken out of service, a site closing early.
	TargetChangeReasonOperational = TargetChangeReason("operational")

	// TargetChangeReasonReservationExpired marks the case where the old hold had already lapsed, so
	// a fresh reservation was taken rather than an existing one moved. Recorded distinctly because
	// the stock genuinely went back to the sellable pool in between, and the history should not
	// suggest an unbroken hold that never existed.
	TargetChangeReasonReservationExpired = TargetChangeReason("reservation_expired")
)

// TargetChangeActorType is what sort of actor made a change, taken from the request's own
// authentication and never accepted from a caller. "The kiosk moved it" and "a support agent moved
// it" are different answers to a customer asking why their order changed machine.
type TargetChangeActorType string

const (
	TargetChangeActorTypeUser    = TargetChangeActorType("user")
	TargetChangeActorTypeService = TargetChangeActorType("service")
	TargetChangeActorTypeSystem  = TargetChangeActorType("system")
)

// SalesReturnType is whether goods are coming back with the money.
//
// The distinction exists because a failed dispense has no goods to return: they never reached the
// customer, and asking Inventory to receive them would book stock that never moved.
type SalesReturnType string

const (
	// SalesReturnTypeGoodsReturn is the ordinary case — the customer brings something back, and the
	// inventory step receives it.
	SalesReturnTypeGoodsReturn = SalesReturnType("goods_return")

	// SalesReturnTypeRefundOnly moves money alone and skips the inventory step entirely. Inventory
	// still owns the physical consequence of a failed dispense; it simply already dealt with it when
	// the attempt result was applied.
	SalesReturnTypeRefundOnly = SalesReturnType("refund_only")
)

// SalesRefundReason is why money is going back, as a category the system acts on. It is deliberately
// separate from the free-text reason a person types, which is what reaches the fiscal adjustment.
type SalesRefundReason string

const (
	// SalesRefundReasonCustomerRequested is somebody asking for their money back.
	SalesRefundReasonCustomerRequested = SalesRefundReason("customer_requested")

	// SalesRefundReasonFulfillmentFailure is a refund raised because goods could not be handed over.
	// The only kind that may exist with nobody having asked, which is what an anonymous kiosk sale
	// needs: there is no customer to come back to.
	SalesRefundReasonFulfillmentFailure = SalesRefundReason("fulfillment_failure")
)

// The two refund-status values doc 05 §5 adds beyond the shared step enum.
const (
	// SalesReturnStepPartiallySucceeded is a refund where some payment legs settled and others did
	// not. Real rather than theoretical: a refund is allocated proportionally across the legs that
	// paid for the order, and one provider can fail while another succeeds.
	SalesReturnStepPartiallySucceeded = SalesReturnStepStatus("partially_succeeded")

	// SalesReturnStepCancelled is a refund called off before any money moved.
	SalesReturnStepCancelled = SalesReturnStepStatus("cancelled")
)
