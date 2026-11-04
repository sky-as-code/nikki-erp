package services

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
)

// InvoiceDomainService closes invoice drafts.
type InvoiceDomainService struct {
	numbering *InvoiceNumbering
}

func NewInvoiceDomainService(numbering *InvoiceNumbering) *InvoiceDomainService {
	return &InvoiceDomainService{numbering: numbering}
}

// IssueCommand asks for a draft to be closed. It carries only the invoice, because everything the
// issued document says is computed from what is already recorded against it.
type IssueCommand struct {
	InvoiceId string
}

// IssueResult is the issued document's identity and its frozen totals.
type IssueResult struct {
	InvoiceId string
	Number    string
	IssuedAt  time.Time

	SubtotalAmount decimal.Decimal
	TaxAmount      decimal.Decimal
	TotalAmount    decimal.Decimal
}

// invoiceLinePageSize bounds how many lines one invoice may carry.
//
// An invoice with more lines than this is a data-entry accident rather than a real document, and
// silently totalling only the first page would produce a document that understates what is owed.
// Exceeding it is refused instead.
const invoiceLinePageSize = 500

// Issue closes a draft: it recomputes the totals from the lines, assigns the number and stamps the
// date.
//
// Everything happens in one transaction, and the ordering matters. The totals are recomputed here
// rather than trusted from whatever a client last wrote, because an issued invoice is an accounting
// document: if the stored total and the lines disagree, the lines are what a reader can verify, so
// the lines win.
//
// Only a draft may be issued. Re-issuing would mint a second number for one document and re-freeze
// totals that may since have been paid against, so an invoice already issued is refused as a client
// error naming its current status.
func (this *InvoiceDomainService) Issue(
	ctx corectx.Context, cmd IssueCommand,
) (*IssueResult, *ft.ClientErrors, error) {
	vErrs := ft.NewClientErrors()

	if cmd.InvoiceId == "" {
		appendFieldViolation(vErrs, models.InvoiceFieldId,
			"paymentinvoice.invoice_required", "no invoice was identified")
		return nil, vErrs, nil
	}

	var result *IssueResult
	err := withInvoiceTransaction(ctx, func(tranxCtx corectx.Context) error {
		issued, innerErrs, err := this.issueWithin(tranxCtx, cmd.InvoiceId)
		if err != nil {
			return err
		}
		if innerErrs != nil && innerErrs.Count() > 0 {
			// Rolled back rather than committed: a refusal from the numbering arrives after its
			// reservation is written, and that reservation must not survive.
			vErrs.ConcatPtr(innerErrs)
			return errRollbackAfterRefusal
		}
		result = issued
		return nil
	})
	if errors.Is(err, errRollbackAfterRefusal) {
		err = nil
	}

	if err != nil || vErrs.Count() > 0 {
		return nil, vErrs, err
	}
	return result, vErrs, nil
}

// issueWithin closes a draft inside a transaction the caller already opened.
//
// Split out of Issue so that raising a document from a source can reuse it: the draft, its lines and
// the number have to be one atomic step, and a second copy of the closing logic would be a second
// place for the totals or the numbering to go wrong.
func (this *InvoiceDomainService) issueWithin(
	ctx corectx.Context, invoiceId string,
) (*IssueResult, *ft.ClientErrors, error) {
	vErrs := ft.NewClientErrors()

	invoice, err := findInvoiceById(ctx, invoiceId)
	if err != nil {
		return nil, nil, err
	}
	if invoice == nil {
		appendFieldViolation(vErrs, models.InvoiceFieldId,
			"paymentinvoice.invoice_not_found", "no invoice with id '"+invoiceId+"'")
		return nil, vErrs, nil
	}

	// The status is re-read inside the transaction rather than trusted from a prior read, so two
	// callers issuing the same draft at once cannot both pass this check.
	if status := derefString(invoice.GetStatus()); status != models.InvoiceStatusDraft {
		appendFieldViolation(vErrs, models.InvoiceFieldStatus,
			"paymentinvoice.invoice_not_draft",
			"only a draft may be issued; this invoice is '"+status+"'")
		return nil, vErrs, nil
	}

	totals, err := this.recomputeLines(ctx, invoiceId, vErrs)
	if err != nil || vErrs.Count() > 0 {
		return nil, vErrs, err
	}

	issued, numberErrs, err := this.allocateNumber(ctx, invoice, invoiceId)
	if err != nil || numberErrs != nil {
		return nil, numberErrs, err
	}
	number, issuedAt := issued.Number, issued.IssuedAt

	if err := writeInvoiceFields(ctx, invoiceId, dmodel.DynamicFields{
		models.InvoiceFieldNumber:         number,
		models.InvoiceFieldStatus:         models.InvoiceStatusIssued,
		models.InvoiceFieldIssuedAt:       issuedAt,
		models.InvoiceFieldSubtotalAmount: totals.Subtotal,
		models.InvoiceFieldTaxAmount:      totals.Tax,
		models.InvoiceFieldTotalAmount:    totals.Total,
	}); err != nil {
		return nil, nil, err
	}

	return &IssueResult{
		InvoiceId:      invoiceId,
		Number:         number,
		IssuedAt:       issuedAt,
		SubtotalAmount: totals.Subtotal,
		TaxAmount:      totals.Tax,
		TotalAmount:    totals.Total,
	}, vErrs, nil
}

