package app

// Permission codes of the custom actions. Each must match a seeded iam_actions row; the resource
// code is the schema name, as for the built-in actions.
//
// These live here rather than beside the routes because the composable engine's RouteDefinition
// carries no permission field: the code is an argument the application service passes to
// AssertAction, so authorization stays in this package and remains reachable from a CQRS handler
// or a job, neither of which goes through a route.
//
// Several codes reuse a built-in verb (create, read, update, set_archived). That is deliberate:
// an action that is an ordinary write under a different name should not require its own grant,
// or an administrator would have to discover and assign it separately to restore a power the
// caller already has. Only a power a grantor would reasonably withhold on its own gets a code.

// Channel and point lifecycle. Suspension is the business lifecycle and archiving is the system
// one; they are separate permissions because they are separate decisions, and the schema comments
// on sales_channel.status insist the two must not substitute for each other.
const (
	PermissionSuspend  = "suspend"
	PermissionActivate = "activate"
)

// Order creation and repricing. Reprice reuses update because it rewrites the same figures an
// ordinary edit would; confirm and cancel are their own codes because each is a state transition
// that releases or commits stock and money.
const (
	PermissionCreateOrder = "create"
	PermissionReprice     = "update"
	PermissionConfirm     = "confirm"
	PermissionCancel      = "cancel"
)

// Discounting. Applying a voucher redeems a code the customer holds; a manual discount is the
// seller overriding the price list on their own authority, which is why it is a distinct grant.
// Granting and revoking share one code: whoever may move the price one way may move it back.
const (
	PermissionApplyVoucher   = "apply_voucher"
	PermissionManualDiscount = "manual_discount"
	PermissionExplainPrice   = "read"
)

// Party assignment. Four codes rather than one because who is billed, who pays and who bought are
// answered by different people in a credit-managed sale.
const (
	PermissionAssignParties = "assign_parties"
	PermissionAssignSoldTo  = "assign_sold_to_party"
	PermissionAssignBillTo  = "assign_bill_to_party"
	PermissionAssignPayer   = "assign_payer_party"
)

// Billing. Splitting and merging restructure what the customer owes, paying and settling move
// money; all four are powers a grantor would withhold independently of ordinary bill edits.
const (
	PermissionSplitBill  = "split"
	PermissionMergeBill  = "merge"
	PermissionPayBill    = "pay"
	PermissionSettleBill = "settle"
)

// Returns. Creating a return is an ordinary create; processing one moves stock and money back,
// so it carries its own code. Cancelling reuses update: it abandons a draft.
const (
	PermissionCreateReturn  = "create"
	PermissionProcessReturn = "process_return"
	PermissionCancelReturn  = "update"
)

// Quotations. Converting produces an order, which is the power worth withholding; sending and
// cancelling are ordinary transitions of a document that binds nobody yet.
const (
	PermissionConvertQuotation    = "convert"
	PermissionTransitionQuotation = "update"
)

// Billing instructions. The lifecycle is draft to ready and back, so mark_ready and
// revert_to_draft are the two codes that matter; create, update and cancel reuse the built-ins.
const (
	PermissionCreateBillingInstruction = "create"
	PermissionUpdateBillingInstruction = "update"
	PermissionMarkBillingReady         = "mark_ready"
	PermissionRevertBillingToDraft     = "revert_to_draft"
	PermissionCancelBillingInstruction = "cancel"
)

// Fiscal. Requesting an invoice creates a fiscal request record, hence create.
const PermissionRequestInvoice = "create"

// Pricelist. Promoting one list demotes another in the same breath, but the power is still an
// ordinary edit of the list's own flag.
const PermissionSetDefaultPricelist = "update"
