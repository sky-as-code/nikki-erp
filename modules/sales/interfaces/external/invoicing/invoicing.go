// Package invoicing is Sales' port onto the issuing of legal fiscal documents.
//
// # Why this is its own package rather than another file in interfaces/external
//
// The other ports name the module on the far side - paymentinvoice's payment methods, accounting's
// tax engine. This one must name NOTHING on the far side (BR 48). A provider's name appearing in a
// type here would spread through every caller, and swapping providers - a routine commercial event -
// would become a change across the module. The package is the boundary that makes the omission
// enforceable rather than merely intended.
//
// # Sales supplies FACTS, the provider decides the DOCUMENT (BR 50)
//
// Sales knows what commercially happened: this sale was made, these items came back, these were the
// amounts and the tax allocations, at this moment, for this reason. It does not know, and must never
// decide, whether that requires an invoice, a credit note or an adjustment declaration. That is
// invoice law, it differs by jurisdiction and it changes, and BR 46 and BR 94.26 put it on the far
// side of this port precisely so that a change in it is not a change in Sales.
//
// So Intent carries ISSUE_ORIGINAL / ADJUST_FOR_FULL_RETURN / ADJUST_FOR_PARTIAL_RETURN /
// ADJUST_PRICE - four things that happened - and there is no field anywhere below for a document
// type, a serial, a template or a tax authority code.
//
// # This port is DECLARED but NOT YET BOUND
//
// There is no eInvoice provider adapter in this repository, and no module publishes this capability.
// Writing one would mean inventing a provider's protocol, and the wrong protocol is worse than none:
// it would look bound while issuing nothing.
//
// Until a real adapter lands in infra/external/invoicing/, sales_fiscal_requests rows are written
// and stay `pending`, and the order's invoice_status stays `requested`. That is the honest state -
// the customer really has asked for a VAT invoice and no provider has yet answered - and it is
// exactly the state BR 77 requires while a request is in flight, so nothing downstream has to
// special-case it. See the SALES-030 note in 02-progress.md.
package invoicing

