// Package invoicing binds Sales' fiscal-document port to whoever issues the documents.
//
// Today that is the Payment & Invoice module's own invoice engine, which produces an internal
// document numbered INV-{year}-{sequence}. That is deliberately swappable: the port names nothing on
// the far side, so pointing Sales at a real e-invoice provider is a change to this file and nothing
// else in the module.
package invoicing

import (
	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itInvoice "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/invoice"

	itInvoicing "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external/invoicing"
)

// sourceTypeFiscalRequest names what these documents are raised for, and is half the replay key.
//
// It is the fiscal request rather than the bill, because the request is what carries the
// idempotency key: one bill can legitimately need a second document later — an adjustment after a
// return — and keying on the bill would make that look like a replay of the first.
const sourceTypeFiscalRequest = "sales_fiscal_request"

// hundred converts a tax rate between the two units either side of this boundary. See the note in
// issueLinesOf.
var hundred = decimal.NewFromInt(100)

// NewAdapter binds the port onto the invoice engine.
func NewAdapter(invoices itInvoice.InvoiceDomainService) itInvoicing.InvoicingExtService {
	return &adapter{invoices: invoices}
}

type adapter struct {
	invoices itInvoice.InvoiceDomainService
}

// Issue raises the document for one fiscal request.
func (this *adapter) Issue(
	ctx corectx.Context, request itInvoicing.IssueRequest,
) (*itInvoicing.IssueResult, error) {
	sourceId := request.IdempotencyKey
	if sourceId == "" {
		// Without a key a retry would mint a second document for one sale. Refused rather than
		// defaulted: choosing a key here would be choosing which retries are safe.
		return &itInvoicing.IssueResult{
			Issued:        false,
			FailureReason: "no idempotency key was supplied, so issuing could not be made safe to retry",
		}, nil
	}

	raised, err := this.invoices.IssueFromSource(ctx, itInvoice.IssueFromSourceCommand{
		SourceType: sourceTypeFiscalRequest,
		SourceId:   sourceId,
		OrgId:      request.OrgId,
		Partner: itInvoice.PartnerInfo{
			Name:    request.Buyer.LegalName,
			TaxCode: request.Buyer.TaxCode,
			Address: request.Buyer.Address,
		},
		Lines: issueLinesOf(request),
		Note:  request.Reason,
	})
	if err != nil {
		return nil, errors.Wrap(err, "raising a fiscal document")
	}

	if raised.Refused || !raised.HasData {
		reason := raised.RefusalReason
		if reason == "" {
			reason = "the document was not raised"
		}
		return &itInvoicing.IssueResult{Issued: false, FailureReason: reason}, nil
	}

	// THE TOTALS ARE CHECKED, NOT TRUSTED, AND A MISMATCH REFUSES THE DOCUMENT.
	//
	// The engine recomputes an invoice from its lines and lets the lines win, which is right for a
	// document it owns. But Sales sends historical amounts — what the customer actually paid, after
	// promotions and a largest-remainder split across the bill — and a document restating them from
	// quantity × price would claim a sum nobody was ever charged.
	//
	// So a disagreement is reported rather than accepted. Confirming a document that says something
	// different from the sale would be the worse outcome: the row would look issued and the paper
	// would be wrong.
	if !raised.Data.TotalAmount.Equal(request.TotalAmount) {
		return &itInvoicing.IssueResult{
			Issued: false,
			FailureReason: "the issued document totals " + raised.Data.TotalAmount.String() +
				" but the sale came to " + request.TotalAmount.String() +
				"; the document was not confirmed",
		}, nil
	}

	return &itInvoicing.IssueResult{
		Issued:            true,
		ProviderReference: raised.Data.Number,
		IssuedAt:          raised.Data.IssuedAt.Unix(),
	}, nil
}

// GetStatus asks what became of a request Sales already sent, which is the recovery path after a
// timeout: the row says pending and only the far side knows whether a document exists.
func (this *adapter) GetStatus(
	ctx corectx.Context, idempotencyKey string,
) (*itInvoicing.IssueResult, error) {
	if idempotencyKey == "" {
		return &itInvoicing.IssueResult{Issued: false}, nil
	}

	found, err := this.invoices.GetBySource(ctx, itInvoice.GetBySourceQuery{
		SourceType: sourceTypeFiscalRequest,
		SourceId:   idempotencyKey,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "reading the document for request '%s'", idempotencyKey)
	}
	if found == nil || !found.HasData {
		// No document. Not an error and not a failure: it means the earlier attempt never landed,
		// so the request may be retried.
		return &itInvoicing.IssueResult{Issued: false}, nil
	}

	return &itInvoicing.IssueResult{
		Issued:            true,
		ProviderReference: found.Data.Number,
		IssuedAt:          found.Data.IssuedAt.Unix(),
	}, nil
}

// issueLinesOf translates the sale's lines into the engine's.
//
// THE TAX RATE CHANGES UNITS HERE, AND THIS IS THE ONLY PLACE IT DOES. Sales carries a fraction —
// 0.1 for 10%, matching sales_order_lines.tax_rate_snapshot and the accounting boundary — while the
// invoice line carries a percentage. Both sides document their unit; the conversion is one
// multiplication, kept alone in one function so a reader can see it happens exactly once. Getting it
// wrong is a silent hundredfold error on a legal document, which is why it has its own test.
func issueLinesOf(request itInvoicing.IssueRequest) []itInvoice.IssueFromSourceLine {
	lines := make([]itInvoice.IssueFromSourceLine, 0, len(request.Lines))
	for _, line := range request.Lines {
		lines = append(lines, itInvoice.IssueFromSourceLine{
			Description:    line.Description,
			Quantity:       line.Quantity,
			UnitPrice:      line.UnitAmount,
			TaxRatePercent: line.TaxRateSnapshot.Mul(hundred),
		})
	}
	return lines
}
