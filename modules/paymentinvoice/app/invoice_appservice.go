package app

import (
	"time"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/constants"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/services"
	itInvoice "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/invoice"
)

// paramInvoiceId names the invoice in the request path (":id/issue").
const paramInvoiceId = "id"

// The invoice's authorized surface: the built-in CRUD plus issue.
//
// Issuing carries its own permission rather than reusing "update": an issued invoice is an
// accounting document, and being allowed to correct a draft's note is not the same authority as
// being allowed to close one and mint its number.

type InvoiceApplicationServiceImpl struct {
	composable.CrudApplicationService
	invoices *services.InvoiceDomainService
}

func NewInvoiceApplicationService(
	base composable.CrudApplicationService, invoices *services.InvoiceDomainService,
) itInvoice.InvoiceCrudApplicationService {
	return &InvoiceApplicationServiceImpl{CrudApplicationService: base, invoices: invoices}
}

// IssueAction closes a draft invoice. It addresses one record, so the invoice must belong to the
// org the caller acts in.
func (this *InvoiceApplicationServiceImpl) IssueAction(
	ctx corectx.Context, cmd itInvoice.InvoiceActionCommand,
) (*itInvoice.InvoiceActionResult, error) {
	orgId, cErrs := this.AssertAction(ctx, constants.ActionIssue, cmd)
	if cErrs != nil {
		return &itInvoice.InvoiceActionResult{ClientErrors: *cErrs}, nil
	}
	if orgId != nil {
		cErrs, err := this.AssertRecordInOrg(ctx, cmd, *orgId)
		if err != nil {
			return nil, err
		}
		if cErrs != nil {
			return &itInvoice.InvoiceActionResult{ClientErrors: *cErrs}, nil
		}
	}

	result, vErrs, err := this.invoices.Issue(ctx, services.IssueCommand{
		InvoiceId: readString(cmd, paramInvoiceId),
	})
	if err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &itInvoice.InvoiceActionResult{ClientErrors: *vErrs}, nil
	}
	// The amounts cross the wire as strings so they survive JSON with the precision they were
	// computed at.
	return &itInvoice.InvoiceActionResult{HasData: true, Data: map[string]any{
		"invoice_id":      result.InvoiceId,
		"number":          result.Number,
		"issued_at":       result.IssuedAt.Format(time.RFC3339),
		"subtotal_amount": result.SubtotalAmount.String(),
		"tax_amount":      result.TaxAmount.String(),
		"total_amount":    result.TotalAmount.String(),
	}}, nil
}
