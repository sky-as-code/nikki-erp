package services

import (
	"time"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	itInvoicing "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external/invoicing"
)

// Requesting a VAT invoice (BR 46-50, 58, 77, 87.7, SALES-030 and SALES-031).
//
// # The row is written BEFORE the provider is called, and that ordering is the design
//
// The obvious shape - call the provider, then store what came back - loses a document on the one
// failure that matters. If the call times out after the provider issued the invoice, Sales has no
// row, no idempotency key and no way to find out; the customer holds a VAT invoice the system does
// not know exists, and the next attempt issues a second one.
//
// So the request is persisted as `pending` first, with its idempotency key, and only then dispatched.
// A crash anywhere after that leaves a row an operator can reconcile through GetStatus. The cost is
// rows for requests that never left, which is the cheaper failure by a wide margin: a duplicate VAT
// invoice is a tax filing to correct.
//
// # PENDING IS NOT ISSUED (BR 77)
//
// invoice_status moves to `issued` only on confirmed provider success, and provider_reference - the
// only durable link to the document - is stored at that same moment. A request in flight leaves the
// order at `requested`. Reporting otherwise would tell a customer they hold an invoice they cannot
// deduct.
//
// # Sales validates what is ITS to validate
//
// The bill is eligible, no successful request already exists for it, and the buyer's fiscal
// information is complete. Not: whether the tax code is genuine, whether the document type is right,
// whether the serial is in range. Those are the provider's, and BR 46 and BR 94.26 keep them there.

// RequestInvoiceParams is what asking for a VAT invoice needs.
type RequestInvoiceParams struct {
	SalesBillId string

	// Intent defaults to ISSUE_ORIGINAL. The adjustment intents are supplied by the return workflow
	// (Phase 6), which knows what came back; a till asking for an invoice never sets it.
	Intent string

	// OriginalFiscalRequestId names the request being adjusted. Required for every intent except
	// ISSUE_ORIGINAL: an adjustment document that names no original is not a document a provider can
	// produce (BR 58).
	OriginalFiscalRequestId string

	Buyer itInvoicing.BuyerInfo

	// Reason justifies an adjustment, in the business's words.
	Reason string

	// IdempotencyKey is supplied by the caller when it has one - a till retrying its own request.
	// Generated when empty, which is correct for a first call and wrong for a retry, so a caller
	// that retries must send its key.
	IdempotencyKey string
}

// RequestInvoiceResult is what asking produced.
type RequestInvoiceResult struct {
	// FiscalDocumentRequestId is what BR 77 says the operation returns.
	FiscalDocumentRequestId string

	SalesBillId string

	// Status is `pending` whenever the provider has not confirmed, which currently includes every
	// call, since no adapter is bound. See the package comment on interfaces/external/invoicing.
	Status string

	// ProviderReference is set only alongside status `issued`, never before.
	ProviderReference string

	// AlreadyExisted marks the replay path: a second call with the same idempotency key returns the
	// first call's request rather than raising another.
	AlreadyExisted bool
}

// The refusal reasons requesting an invoice can produce.
const (
	ReasonBillNotEligibleForInvoice = "sales_fiscal_request.bill_not_eligible"
	ReasonInvoiceAlreadyRequested   = "sales_fiscal_request.already_requested"
	ReasonBuyerInfoIncomplete       = "sales_fiscal_request.buyer_info_incomplete"
	ReasonOriginalRequestRequired   = "sales_fiscal_request.original_request_required"
	ReasonOriginalRequestNotIssued  = "sales_fiscal_request.original_request_not_issued"
	ReasonUnknownFiscalIntent       = "sales_fiscal_request.unknown_intent"
)

