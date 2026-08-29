// Package invoicing is Sales' port onto the issuing of legal fiscal documents. Unlike the other
// ports it names nothing on the far side: a provider's name in a type here would spread through
// every caller, so swapping providers would become a change across the module. Its own package is
// what makes that omission enforceable.
//
// Sales supplies facts and the provider decides the document. Sales knows what commercially
// happened; it must never decide whether that requires an invoice, a credit note or an adjustment
// declaration, which is invoice law and differs by jurisdiction. So Intent names four things that
// happened, and there is no field below for a document type, serial, template or tax authority code.
//
// The port is declared but NOT YET BOUND — no adapter exists. Until one lands in
// infra/external/invoicing/, sales_fiscal_requests rows are written and stay `pending` and the
// order's invoice_status stays `requested`, which is the state a request in flight is meant to have,
// so nothing downstream special-cases it.
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
	// provider may treat them as different documents, and that judgement is the provider's.
	IntentAdjustForPartialReturn = FiscalIntent("ADJUST_FOR_PARTIAL_RETURN")

	// IntentAdjustPrice: the amounts changed without goods moving.
	IntentAdjustPrice = FiscalIntent("ADJUST_PRICE")
)

// InvoicingExtService is the capability Sales needs from the eInvoice provider. There is no method
// for cancelling or editing a document — those are the provider's own regulated workflows, and a way
// to call them would make it possible to void a legal document from a till.
type InvoicingExtService interface {
	// Issue asks for a fiscal document. The request's IdempotencyKey must reach the provider: after
	// a timeout Sales cannot tell whether the document was issued, and only a key the provider
	// recognises makes the retry safe.
	Issue(ctx corectx.Context, request IssueRequest) (*IssueResult, error)

	// GetStatus asks what became of a request Sales already sent. It is the recovery path after a
	// timeout, where the row says `pending` and only the provider knows whether a document exists.
	GetStatus(ctx corectx.Context, idempotencyKey string) (*IssueResult, error)
}

// IssueRequest is one fiscal event, stated in commercial terms.
type IssueRequest struct {
	// IdempotencyKey travels to the provider. See the note on Issue.
	IdempotencyKey string

	Intent FiscalIntent

	// Opaque to the provider: it echoes them back and never resolves them against Sales.
	SalesFiscalRequestId string
	SalesBillId          string

	// OriginalProviderReference names the document being adjusted, empty only on
	// IntentIssueOriginal. An adjustment without it is a credit note that credits nothing.
	OriginalProviderReference string

	Buyer BuyerInfo

	CurrencyCode string

	// The historical amounts of the transaction, never recomputed from current prices: a document
	// reissued from today's prices would state a sum the customer never paid.
	Subtotal    decimal.Decimal
	TaxTotal    decimal.Decimal
	TotalAmount decimal.Decimal

	Lines []IssueLine

	// TaxSnapshot is accounting's immutable record of how the tax was determined. Passed through
	// verbatim and never interpreted here, so the provider states what was actually charged rather
	// than what today's tax master would charge.
	TaxSnapshot any

	// Reason and OccurredAt justify an adjustment. Free text rather than an enum, because
	// constraining the reasons here would mean Sales deciding which ones the law accepts.
	Reason     string
	OccurredAt int64
}

// BuyerInfo is the buyer's fiscal identity as supplied at issuance — a snapshot, not a reference: a
// later change of company name or address must not rewrite what a historical invoice said.
type BuyerInfo struct {
	// TaxCode makes an invoice deductible for a business buyer. Its completeness is checked before
	// the request is written, so a bad one fails as a correctable 400 rather than a provider
	// rejection hours later.
	TaxCode string

	LegalName string
	Address   string
	Email     string
}

// IssueLine is one line of the transaction, in historical amounts.
type IssueLine struct {
	// SalesOrderLineId is echoed back so Sales attributes what the provider reports to the right
	// line without matching on description and amount, which are not unique within a sale.
	SalesOrderLineId string

	Description string
	Quantity    decimal.Decimal
	UomId       string

	UnitAmount decimal.Decimal
	NetAmount  decimal.Decimal
	TaxAmount  decimal.Decimal

	// TaxRateSnapshot is a FRACTION — 0.1 for 10% — matching sales_order_lines.tax_rate_snapshot and
	// the accounting boundary conversion. A provider API taking a percentage is the likeliest place
	// for a silent 100x error.
	TaxRateSnapshot decimal.Decimal
}

// IssueResult is what the provider answered.
type IssueResult struct {
	// Issued says the document exists. False is not an error: a provider is unreachable,
	// rate-limited or slow in normal operation, and a Go error means Sales itself is broken.
	Issued bool

	// ProviderReference is the only durable link to the document. Stored on confirmation and never
	// regenerated.
	ProviderReference string

	// FailureReason explains a refusal in the provider's words: "failed" alone would not tell an
	// operator whether to retry, correct the buyer information or escalate.
	FailureReason string

	// IssuedAt is when the provider says the document came into existence — the date the document
	// bears, not when Sales asked or when this reply was read.
	IssuedAt int64
}