import (
	"github.com/shopspring/decimal"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// FiscalIntent is what commercially happened, never what document to produce.
type FiscalIntent string

const (
	// IntentIssueOriginal: a sale was made and needs its VAT invoice.
	IntentIssueOriginal = FiscalIntent("ISSUE_ORIGINAL")

	// IntentAdjustForFullReturn: the whole sale came back.
	IntentAdjustForFullReturn = FiscalIntent("ADJUST_FOR_FULL_RETURN")

	// IntentAdjustForPartialReturn: some of it came back. Separate from the full return because the
	// provider may treat them as different documents, and that judgement is the provider's - Sales
	// reports which happened and does not choose between the outcomes.
	IntentAdjustForPartialReturn = FiscalIntent("ADJUST_FOR_PARTIAL_RETURN")

	// IntentAdjustPrice: the amounts changed without goods moving.
	IntentAdjustPrice = FiscalIntent("ADJUST_PRICE")
)

// InvoicingExtService is the capability Sales needs from the eInvoice provider.
//
// Narrow on purpose. There is no method for cancelling a document, editing one, or fetching a PDF:
// the first two are the provider's own regulated workflows, and Sales having a way to call them
// would make it possible to void a legal document from a till.
type InvoicingExtService interface {
	// Issue asks for a fiscal document and returns what the provider decided.
	//
	// The IdempotencyKey on the request must reach the provider. This is the whole reason the
	// method takes a request struct rather than the fields: after a timeout Sales cannot tell
	// whether the document was issued, and only a key the provider recognises makes the retry safe.
	Issue(ctx corectx.Context, request IssueRequest) (*IssueResult, error)

	// GetStatus asks what became of a request Sales already sent.
	//
	// The recovery path for the case that has no local answer: the call timed out, so the row says
	// `pending`, and only the provider knows whether a document exists. Without this, a timeout
	// would leave Sales permanently unable to tell a pending request from an issued one.
	GetStatus(ctx corectx.Context, idempotencyKey string) (*IssueResult, error)
}

// IssueRequest is one fiscal event, stated in commercial terms.
type IssueRequest struct {
	// IdempotencyKey travels to the provider. See the note on Issue.
	IdempotencyKey string

	Intent FiscalIntent

	// SalesFiscalRequestId and SalesBillId let the provider echo back what it was asked and by whom.
	// Opaque to the provider: it never resolves them against Sales.
	SalesFiscalRequestId string
	SalesBillId          string

	// OriginalProviderReference names the document being adjusted, and is empty only on
	// IntentIssueOriginal. An adjustment without it is not a document the provider can produce - a
	// credit note that credits nothing - so the port carries it rather than leaving the adapter to
	// look it up.
	OriginalProviderReference string

	Buyer BuyerInfo

	CurrencyCode string

	// Subtotal, TaxTotal and TotalAmount are the HISTORICAL amounts of the transaction, never
	// recomputed from current prices (BR 54, BR 60, acceptance BR 94.22). A document reissued from
	// today's prices would state a sum the customer never paid.
	Subtotal    decimal.Decimal
	TaxTotal    decimal.Decimal
	TotalAmount decimal.Decimal

	Lines []IssueLine

	// TaxSnapshot is accounting's immutable record of how the tax was determined - the rates and
	// configuration versions in force and the rounding policy applied. Passed through verbatim and
	// never interpreted here, so that the provider can state what was actually charged rather than
	// what today's tax master would charge.
	TaxSnapshot any

	// Reason and OccurredAt are the facts a provider needs to justify an adjustment. Free text
	// rather than an enum: the reasons are the business's, and constraining them here would mean
	// Sales deciding which ones the law accepts.
	Reason     string
	OccurredAt int64
}

// BuyerInfo is the buyer's fiscal identity as supplied at issuance.
//
// A snapshot, not a reference. A later change of company name or address must not rewrite what a
// historical invoice said (BR 87.7).
type BuyerInfo struct {
	// TaxCode is what makes an invoice deductible for a business buyer, and is the field most often
	// missing or wrong. Its completeness is checked before the request is written, so a bad one
	// fails as a correctable 400 rather than as a provider rejection hours later.
	TaxCode string

	LegalName string
	Address   string
	Email     string
}

// IssueLine is one line of the transaction, in historical amounts.
type IssueLine struct {
	// SalesOrderLineId is echoed back so Sales can attribute what the provider reports to the right
	// line without matching on description and amount, which are not unique within a sale.
	SalesOrderLineId string

	Description string
	Quantity    decimal.Decimal
	UomId       string

	UnitAmount decimal.Decimal
	NetAmount  decimal.Decimal
	TaxAmount  decimal.Decimal

	// TaxRateSnapshot is a FRACTION - 0.1 for 10% - matching sales_order_lines.tax_rate_snapshot
	// and the accounting boundary conversion. The unit is stated here because a provider API taking
	// a percentage is the likeliest place for a silent 100x error.
	TaxRateSnapshot decimal.Decimal
}

// IssueResult is what the provider answered.
type IssueResult struct {
	// Issued says the document exists. FALSE IS NOT AN ERROR: a provider is unreachable,
	// rate-limited or slow in normal operation, and a Go error is reserved for the case where Sales
	// itself is broken.
	Issued bool

	// ProviderReference is the provider's identifier for the document, and the only durable link to
	// it. Stored on confirmation and never regenerated.
	ProviderReference string

	// FailureReason explains a refusal in the provider's words. An operator has to decide whether to
	// retry, correct the buyer information or escalate, and "failed" alone tells them none of that.
	FailureReason string

	// IssuedAt is when the provider says the document came into existence, which is not when Sales
	// asked and not when this reply was read. It is the date the document bears.
	IssuedAt int64
}