// RequestInvoice asks the eInvoice provider for a fiscal document (BR 77).
//
// The invoicing port may be nil, and that is a supported state rather than a bug: no adapter is
// bound yet. The request is written and left `pending`, which is exactly what BR 77 requires of a
// request in flight, so no caller has to distinguish "not dispatched" from "not answered".
func RequestInvoice(
	ctx corectx.Context,
	params RequestInvoiceParams,
	provider itInvoicing.InvoicingExtService,
) (*RequestInvoiceResult, *ft.ClientErrors, error) {
	bill, err := loadRecord(ctx,
		models.SalesBillSchemaName, models.SalesBillFieldId, params.SalesBillId)
	if err != nil {
		return nil, nil, err
	}
	if bill == nil {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation("sales_bill_id", ReasonBillNotFound,
			"no bill exists with id '"+params.SalesBillId+"'"))
		return nil, vErrs, nil
	}

	intent := params.Intent
	if intent == "" {
		intent = string(models.SalesFiscalIntentIssueOriginal)
	}

	// The replay check comes before the gates, for the same reason it does in RecordPayment: a
	// caller retrying a request it already made must get the first answer back even if the bill has
	// since moved on. Refusing the retry is how a caller ends up asking a third time.
	if params.IdempotencyKey != "" {
		existing, err := findFiscalRequestByKey(ctx, params.IdempotencyKey)
		if err != nil {
			return nil, nil, err
		}
		if existing != nil {
			return replayFiscalResult(existing), nil, nil
		}
	}

	if vErrs := assertInvoiceRequestable(ctx, bill, intent, params); vErrs != nil {
		return nil, vErrs, nil
	}

	originalReference, vErrs, err := resolveOriginalReference(ctx, intent, params)
	if err != nil {
		return nil, nil, err
	}
	if vErrs != nil {
		return nil, vErrs, nil
	}

	idempotencyKey := params.IdempotencyKey
	if idempotencyKey == "" {
		generated, err := model.NewId()
		if err != nil {
			return nil, nil, err
		}
		idempotencyKey = string(*generated)
	}

	requestId, err := writeFiscalRequest(ctx, bill, intent, idempotencyKey, params)
	if err != nil {
		return nil, nil, err
	}

	// The order moves to `requested`, not `issued`. It is the honest state while the provider has
	// not answered, and it is what makes an unanswered request visible instead of silent.
	if err := syncOrderInvoiceStatus(ctx,
		stringOf(bill, models.SalesBillFieldSalesOrderId),
		string(models.SalesOrderInvoiceStatusRequested)); err != nil {
		return nil, nil, err
	}

	result := &RequestInvoiceResult{
		FiscalDocumentRequestId: requestId,
		SalesBillId:             params.SalesBillId,
		Status:                  string(models.SalesFiscalStatusPending),
	}
	if provider == nil {
		// No adapter bound. The row stands as `pending` and an operator or a later adapter picks it
		// up; inventing a success here would report a document that does not exist.
		return result, nil, nil
	}

	issueRequest, err := buildIssueRequest(ctx, bill, requestId, idempotencyKey,
		intent, originalReference, params)
	if err != nil {
		return nil, nil, err
	}

	response, err := provider.Issue(ctx, *issueRequest)
	if err != nil {
		// A transport failure is NOT a Go error to the caller. The provider being unreachable is
		// normal operation, the request is already recorded, and turning it into a 500 would tell an
		// operator the system is broken when the correct action is to retry.
		if recordErr := recordFiscalFailure(ctx, requestId, err.Error()); recordErr != nil {
			return nil, nil, recordErr
		}
		result.Status = string(models.SalesFiscalStatusFailed)
		return result, nil, nil
	}

	if err := recordFiscalOutcome(ctx, bill, requestId, response); err != nil {
		return nil, nil, err
	}
	if response.Issued {
		result.Status = string(models.SalesFiscalStatusIssued)
		result.ProviderReference = response.ProviderReference
	} else {
		result.Status = string(models.SalesFiscalStatusFailed)
	}
	return result, nil, nil
}

