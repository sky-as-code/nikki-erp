package services

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
)

// InvoiceDomainService closes invoice drafts.
type InvoiceDomainService struct{}

func NewInvoiceDomainService() *InvoiceDomainService {
	return &InvoiceDomainService{}
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
		invoice, err := findInvoiceById(tranxCtx, cmd.InvoiceId)
		if err != nil {
			return err
		}
		if invoice == nil {
			appendFieldViolation(vErrs, models.InvoiceFieldId,
				"paymentinvoice.invoice_not_found", "no invoice with id '"+cmd.InvoiceId+"'")
			return nil
		}

		// The status is re-read inside the transaction rather than trusted from a prior read, so
		// two callers issuing the same draft at once cannot both pass this check.
		if status := derefString(invoice.GetStatus()); status != models.InvoiceStatusDraft {
			appendFieldViolation(vErrs, models.InvoiceFieldStatus,
				"paymentinvoice.invoice_not_draft",
				"only a draft may be issued; this invoice is '"+status+"'")
			return nil
		}

		totals, err := this.recomputeLines(tranxCtx, cmd.InvoiceId, vErrs)
		if err != nil || vErrs.Count() > 0 {
			return err
		}

		issuedAt := time.Now().UTC()
		number, err := allocateInvoiceNumber(tranxCtx, issuedAt.Year())
		if err != nil {
			return err
		}

		if err := writeInvoiceFields(tranxCtx, cmd.InvoiceId, dmodel.DynamicFields{
			models.InvoiceFieldNumber:         number,
			models.InvoiceFieldStatus:         models.InvoiceStatusIssued,
			models.InvoiceFieldIssuedAt:       issuedAt,
			models.InvoiceFieldSubtotalAmount: totals.Subtotal,
			models.InvoiceFieldTaxAmount:      totals.Tax,
			models.InvoiceFieldTotalAmount:    totals.Total,
		}); err != nil {
			return err
		}

		result = &IssueResult{
			InvoiceId:      cmd.InvoiceId,
			Number:         number,
			IssuedAt:       issuedAt,
			SubtotalAmount: totals.Subtotal,
			TaxAmount:      totals.Tax,
			TotalAmount:    totals.Total,
		}
		return nil
	})

	if err != nil || vErrs.Count() > 0 {
		return nil, vErrs, err
	}
	return result, vErrs, nil
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
		quantity := int64(0)
		if q := line.GetQuantity(); q != nil {
			quantity = int64(*q)
		}
		unitPrice := derefDecimal(line.GetUnitPrice())

		amount := unitPrice.Mul(decimal.NewFromInt(quantity))
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

// allocateInvoiceNumber mints the next number for the given year.
//
// The sequence is per-year and derived by counting the invoices already issued in it. That is
// sound only because this runs inside the issue transaction: the count and the write that follows
// it are one atomic step, so two invoices issued at the same instant cannot read the same count.
// The unique index on number is the backstop if that assumption ever breaks — a collision fails
// the transaction rather than producing two documents with one number.
func allocateInvoiceNumber(ctx corectx.Context, year int) (string, error) {
	engine, err := engineFor(models.InvoiceSchemaName)
	if err != nil {
		return "", err
	}

	prefix := fmt.Sprintf("INV-%d-", year)

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(models.InvoiceFieldNumber, dmodel.StartsWith, prefix),
	)
	// Highest number first: the next one follows the largest already taken, so a gap left by a
	// deleted invoice is not re-used. Re-using a number would put two documents in the record
	// under one identity.
	graph.OrderBy(models.InvoiceFieldNumber, dmodel.Desc)

	found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  1,
	})
	if err != nil {
		return "", errors.Wrap(err, "allocateInvoiceNumber")
	}

	next := 1
	if found != nil && found.HasData && len(found.Data.Items) > 0 {
		highest := derefString(models.NewInvoiceFrom(found.Data.Items[0]).GetNumber())
		if sequence, ok := sequenceOfInvoiceNumber(highest, prefix); ok {
			next = sequence + 1
		}
	}

	return fmt.Sprintf("%s%06d", prefix, next), nil
}

// sequenceOfInvoiceNumber reads the counter out of a number this module minted.
//
// A number that does not parse is treated as absent rather than failing the issue: it was not
// minted here, and refusing to issue anything until someone corrects a stray row would be a worse
// outcome than starting the sequence afresh alongside it. The unique index still prevents a
// collision.
func sequenceOfInvoiceNumber(number string, prefix string) (int, bool) {
	if len(number) <= len(prefix) || number[:len(prefix)] != prefix {
		return 0, false
	}

	sequence := 0
	if _, err := fmt.Sscanf(number[len(prefix):], "%d", &sequence); err != nil {
		return 0, false
	}
	return sequence, sequence > 0
}
