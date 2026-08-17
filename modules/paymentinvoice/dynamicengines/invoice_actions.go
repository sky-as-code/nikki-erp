package dynamicengines

import (
	"time"

	"go.bryk.io/pkg/errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/constants"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/services"
)

// paramInvoiceId names the invoice in the request path.
//
// The action is scoped to one invoice, so the id arrives as a path segment rather than in the body,
// which is the shape the engine's ":id" RestPath produces.
const paramInvoiceId = "id"

// defineInvoiceActions adds the issue action.
//
// Issuing carries its own permission rather than reusing "update" for the same reason refunding
// does: an issued invoice is an accounting document, and being allowed to correct a draft's note is
// not the same authority as being allowed to close one and mint its number.
func defineInvoiceActions(engine drif.DynamicResourceEngine) error {
	return engine.DefineAction(drif.DynamicActionDefinition{
		ActionName:  constants.ActionIssue,
		ActionType:  drif.ActionTypeGeneric,
		RestPath:    ":id/" + constants.ActionIssue,
		Permission:  constants.ActionIssue,
		MainProcess: processIssue,
	})
}

// processIssue closes a draft invoice.
func processIssue(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := requireInvoiceService()
	if err != nil {
		return nil, err
	}

	result, cErrs, err := service.Issue(ctx, services.IssueCommand{
		InvoiceId: readString(input.Params, paramInvoiceId),
	})
	if err != nil {
		return nil, err
	}
	if cErrs.Count() > 0 {
		return &drif.ActionResult{ClientErrors: *cErrs}, nil
	}

	// The amounts cross the wire as strings so they survive JSON with the precision they were
	// computed at: a float64 cannot hold every decimal, and an invoice total rounded in transit is
	// a document that disagrees with its own lines.
	return &drif.ActionResult{
		HasData: true,
		Data: map[string]any{
			"invoice_id":      result.InvoiceId,
			"number":          result.Number,
			"issued_at":       result.IssuedAt.Format(time.RFC3339),
			"subtotal_amount": result.SubtotalAmount.String(),
			"tax_amount":      result.TaxAmount.String(),
			"total_amount":    result.TotalAmount.String(),
		},
	}, nil
}

// invoiceService is the domain service the issue action delegates to. Like the order service it is
// a package variable rather than a derived resource service: issuing is not CRUD on an invoice.
var invoiceService *services.InvoiceDomainService

// SetInvoiceService installs the service the invoice action delegates to. Init calls it before any
// request is served.
func SetInvoiceService(service *services.InvoiceDomainService) {
	invoiceService = service
}

func requireInvoiceService() (*services.InvoiceDomainService, error) {
	if invoiceService == nil {
		return nil, errors.New(
			"the invoice domain service was not installed; PaymentInvoiceModule.Init must call " +
				"dynamicengines.SetInvoiceService")
	}
	return invoiceService, nil
}