// assertInvoiceRequestable applies the three gates that are Sales' own (BR 77).
func assertInvoiceRequestable(
	ctx corectx.Context, bill dmodel.DynamicFields, intent string, params RequestInvoiceParams,
) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()

	if !isKnownFiscalIntent(intent) {
		vErrs.Append(*ft.NewBusinessViolation("intent", ReasonUnknownFiscalIntent,
			"unrecognised fiscal intent '"+intent+"'"))
		return vErrs
	}

	// A cancelled bill settled nothing, so there is no transaction to invoice. An OPEN one is
	// allowed: BR 33 separates the bill from the invoice, and a business may legitimately issue an
	// invoice for goods delivered on credit before the money arrives.
	if stringOf(bill, models.SalesBillFieldStatus) == string(models.SalesBillStatusCancelled) {
		vErrs.Append(*ft.NewBusinessViolation("sales_bill_id", ReasonBillNotEligibleForInvoice,
			"a cancelled bill has no transaction to invoice"))
	}

	if missing := missingBuyerFields(params.Buyer); missing != "" {
		vErrs.Append(*ft.NewBusinessViolation("buyer", ReasonBuyerInfoIncomplete,
			"the buyer fiscal information is missing "+missing))
	}

	// Only ONE original invoice per bill. A second is a tax filing to correct, not a row to delete,
	// so both an issued and an in-flight request block: Sales does not know whether an in-flight one
	// became issued, and assuming it did not is how the duplicate happens.
	if intent == string(models.SalesFiscalIntentIssueOriginal) {
		blocking, err := findBlockingOriginalRequest(ctx, stringOf(bill, models.SalesBillFieldId))
		if err != nil {
			// A read failure here must not be swallowed into a permit. Refusing is the safe side:
			// the caller retries, whereas a wrongly permitted request issues a second document.
			vErrs.Append(*ft.NewBusinessViolation("sales_bill_id", ReasonInvoiceAlreadyRequested,
				"could not confirm whether an invoice was already requested for this bill"))
			return vErrs
		}
		if blocking != nil {
			vErrs.Append(*ft.NewBusinessViolation("sales_bill_id", ReasonInvoiceAlreadyRequested,
				"an invoice has already been requested for this bill and is "+
					stringOf(blocking, models.SalesFiscalRequestFieldStatus)))
		}
	}

	if vErrs.Count() > 0 {
		return vErrs
	}
	return nil
}

// missingBuyerFields names what the buyer information lacks, or empty when it is complete.
//
// Tax code and legal name only. An address and an email make an invoice more useful and neither is
// what makes it valid, and requiring them would refuse invoices a provider would have accepted -
// Sales deciding invoice law by the back door.
func missingBuyerFields(buyer itInvoicing.BuyerInfo) string {
	switch {
	case buyer.TaxCode == "" && buyer.LegalName == "":
		return "a tax code and a legal name"
	case buyer.TaxCode == "":
		return "a tax code"
	case buyer.LegalName == "":
		return "a legal name"
	}
	return ""
}

func isKnownFiscalIntent(intent string) bool {
	switch models.SalesFiscalIntent(intent) {
	case models.SalesFiscalIntentIssueOriginal,
		models.SalesFiscalIntentAdjustForFullReturn,
		models.SalesFiscalIntentAdjustForPartialReturn,
		models.SalesFiscalIntentAdjustPrice:
		return true
	}
	return false
}

// resolveOriginalReference finds the provider reference of the document being adjusted (BR 58).
//
// An adjustment needs the ORIGINAL'S provider reference, not its Sales id: the provider knows the
// document it issued and has never heard of a sales_fiscal_requests row. An original that was never
// issued has no reference, and adjusting it is meaningless rather than merely unsupported.
func resolveOriginalReference(
	ctx corectx.Context, intent string, params RequestInvoiceParams,
) (string, *ft.ClientErrors, error) {
	if intent == string(models.SalesFiscalIntentIssueOriginal) {
		return "", nil, nil
	}

	vErrs := ft.NewClientErrors()
	if params.OriginalFiscalRequestId == "" {
		vErrs.Append(*ft.NewBusinessViolation("original_fiscal_request_id",
			ReasonOriginalRequestRequired,
			"an adjustment must name the fiscal request it adjusts"))
		return "", vErrs, nil
	}

	original, err := loadRecord(ctx, models.SalesFiscalRequestSchemaName,
		models.SalesFiscalRequestFieldId, params.OriginalFiscalRequestId)
	if err != nil {
		return "", nil, err
	}
	if original == nil || !models.NewSalesFiscalRequestFrom(original).IsIssued() {
		vErrs.Append(*ft.NewBusinessViolation("original_fiscal_request_id",
			ReasonOriginalRequestNotIssued,
			"the fiscal request being adjusted has not been issued, so there is no document to adjust"))
		return "", vErrs, nil
	}
	return stringOf(original, models.SalesFiscalRequestFieldProviderRef), nil, nil
}

