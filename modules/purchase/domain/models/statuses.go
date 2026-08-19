package models

// The status values of the two lifecycle-bearing resources, and the enums that go with them.
//
// They are declared here rather than inline at their use sites so that the schema JSON and the code
// reading it cannot drift: a typo in a comparison would otherwise be a condition that is silently
// never true, which no compiler catches.
//
// Stored lower-case. The requirement writes them upper-case, which is presentation — every enum in
// this codebase stores lower-case, and the JSON schemas declare these exact strings.

type PurchaseOrderStatus string

const (
	// PurchaseOrderStatusRfq is a draft request for quotation, not yet sent to anyone.
	PurchaseOrderStatusRfq = PurchaseOrderStatus("rfq")
	// PurchaseOrderStatusRfqSent means the vendor has been asked to quote.
	PurchaseOrderStatusRfqSent = PurchaseOrderStatus("rfq_sent")
	// PurchaseOrderStatusToApprove means confirmation is waiting on an approver, which happens
	// only when the organization's policy requires it for this order's total.
	PurchaseOrderStatusToApprove = PurchaseOrderStatus("to_approve")
	// PurchaseOrderStatusPurchaseOrder is a committed order. The same record as the RFQ it began
	// as: confirming changes the status and nothing about its identity.
	PurchaseOrderStatusPurchaseOrder = PurchaseOrderStatus("purchase_order")
	// PurchaseOrderStatusCancelled is terminal, and the only status a purchase order may be
	// deleted from.
	PurchaseOrderStatusCancelled = PurchaseOrderStatus("cancelled")
)

type AgreementStatus string

const (
	AgreementStatusDraft     = AgreementStatus("draft")
	AgreementStatusConfirmed = AgreementStatus("confirmed")
	AgreementStatusClosed    = AgreementStatus("closed")
	AgreementStatusCancelled = AgreementStatus("cancelled")
)

type AgreementType string

const (
	// AgreementTypeBlanketOrder commits to quantities at agreed prices, and tracks what has been
	// drawn down against it.
	AgreementTypeBlanketOrder = AgreementType("blanket_order")
	// AgreementTypePurchaseTemplate is a reusable skeleton with no commitment attached.
	AgreementTypePurchaseTemplate = AgreementType("purchase_template")
)

type PurchaseOrderLineType string

const (
	// PurchaseOrderLineTypeProduct is the only line type that buys anything and the only one that
	// contributes to the totals.
	PurchaseOrderLineTypeProduct = PurchaseOrderLineType("product")
	// The remaining three organize the printed document and carry no quantity or price.
	PurchaseOrderLineTypeSection    = PurchaseOrderLineType("section")
	PurchaseOrderLineTypeSubsection = PurchaseOrderLineType("subsection")
	PurchaseOrderLineTypeNote       = PurchaseOrderLineType("note")
)

type PurchasePriority string

const (
	PurchasePriorityNormal = PurchasePriority("normal")
	PurchasePriorityUrgent = PurchasePriority("urgent")
)

type ApprovalMode string

const (
	// ApprovalModeOneStep sends a confirmed order straight to PURCHASE_ORDER.
	ApprovalModeOneStep = ApprovalMode("one_step")
	// ApprovalModeTwoStep routes it through TO_APPROVE when it reaches the threshold.
	ApprovalModeTwoStep = ApprovalMode("two_step")
)

type PoModificationPolicy string

const (
	// PoModificationPolicyAllowEdit leaves a confirmed order editable.
	PoModificationPolicyAllowEdit = PoModificationPolicy("allow_edit")
	// PoModificationPolicyAutoLock sets is_locked at confirmation.
	PoModificationPolicyAutoLock = PoModificationPolicy("auto_lock")
)