// firstViolationMessage flattens a refusal into the one line a calling module can show. The full set
// stays here: a caller reports why a document could not be raised, not this module's whole rule book.
func firstViolationMessage(vErrs *ft.ClientErrors) string {
	if vErrs == nil || vErrs.Count() == 0 {
		return "the invoice was refused"
	}
	for _, violation := range *vErrs {
		if message := violation.String(); message != "" {
			return message
		}
	}
	return "the invoice was refused"
}

// invoiceTotals is what the lines of one invoice come to.
type invoiceTotals struct {
	Subtotal decimal.Decimal
	Tax      decimal.Decimal
	Total    decimal.Decimal
}

// recomputeLines rewrites each line's amount and returns what the invoice comes to.
//
// Each line's amount is recomputed as quantity × unit_price rather than trusted, for the same
// reason the invoice totals are: the quantity and the price are what a reader can check, and a
// stored amount that disagrees with them is the field that is wrong.
//
// Tax is accumulated per line, not applied to the subtotal, because lines may carry different
// rates — a single rate over the subtotal would silently be wrong for any invoice that mixes them.
func (this *InvoiceDomainService) recomputeLines(
	ctx corectx.Context, invoiceId string, vErrs *ft.ClientErrors,
) (invoiceTotals, error) {
	lines, err := findInvoiceLines(ctx, invoiceId)
	if err != nil {
		return invoiceTotals{}, err
	}

	if len(lines) == 0 {
		// An invoice with no lines totals zero, which is not a document anyone meant to issue.
		appendFieldViolation(vErrs, models.InvoiceFieldId,
			"paymentinvoice.invoice_has_no_lines",
			"an invoice must have at least one line before it can be issued")
		return invoiceTotals{}, nil
	}
	if len(lines) >= invoiceLinePageSize {
		appendFieldViolation(vErrs, models.InvoiceFieldId,
			"paymentinvoice.invoice_too_many_lines",
			fmt.Sprintf("an invoice may carry at most %d lines", invoiceLinePageSize-1))
		return invoiceTotals{}, nil
	}

	totals := invoiceTotals{
		Subtotal: decimal.Zero,
		Tax:      decimal.Zero,
	}

	for _, line := range lines {
		quantity := derefDecimal(line.GetQuantity())
		unitPrice := derefDecimal(line.GetUnitPrice())

		amount := unitPrice.Mul(quantity)
		lineTax := amount.Mul(derefDecimal(line.GetTaxRatePercent())).Div(decimal.NewFromInt(100))

		totals.Subtotal = totals.Subtotal.Add(amount)
		totals.Tax = totals.Tax.Add(lineTax)

		// The line is written back only when the stored amount disagrees, so issuing an invoice
		// whose lines are already correct does not touch every row.
		if !amount.Equal(derefDecimal(line.GetAmount())) {
			if err := writeInvoiceLineFields(ctx, derefString(line.GetId()), dmodel.DynamicFields{
				models.InvoiceLineFieldAmount: amount,
			}); err != nil {
				return invoiceTotals{}, err
			}
		}
	}

	totals.Total = totals.Subtotal.Add(totals.Tax)
	return totals, nil
}

// allocateNumber issues the invoice's number from its organization's paymentinvoice.invoice
// sequence, inside the issue transaction ctx carries. The per-organization partial unique on
// (org_id, number) stays as the backstop: a collision fails the transaction rather than producing
// two documents with one number.
func (this *InvoiceDomainService) allocateNumber(
	ctx corectx.Context, invoice *models.Invoice, invoiceId string,
) (*IssuedNumber, *ft.ClientErrors, error) {
	if this.numbering == nil {
		return nil, nil, errors.New("allocateNumber: invoice numbering is not installed")
	}
	orgId := invoice.GetOrgId()
	if orgId == nil {
		return nil, nil, errors.Errorf("allocateNumber: invoice %s names no organization", invoiceId)
	}
	return this.numbering.Next(ctx, *orgId, invoiceId)
}