// findBlockingOriginalRequest answers whether an original invoice already stands for this bill.
func findBlockingOriginalRequest(
	ctx corectx.Context, billId string,
) (dmodel.DynamicFields, error) {
	requests, err := searchBy(ctx,
		models.SalesFiscalRequestSchemaName, models.SalesFiscalRequestFieldSalesBillId, billId)
	if err != nil {
		return nil, err
	}
	for _, request := range requests {
		if stringOf(request, models.SalesFiscalRequestFieldIntent) !=
			string(models.SalesFiscalIntentIssueOriginal) {
			continue
		}
		if models.NewSalesFiscalRequestFrom(request).BlocksNewRequest() {
			return request, nil
		}
	}
	return nil, nil
}

// findFiscalRequestByKey backs the replay path.
func findFiscalRequestByKey(
	ctx corectx.Context, idempotencyKey string,
) (dmodel.DynamicFields, error) {
	found, err := searchBy(ctx, models.SalesFiscalRequestSchemaName,
		models.SalesFiscalRequestFieldIdempotencyKey, idempotencyKey)
	if err != nil {
		return nil, err
	}
	if len(found) == 0 {
		return nil, nil
	}
	return found[0], nil
}

func replayFiscalResult(existing dmodel.DynamicFields) *RequestInvoiceResult {
	return &RequestInvoiceResult{
		FiscalDocumentRequestId: stringOf(existing, models.SalesFiscalRequestFieldId),
		SalesBillId:             stringOf(existing, models.SalesFiscalRequestFieldSalesBillId),
		Status:                  stringOf(existing, models.SalesFiscalRequestFieldStatus),
		ProviderReference:       stringOf(existing, models.SalesFiscalRequestFieldProviderRef),
		AlreadyExisted:          true,
	}
}

// writeFiscalRequest stores the intent before the provider is called. See the package comment.
func writeFiscalRequest(
	ctx corectx.Context,
	bill dmodel.DynamicFields,
	intent string,
	idempotencyKey string,
	params RequestInvoiceParams,
) (string, error) {
	id, err := model.NewId()
	if err != nil {
		return "", err
	}
	requestId := string(*id)

	engine, err := engineFor(models.SalesFiscalRequestSchemaName)
	if err != nil {
		return "", err
	}

	record := dmodel.DynamicFields{
		models.SalesFiscalRequestFieldId:             requestId,
		models.SalesFiscalRequestFieldSalesBillId:    stringOf(bill, models.SalesBillFieldId),
		models.SalesFiscalRequestFieldIntent:         intent,
		models.SalesFiscalRequestFieldStatus:         string(models.SalesFiscalStatusPending),
		models.SalesFiscalRequestFieldIdempotencyKey: idempotencyKey,
		models.SalesFiscalRequestFieldRequestedAt:    model.ModelDateTime(time.Now().UTC()),

		// The buyer information is frozen here (BR 87.7). A later change of company name or address
		// must not rewrite what a historical invoice said.
		models.SalesFiscalRequestFieldBuyerSnapshot: buyerSnapshotOf(params.Buyer),

		basemodel.FieldOrgId: stringOf(bill, basemodel.FieldOrgId),
	}
	if params.OriginalFiscalRequestId != "" {
		record[models.SalesFiscalRequestFieldOriginalId] = params.OriginalFiscalRequestId
	}

	if _, err := engine.ResourceRepository().Insert(ctx, record); err != nil {
		return "", err
	}

	// The event says a document was REQUESTED, never that one was issued - matching the row, which
	// is written as pending for the same reason (BR 77). A consumer learning that an invoice
	// exists must learn it from the provider having confirmed, not from Sales having asked.
	eventType := models.EventFiscalDocumentRequested
	if intent != string(models.SalesFiscalIntentIssueOriginal) {
		eventType = models.EventFiscalAdjustmentRequested
	}
	if _, err := RecordEvent(ctx, RecordEventParams{
		EventType:   eventType,
		AggregateId: stringOf(bill, models.SalesBillFieldSalesOrderId),
		OrgId:       stringOf(bill, basemodel.FieldOrgId),
		Payload: map[string]any{
			"sales_fiscal_request_id": requestId,
			"sales_bill_id":           stringOf(bill, models.SalesBillFieldId),
			"intent":                  intent,
			"total_amount":            decimalOf(bill, models.SalesBillFieldTotalAmount),
			"currency_code":           stringOf(bill, models.SalesBillFieldCurrencyCode),
		},
	}); err != nil {
		return "", err
	}
	return requestId, nil
}

