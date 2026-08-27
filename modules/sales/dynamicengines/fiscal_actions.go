package dynamicengines

import (
	stdErr "errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
	itInvoicing "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external/invoicing"
)

// The request-invoice action (BR 77, SALES-031).
//
// # It hangs off the FISCAL REQUEST, not the bill
//
// The natural reading puts it on sales_bill - a bill is what gets invoiced. But the permission then
// has to be a power over the bill, and asking for a VAT invoice is not a power over the settlement:
// a customer who paid at a till and afterwards wants an invoice for their company is served by
// somebody who may create a fiscal document, and who need not be able to split, merge or settle the
// bill they are invoicing.
//
// So the permission is `create` on `sales_fiscal_request` - the power to ask for a legal document -
// and the bill travels in the body. The action is collection-level for the same reason: it CREATES
// the request rather than acting on an existing one.

const (
	// PermissionRequestInvoice is `create` on sales_fiscal_request: the power to ask a tax authority,
	// through a provider, for a legal document. Deliberately not a power over the bill - see above.
	PermissionRequestInvoice = "create"

	ActionRequestInvoice = "request_invoice"
)

// invoicingProvider is the bound eInvoice port, or nil when no adapter ships in this build.
//
// A package variable for the same reason the payment-method port is one: an action callback is
// handed only its own engine, so the container cannot reach it. NIL IS A SUPPORTED VALUE, not an
// uninitialised one - there is no provider adapter in this repository, and RequestInvoice writes the
// request and leaves it `pending`, which is exactly BR 77's in-flight state.
var invoicingProvider itInvoicing.InvoicingExtService

// SetInvoicingPort binds the eInvoice provider, if a build ships one.
func SetInvoicingPort(provider itInvoicing.InvoicingExtService) {
	invoicingProvider = provider
}

// defineSalesFiscalRequestActions adds request_invoice to the fiscal request engine.
func defineSalesFiscalRequestActions(engine drif.DynamicResourceEngine) error {
	return stdErr.Join(
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName: ActionRequestInvoice,
			ActionType: drif.ActionTypeGeneric,

			// Collection-level: the operation creates the request, so there is no record to hang it
			// off. Underscores, never hyphens - the route path must match the action code.
			RestPath:    "request_invoice",
			Permission:  PermissionRequestInvoice,
			MainProcess: processRequestInvoice,
		}),
	)
}

// processRequestInvoice asks the provider for a VAT invoice (BR 77, SALES-031).
func processRequestInvoice(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	result, vErrs, err := services.RequestInvoice(ctx, services.RequestInvoiceParams{
		SalesBillId:             readStringParam(input.Params, "sales_bill_id"),
		Intent:                  readStringParam(input.Params, "intent"),
		OriginalFiscalRequestId: readStringParam(input.Params, "original_fiscal_request_id"),
		Reason:                  readStringParam(input.Params, "reason"),
		IdempotencyKey:          readStringParam(input.Params, "idempotency_key"),
		Buyer:                   readBuyerInfo(input.Params),
	}, invoicingProvider)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &drif.ActionResult{ClientErrors: *vErrs}, nil
	}

	return &drif.ActionResult{
		HasData: true,
		Data: map[string]any{
			// The name BR 77 gives the output.
			"fiscal_document_request_id": result.FiscalDocumentRequestId,
			"sales_bill_id":              result.SalesBillId,

			// Returned so a caller can see that a request in flight is PENDING rather than assuming
			// a successful call means an issued document (BR 77).
			"status": result.Status,

			// Empty unless the provider confirmed. The only durable link to the document.
			"provider_reference": result.ProviderReference,
			"already_existed":    result.AlreadyExisted,
		},
	}, nil
}

// readBuyerInfo pulls the buyer's fiscal identity out of the request body.
//
// Nested under `buyer` rather than flattened, because it is stored as one snapshot and reading it
// as a unit keeps the wire shape and the stored shape the same.
func readBuyerInfo(params map[string]any) itInvoicing.BuyerInfo {
	nested, ok := params["buyer"].(map[string]any)
	if !ok {
		return itInvoicing.BuyerInfo{}
	}
	return itInvoicing.BuyerInfo{
		TaxCode:   readAnyString(nested, "tax_code"),
		LegalName: readAnyString(nested, "legal_name"),
		Address:   readAnyString(nested, "address"),
		Email:     readAnyString(nested, "email"),
	}
}

// readAnyString reads a string out of an untyped map without asserting bare, for the same reason
// readStringParam does not: a JSON round-trip can hand back a different concrete type, and a bare
// assertion panics the request.
func readAnyString(values map[string]any, field string) string {
	value, ok := values[field]
	if !ok || value == nil {
		return ""
	}
	if typed, ok := value.(string); ok {
		return typed
	}
	if typed, ok := value.(*string); ok && typed != nil {
		return *typed
	}
	return ""
}
