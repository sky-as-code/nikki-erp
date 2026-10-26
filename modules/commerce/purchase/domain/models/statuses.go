package models

// Status values for the lifecycle-bearing resources. Declared here, not inline, so a typo cannot
// silently make a comparison never true. Stored lower-case, matching the JSON schemas exactly;
// upper-case spellings elsewhere are presentation only.

type PurchaseOrderStatus string

const (
	// PurchaseOrderStatusRfq is a draft request for quotation, not yet sent.
	PurchaseOrderStatusRfq     = PurchaseOrderStatus("rfq")
	PurchaseOrderStatusRfqSent = PurchaseOrderStatus("rfq_sent")
	// PurchaseOrderStatusToApprove is reached only when the org's policy requires approval for this
	// order's total.
	PurchaseOrderStatusToApprove = PurchaseOrderStatus("to_approve")
	// PurchaseOrderStatusPurchaseOrder is a committed order — the same record as the RFQ it began as.
	PurchaseOrderStatusPurchaseOrder = PurchaseOrderStatus("purchase_order")
	// PurchaseOrderStatusCancelled is terminal, and the only status an order may be deleted from.
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
	// AgreementTypeBlanketOrder commits to quantities at agreed prices and tracks drawdown.
	AgreementTypeBlanketOrder = AgreementType("blanket_order")
	// AgreementTypePurchaseTemplate is a reusable skeleton with no commitment attached.
	AgreementTypePurchaseTemplate = AgreementType("purchase_template")
)

type PurchaseOrderLineType string

const (
	// PurchaseOrderLineTypeProduct is the only line type contributing to the totals.
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
	// ApprovalModeOneStep sends a confirmed order straight to purchase_order.
	ApprovalModeOneStep = ApprovalMode("one_step")
	// ApprovalModeTwoStep routes it through to_approve when it reaches the threshold.
	ApprovalModeTwoStep = ApprovalMode("two_step")
)

type PoModificationPolicy string

const (
	PoModificationPolicyAllowEdit = PoModificationPolicy("allow_edit")
	// PoModificationPolicyAutoLock sets is_locked at confirmation.
	PoModificationPolicyAutoLock = PoModificationPolicy("auto_lock")
)