// buyerSnapshotOf freezes the buyer's fiscal identity as a plain map.
//
// A map rather than the struct, because it goes into a jsonmap column and comes back as one. Storing
// the struct would make the read path depend on the struct's field names never changing, which is
// exactly the coupling a snapshot exists to avoid.
func buyerSnapshotOf(buyer itInvoicing.BuyerInfo) map[string]any {
	return map[string]any{
		"tax_code":   buyer.TaxCode,
		"legal_name": buyer.LegalName,
		"address":    buyer.Address,
		"email":      buyer.Email,
	}
}

// buildIssueRequest assembles the FACTS the provider needs (BR 50).
//
// Historical amounts throughout, read from the bill and its allocations rather than recomputed from
// current prices (BR 54, BR 60, BR 94.22). A document stating today's prices would state a sum the
// customer never paid.
func buildIssueRequest(
	ctx corectx.Context,
	bill dmodel.DynamicFields,
	requestId string,
	idempotencyKey string,
	intent string,
	originalReference string,
	params RequestInvoiceParams,
) (*itInvoicing.IssueRequest, error) {
	billId := stringOf(bill, models.SalesBillFieldId)

	allocations, err := searchBy(ctx,
		models.SalesBillLineSchemaName, models.SalesBillLineFieldSalesBillId, billId)
	if err != nil {
		return nil, err
	}

	lines := make([]itInvoicing.IssueLine, 0, len(allocations))
	for _, allocation := range allocations {
		orderLineId := stringOf(allocation, models.SalesBillLineFieldSalesOrderLineId)

		line := itInvoicing.IssueLine{
			SalesOrderLineId: orderLineId,
			Quantity:         decimalOf(allocation, models.SalesBillLineFieldQuantity),
			NetAmount:        decimalOf(allocation, models.SalesBillLineFieldAllocatedNetAmount),
			TaxAmount:        decimalOf(allocation, models.SalesBillLineFieldAllocatedTaxAmount),
		}

		// The description, unit and rate live on the order line, which the allocation points at.
		// Read rather than derived: the line froze them at confirmation, and a provider must state
		// what was sold rather than what the product master says today.
		orderLine, err := loadRecord(ctx,
			models.SalesOrderLineSchemaName, models.SalesOrderLineFieldId, orderLineId)
		if err != nil {
			return nil, err
		}
		if orderLine != nil {
			line.Description = stringOf(orderLine, models.SalesOrderLineFieldProductNameSnapshot)
			line.UomId = stringOf(orderLine, models.SalesOrderLineFieldUomId)
			line.UnitAmount = decimalOf(orderLine, models.SalesOrderLineFieldEffectiveUnitPrice)
			line.TaxRateSnapshot = decimalOf(orderLine, models.SalesOrderLineFieldTaxRateSnapshot)
		}
		lines = append(lines, line)
	}

	request := &itInvoicing.IssueRequest{
		IdempotencyKey:            idempotencyKey,
		Intent:                    itInvoicing.FiscalIntent(intent),
		SalesFiscalRequestId:      requestId,
		SalesBillId:               billId,
		OriginalProviderReference: originalReference,
		Buyer:                     params.Buyer,
		CurrencyCode:              stringOf(bill, models.SalesBillFieldCurrencyCode),
		Subtotal:                  decimalOf(bill, models.SalesBillFieldSubtotal),
		TaxTotal:                  decimalOf(bill, models.SalesBillFieldTaxTotal),
		TotalAmount:               decimalOf(bill, models.SalesBillFieldTotalAmount),
		Lines:                     lines,
		Reason:                    params.Reason,
		OccurredAt:                NowUnix(),
	}

	// The tax snapshot travels verbatim, so the provider can state what was actually charged rather
	// than what today's tax master would charge.
	order, err := loadRecord(ctx,
		models.SalesOrderSchemaName, models.SalesOrderFieldId,
		stringOf(bill, models.SalesBillFieldSalesOrderId))
	if err != nil {
		return nil, err
	}
	if order != nil {
		request.TaxSnapshot = models.NewSalesOrderFrom(order).GetTaxSnapshot()
	}
	return request, nil
}

