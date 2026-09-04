package dynamicengines

import (
	stdErr "errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
	itInvoicing "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external/invoicing"
)

// The request-invoice action hangs off the fiscal request, not the bill: asking for a VAT invoice
// is the power to create a legal document, not a power to split, merge or settle the bill. The bill
// travels in the body, and the action is collection-level because it creates the request.
//
// It ADJUSTS an issued document; it no longer raises the original. An original VAT invoice is asked
// for by putting a billing instruction on the order, and only the scheduled issuance job raises one
// — so an ISSUE_ORIGINAL sent here is refused rather than written, because a request this path
// wrote would sit `pending` for ever, unread by the job that now does the issuing.

const (
	PermissionRequestInvoice = "create"

	ActionRequestInvoice = "request_invoice"
)

// invoicingProvider is the bound eInvoice port. A package variable because an action callback is
// handed only its own engine and cannot reach the container. Nil is a supported value, not an
// uninitialised one: with no adapter, RequestInvoice writes the request and leaves it `pending`.
var invoicingProvider itInvoicing.InvoicingExtService

func SetInvoicingPort(provider itInvoicing.InvoicingExtService) {
	invoicingProvider = provider
}

func defineSalesFiscalRequestActions(engine drif.DynamicResourceEngine) error {
	return stdErr.Join(
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName: ActionRequestInvoice,
			ActionType: drif.ActionTypeGeneric,

			// Underscores, never hyphens: the route path must match the action code.
			RestPath:    "request_invoice",
			Permission:  PermissionRequestInvoice,
			MainProcess: processRequestInvoice,
		}),
	)
}

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
			"fiscal_document_request_id": result.FiscalDocumentRequestId,
			"sales_bill_id":              result.SalesBillId,

			// Returned so a caller does not read a successful call as an issued document; an
			// in-flight request is pending.
			"status": result.Status,

			// Empty unless the provider confirmed. The only durable link to the document.
			"provider_reference": result.ProviderReference,
			"already_existed":    result.AlreadyExisted,
		},
	}, nil
}

// readBuyerInfo reads the buyer's fiscal identity, nested under `buyer` rather than flattened so
// the wire shape matches the stored snapshot.
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

// readAnyString avoids a bare type assertion: a JSON round-trip can hand back a different concrete
// type, and a bare assertion panics the request.
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