// recordFiscalOutcome stamps what the provider answered onto the request.
//
// provider_reference and issued_at are written HERE and nowhere else, and only when the provider
// confirmed. That single write point is what makes BR 77 checkable rather than a convention.
func recordFiscalOutcome(
	ctx corectx.Context,
	bill dmodel.DynamicFields,
	requestId string,
	response *itInvoicing.IssueResult,
) error {
	engine, err := engineFor(models.SalesFiscalRequestSchemaName)
	if err != nil {
		return err
	}

	update := dmodel.DynamicFields{models.SalesFiscalRequestFieldId: requestId}
	if response.Issued {
		update[models.SalesFiscalRequestFieldStatus] = string(models.SalesFiscalStatusIssued)
		update[models.SalesFiscalRequestFieldProviderRef] = response.ProviderReference
		update[models.SalesFiscalRequestFieldIssuedAt] = issuedAtOf(response)
	} else {
		update[models.SalesFiscalRequestFieldStatus] = string(models.SalesFiscalStatusFailed)
		update[models.SalesFiscalRequestFieldLastError] = response.FailureReason
	}
	if _, err := engine.ResourceRepository().Update(ctx, update); err != nil {
		return err
	}

	orderStatus := string(models.SalesOrderInvoiceStatusFailed)
	if response.Issued {
		orderStatus = string(models.SalesOrderInvoiceStatusIssued)
	}
	return syncOrderInvoiceStatus(ctx,
		stringOf(bill, models.SalesBillFieldSalesOrderId), orderStatus)
}

// issuedAtOf takes the date the DOCUMENT bears, which is the provider's, not Sales'.
//
// Falls back to now only when the provider said nothing, because a document with no date at all is
// harder to reconcile than one dated a few seconds late.
func issuedAtOf(response *itInvoicing.IssueResult) model.ModelDateTime {
	if response.IssuedAt > 0 {
		return model.ModelDateTime(time.Unix(response.IssuedAt, 0).UTC())
	}
	return model.ModelDateTime(time.Now().UTC())
}

// recordFiscalFailure notes a transport failure and counts the attempt.
//
// attempt_count is incremented from what was read rather than left to the caller, so a retry loop
// cannot forget to count and a bounded policy has something to bound itself by.
func recordFiscalFailure(ctx corectx.Context, requestId, message string) error {
	existing, err := loadRecord(ctx,
		models.SalesFiscalRequestSchemaName, models.SalesFiscalRequestFieldId, requestId)
	if err != nil {
		return err
	}

	engine, err := engineFor(models.SalesFiscalRequestSchemaName)
	if err != nil {
		return err
	}
	_, err = engine.ResourceRepository().Update(ctx, dmodel.DynamicFields{
		models.SalesFiscalRequestFieldId:           requestId,
		models.SalesFiscalRequestFieldStatus:       string(models.SalesFiscalStatusFailed),
		models.SalesFiscalRequestFieldLastError:    truncateError(message),
		models.SalesFiscalRequestFieldAttemptCount: int32Of(existing, models.SalesFiscalRequestFieldAttemptCount) + 1,
	})
	return err
}

// truncateError keeps a provider message inside the column.
//
// A message longer than the column is a write failure, and losing the whole record of why an invoice
// failed because the explanation was verbose is a poor trade for the first thousand characters.
func truncateError(message string) string {
	const maxLength = 1000
	if len(message) <= maxLength {
		return message
	}
	return message[:maxLength]
}

// syncOrderInvoiceStatus reflects the fiscal request onto the order.
//
// The order's invoice_status is a summary of its bills' fiscal requests, held on the order because
// that is where an operator looks. It is written through the repository because the field is
// declared no_update: it must not be editable through a plain PATCH, and this operation is the only
// sanctioned way to move it.
func syncOrderInvoiceStatus(ctx corectx.Context, orderId, status string) error {
	if orderId == "" {
		return nil
	}
	engine, err := engineFor(models.SalesOrderSchemaName)
	if err != nil {
		return err
	}
	_, err = engine.ResourceRepository().Update(ctx, dmodel.DynamicFields{
		models.SalesOrderFieldId:            orderId,
		models.SalesOrderFieldInvoiceStatus: status,
	})
	return err
}
