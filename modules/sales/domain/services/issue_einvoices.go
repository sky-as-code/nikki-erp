package services

import (
	"strconv"
	"time"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
	itInvoicing "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external/invoicing"
)

// Issuing the electronic invoices that are due.
//
// THIS IS THE ONLY THING THAT ISSUES A DOCUMENT. Not the payment webhook, not the fulfilment event,
// not the refund: those tell Sales what happened commercially, and turning any of them into an
// issuance trigger would put a legal document on the end of a callback that can fire twice, arrive
// late, or arrive for a sale that is about to be reversed.
//
// A sale becomes due when four things are all true — the buyer asked to be invoiced and confirmed
// their details, the money is in, the goods have gone, and enough time has passed that a
// same-visit correction would already have happened. Anything less and issuing is premature: a VAT
// invoice cannot simply be deleted afterwards.

// IssueEinvoicesResult reports what one pass did.
type IssueEinvoicesResult struct {
	Examined int
	Issued   int
	Failed   int

	// Indeterminate counts attempts whose reply never came back. Neither issued nor failed, and the
	// ones a human must resolve: a document may exist.
	Indeterminate int
}

// IssueDueEinvoices issues for every instruction that is due.
func IssueDueEinvoices(
	ctx corectx.Context,
	provider itInvoicing.InvoicingExtService,
	parties itExt.PartyExtService,
	policy SalesPolicy,
	now time.Time,
	limit int,
) (*IssueEinvoicesResult, error) {
	result := &IssueEinvoicesResult{}
	if provider == nil {
		// Nothing can issue. Instructions stay `ready`, which is exactly the state they should be in
		// while there is no provider — they are picked up whenever one is bound.
		return result, nil
	}

	cutoff := now.Add(-time.Duration(policy.InvoiceIssueDelayMinutes) * time.Minute)
	due, err := instructionsDueForIssuance(ctx, cutoff, limit)
	if err != nil {
		return nil, err
	}

	for _, instruction := range due {
		result.Examined++

		outcome, err := issueOneInstruction(ctx, provider, parties, instruction, now)
		if err != nil {
			// One instruction failing must not end the pass: the rest are independent, and this one
			// comes round on the next run.
			continue
		}
		switch outcome {
		case issuanceIssued:
			result.Issued++
		case issuanceFailed:
			result.Failed++
		case issuanceIndeterminate:
			result.Indeterminate++
		}
	}
	return result, nil
}

type issuanceOutcome int

const (
	issuanceSkipped issuanceOutcome = iota
	issuanceIssued
	issuanceFailed
	issuanceIndeterminate
)

// issueOneInstruction claims one instruction and raises its document.
func issueOneInstruction(
	ctx corectx.Context,
	provider itInvoicing.InvoicingExtService,
	parties itExt.PartyExtService,
	instruction dmodel.DynamicFields,
	now time.Time,
) (issuanceOutcome, error) {
	instructionId := stringOf(instruction, models.SalesBillingInstructionFieldId)

	// THE CLAIM IS WHAT PREVENTS TWO DOCUMENTS FOR ONE SALE. It moves ready→processing conditionally
	// on the row still being `ready`, carrying the etag read a moment ago, so of two workers reaching
	// the same instruction exactly one wins and the loser moves on.
	claimed, err := claimInstructionForIssuance(ctx, instruction, now)
	if err != nil || !claimed {
		return issuanceSkipped, err
	}

	// THE SNAPSHOT IS TAKEN HERE AND NOWHERE ELSE. Until this moment the instruction held only a
	// reference to the party, so the screen and the document read the same live record and cannot
	// disagree. From here they must: the document about to be raised says who the buyer was when it
	// was raised, and no later edit to their contact record may rewrite it.
	//
	// Captured after the claim, so exactly one worker takes it, and before the provider call, so the
	// row records what was sent even if the reply is lost.
	snapshot, vErrs, err := captureBillingSnapshot(ctx, instruction, parties, now)
	if err != nil {
		return issuanceSkipped, err
	}
	if vErrs != nil {
		// The buyer's record cannot carry a tax document. Recorded as a failure and the instruction
		// released, rather than issuing one that names nobody: a VAT invoice with an empty tax code
		// is a filing to correct, and correcting it is far more work than fixing the party first.
		return issuanceFailed, releaseUnissuable(ctx, instruction, vErrs, now)
	}

	attemptNo, err := nextAttemptNumber(ctx, instructionId)
	if err != nil {
		return issuanceSkipped, err
	}

	// Deterministic, and derived from the attempt rather than the clock: it is the key the provider
	// deduplicates on, so a retry of THIS attempt must present the same one while a genuinely new
	// attempt must not.
	providerRequestId := "bi:" + instructionId + ":" + strconv.Itoa(int(attemptNo))

	attemptId, err := openIssuanceAttempt(ctx, instruction, attemptNo, providerRequestId, now)
	if err != nil {
		return issuanceSkipped, err
	}

	request, err := buildInstructionIssueRequest(ctx, instruction, snapshot, providerRequestId)
	if err != nil {
		return issuanceSkipped, err
	}

	response, err := provider.Issue(ctx, *request)
	if err != nil {
		// A TRANSPORT FAILURE IS NOT A REFUSAL. The request may have reached the provider and the
		// reply been lost, so nobody can say whether a document exists. The instruction is left in
		// `processing` and the attempt marked indeterminate: reconciliation resolves it against the
		// provider, and a retry before then is how one sale acquires two invoices.
		if err := markAttemptIndeterminate(ctx, attemptId, err.Error(), now); err != nil {
			return issuanceSkipped, err
		}
		return issuanceIndeterminate, nil
	}

	if response != nil && response.Issued {
		return issuanceIssued, completeIssuance(ctx, instruction, attemptId, response, now)
	}
	return issuanceFailed, failIssuance(ctx, instruction, attemptId, response, now)
}

// captureBillingSnapshot reads the buyer from the live party and freezes it onto the instruction.
//
// This is the moment the instruction stops pointing at a record and starts stating one. Before it,
// a screen showing the instruction reads through to the party, so nothing can be stale; after it,
// the columns are what the document said and no edit to the party may change them.
//
// Written to the row BEFORE the provider is called, so that a reply which never arrives still leaves
// behind exactly what was sent — which is what reconciliation compares against.
func captureBillingSnapshot(
	ctx corectx.Context,
	instruction dmodel.DynamicFields,
	parties itExt.PartyExtService,
	now time.Time,
) (BillingSnapshot, *ft.ClientErrors, error) {
	// An instruction that came back from a failed issuance already holds a snapshot somebody
	// reviewed. It is reused as-is unless the operator ticked the refresh box when releasing it
	// again: re-reading by default would let an edit made to the party meanwhile change what is
	// billed, without anybody having asked for that.
	existing := existingSnapshotOf(instruction)
	if existing != nil && !models.NewSalesBillingInstructionFrom(instruction).WantsLatestPartyDetails() {
		if vErrs := assertSnapshotIsInvoiceable(*existing); vErrs != nil {
			return BillingSnapshot{}, vErrs, nil
		}
		return *existing, nil, nil
	}

	partyId := stringOf(instruction, models.SalesBillingInstructionFieldBillToPartyId)
	if partyId == "" {
		return BillingSnapshot{}, refusal(models.SalesBillingInstructionFieldBillToPartyId,
			ReasonBillingInstructionIncomplete,
			"the sale names nobody to invoice"), nil
	}
	if parties == nil {
		return BillingSnapshot{}, refusal(models.SalesBillingInstructionFieldBillToPartyId,
			ReasonBillingInstructionIncomplete,
			"the buyer's details cannot be read"), nil
	}

	identity, err := parties.GetFiscalIdentity(ctx, itExt.GetPartyFiscalIdentityQuery{
		PartyId: partyId,
		OrgId:   stringOf(instruction, basemodel.FieldOrgId),
	})
	if err != nil {
		return BillingSnapshot{}, nil, err
	}
	if identity == nil {
		return BillingSnapshot{}, refusal(models.SalesBillingInstructionFieldBillToPartyId,
			ReasonBillingInstructionIncomplete,
			"the party this sale is to be invoiced to no longer exists"), nil
	}
	if vErrs := assertFiscalIdentityComplete(*identity); vErrs != nil {
		return BillingSnapshot{}, vErrs, nil
	}

	snapshot := BillingSnapshot{
		TaxId:          identity.TaxCode,
		LegalName:      identity.LegalName,
		BillingAddress: identity.Address,
		BillingEmail:   identity.Email,
	}

	changes := dmodel.DynamicFields{
		models.SalesBillingInstructionFieldSnapshotAt: model.ModelDateTime(now),
	}
	applySnapshotFields(changes, snapshot)
	if err := writeChanges(ctx,
		models.SalesBillingInstructionSchemaName, instruction, changes); err != nil {
		return BillingSnapshot{}, nil, err
	}
	return snapshot, nil, nil
}

// existingSnapshotOf returns the snapshot already frozen onto an instruction, or nil when none is.
//
// Presence is judged on the two fields a document cannot go out without. An instruction carrying
// only an address would be one somebody half-filled, not one that was captured.
func existingSnapshotOf(instruction dmodel.DynamicFields) *BillingSnapshot {
	snapshot := BillingSnapshot{
		TaxId:          stringOf(instruction, models.SalesBillingInstructionFieldTaxId),
		LegalName:      stringOf(instruction, models.SalesBillingInstructionFieldLegalName),
		BillingAddress: stringOf(instruction, models.SalesBillingInstructionFieldBillingAddress),
		BillingEmail:   stringOf(instruction, models.SalesBillingInstructionFieldBillingEmail),
	}
	if snapshot.TaxId == "" && snapshot.LegalName == "" {
		return nil
	}
	return &snapshot
}

// assertSnapshotIsInvoiceable refuses a stored snapshot that could not carry a document.
//
// Re-checked rather than trusted: it may have been written by a correction that left it incomplete,
// and the point of failing here is that no document goes out naming half a buyer.
func assertSnapshotIsInvoiceable(snapshot BillingSnapshot) *ft.ClientErrors {
	return assertFiscalIdentityComplete(itExt.PartyFiscalIdentity{
		TaxCode:   snapshot.TaxId,
		LegalName: snapshot.LegalName,
	})
}

// releaseUnissuable records why a claimed instruction cannot be invoiced and lets it go.
//
// Back to `draft` rather than left in `processing`: nothing was sent, so no document can exist, and
// leaving it claimed would strand it there until somebody reconciled a provider call that never
// happened. Draft is also where it can be corrected, which is what has to happen next.
func releaseUnissuable(
	ctx corectx.Context, instruction dmodel.DynamicFields, vErrs *ft.ClientErrors, now time.Time,
) error {
	return writeChanges(ctx, models.SalesBillingInstructionSchemaName, instruction,
		dmodel.DynamicFields{
			models.SalesBillingInstructionFieldStatus: string(
				models.SalesBillingInstructionStatusDraft),
			models.SalesBillingInstructionFieldLockedAt:         nil,
			models.SalesBillingInstructionFieldLastErrorCode:    ReasonBillingInstructionIncomplete,
			models.SalesBillingInstructionFieldLastErrorMessage: firstViolationMessageOf(vErrs),
		})
}

// firstViolationMessageOf names the reason a person can act on, or a fallback when there is none.
func firstViolationMessageOf(vErrs *ft.ClientErrors) string {
	if vErrs != nil {
		for _, item := range *vErrs {
			if item.Message != "" {
				return item.Message
			}
		}
	}
	return "the buyer's details are incomplete"
}

// claimInstructionForIssuance moves ready→processing, and reports whether this caller won.
func claimInstructionForIssuance(
	ctx corectx.Context, instruction dmodel.DynamicFields, now time.Time,
) (bool, error) {
	if !models.NewSalesBillingInstructionFrom(instruction).IsIssuable() {
		return false, nil
	}

	err := writeChanges(ctx, models.SalesBillingInstructionSchemaName, instruction,
		dmodel.DynamicFields{
			models.SalesBillingInstructionFieldStatus: string(
				models.SalesBillingInstructionStatusProcessing),
			models.SalesBillingInstructionFieldLockedAt: model.ModelDateTime(now),
		})
	if err != nil {
		// The etag moved under us: another worker claimed it first. Not an error — that is the guard
		// doing its job.
		return false, nil
	}
	return true, nil
}

// completeIssuance records the document on both the attempt and the instruction.
func completeIssuance(
	ctx corectx.Context,
	instruction dmodel.DynamicFields,
	attemptId string,
	response *itInvoicing.IssueResult,
	now time.Time,
) error {
	issuedAt := now
	if response.IssuedAt > 0 {
		// The provider's date, not ours: a document bears the date its issuer gave it.
		issuedAt = time.Unix(response.IssuedAt, 0).UTC()
	}

	if err := updateAttempt(ctx, attemptId, dmodel.DynamicFields{
		models.SalesBillingIssuanceAttemptFieldStatus: string(
			models.SalesBillingAttemptStatusSucceeded),
		models.SalesBillingIssuanceAttemptFieldCompletedAt: model.ModelDateTime(now),
		models.SalesBillingIssuanceAttemptFieldProviderRef: response.ProviderReference,
	}); err != nil {
		return err
	}

	fresh, err := loadBillingInstruction(ctx,
		stringOf(instruction, models.SalesBillingInstructionFieldId))
	if err != nil || fresh == nil {
		return err
	}
	if err := writeChanges(ctx, models.SalesBillingInstructionSchemaName, fresh,
		dmodel.DynamicFields{
			models.SalesBillingInstructionFieldStatus: string(
				models.SalesBillingInstructionStatusIssued),
			models.SalesBillingInstructionFieldIssuedAt:    model.ModelDateTime(issuedAt),
			models.SalesBillingInstructionFieldEinvoiceRef: response.ProviderReference,
		}); err != nil {
		return err
	}

	return syncOrderInvoiceStatus(ctx,
		stringOf(instruction, models.SalesBillingInstructionFieldSalesOrderId),
		string(models.SalesOrderInvoiceStatusIssued))
}

// failIssuance records a definite refusal, which is correctable and retryable.
func failIssuance(
	ctx corectx.Context,
	instruction dmodel.DynamicFields,
	attemptId string,
	response *itInvoicing.IssueResult,
	now time.Time,
) error {
	reason := "the provider did not issue the document"
	if response != nil && response.FailureReason != "" {
		reason = response.FailureReason
	}

	if err := updateAttempt(ctx, attemptId, dmodel.DynamicFields{
		models.SalesBillingIssuanceAttemptFieldStatus: string(
			models.SalesBillingAttemptStatusFailed),
		models.SalesBillingIssuanceAttemptFieldCompletedAt:  model.ModelDateTime(now),
		models.SalesBillingIssuanceAttemptFieldErrorMessage: truncateError(reason),
	}); err != nil {
		return err
	}

	fresh, err := loadBillingInstruction(ctx,
		stringOf(instruction, models.SalesBillingInstructionFieldId))
	if err != nil || fresh == nil {
		return err
	}
	if err := writeChanges(ctx, models.SalesBillingInstructionSchemaName, fresh,
		dmodel.DynamicFields{
			models.SalesBillingInstructionFieldStatus: string(
				models.SalesBillingInstructionStatusFailed),
			models.SalesBillingInstructionFieldLastErrorMessage: truncateError(reason),
		}); err != nil {
		return err
	}

	return syncOrderInvoiceStatus(ctx,
		stringOf(instruction, models.SalesBillingInstructionFieldSalesOrderId),
		string(models.SalesOrderInvoiceStatusFailed))
}

// markAttemptIndeterminate records that nobody knows what happened.
//
// The INSTRUCTION IS LEFT IN `processing` on purpose. Moving it to failed would invite a retry, and
// the document may already exist; leaving it claimed is what forces a human or a reconciliation pass
// to establish the truth before anything else is attempted.
func markAttemptIndeterminate(
	ctx corectx.Context, attemptId, detail string, now time.Time,
) error {
	return updateAttempt(ctx, attemptId, dmodel.DynamicFields{
		models.SalesBillingIssuanceAttemptFieldStatus: string(
			models.SalesBillingAttemptStatusUnknown),
		models.SalesBillingIssuanceAttemptFieldCompletedAt:  model.ModelDateTime(now),
		models.SalesBillingIssuanceAttemptFieldErrorMessage: truncateError(detail),
	})
}

// buildInstructionIssueRequest assembles what the provider is sent for one billing instruction.
//
// Named apart from request_invoice.go's buildIssueRequest because the two start from different
// things: that one from a fiscal request, which already carries the buyer, and this one from an
// instruction, whose whole point is that it carries the buyer's own confirmed snapshot.
//
// The recipient comes ENTIRELY from the instruction's snapshot. The bill-to party is never read
// here, even though the instruction names one: the buyer confirmed these details, and navigating to
// the partner record would let a change made since then alter what the document says.
func buildInstructionIssueRequest(
	ctx corectx.Context,
	instruction dmodel.DynamicFields,
	snapshot BillingSnapshot,
	providerRequestId string,
) (*itInvoicing.IssueRequest, error) {
	orderId := stringOf(instruction, models.SalesBillingInstructionFieldSalesOrderId)

	bill, err := findPrimaryBillOfOrder(ctx, orderId)
	if err != nil {
		return nil, err
	}
	if bill == nil {
		return nil, nil
	}

	lines, err := buildIssueLinesForOrder(ctx, orderId)
	if err != nil {
		return nil, err
	}

	return &itInvoicing.IssueRequest{
		IdempotencyKey: providerRequestId,
		Intent:         itInvoicing.IntentIssueOriginal,
		SalesBillId:    stringOf(bill, models.SalesBillFieldId),
		OrgId:          stringOf(instruction, basemodel.FieldOrgId),
		// From the snapshot just captured, not from the instruction's columns: those are still empty
		// on the copy loaded before the capture, and reading them would send an empty buyer.
		Buyer: itInvoicing.BuyerInfo{
			TaxCode:   snapshot.TaxId,
			LegalName: snapshot.LegalName,
			Address:   snapshot.BillingAddress,
			Email:     snapshot.BillingEmail,
		},
		CurrencyCode: stringOf(bill, models.SalesBillFieldCurrencyCode),
		Subtotal:     decimalOf(bill, models.SalesBillFieldSubtotal),
		TaxTotal:     decimalOf(bill, models.SalesBillFieldTaxTotal),
		TotalAmount:  decimalOf(bill, models.SalesBillFieldTotalAmount),
		Lines:        lines,
		OccurredAt:   NowUnix(),
	}, nil
}

// buildIssueLinesForOrder states the sale's lines in their historical amounts.
func buildIssueLinesForOrder(
	ctx corectx.Context, orderId string,
) ([]itInvoicing.IssueLine, error) {
	orderLines, err := searchBy(ctx, models.SalesOrderLineSchemaName,
		models.SalesOrderLineFieldSalesOrderId, orderId)
	if err != nil {
		return nil, err
	}

	lines := make([]itInvoicing.IssueLine, 0, len(orderLines))
	for _, line := range orderLines {
		lines = append(lines, itInvoicing.IssueLine{
			SalesOrderLineId: stringOf(line, models.SalesOrderLineFieldId),
			Description:      stringOf(line, models.SalesOrderLineFieldProductNameSnapshot),
			Quantity:         decimalOf(line, models.SalesOrderLineFieldOrderedQuantity),
			UomId:            stringOf(line, models.SalesOrderLineFieldUomId),
			UnitAmount:       decimalOf(line, models.SalesOrderLineFieldEffectiveUnitPrice),
			NetAmount:        decimalOf(line, models.SalesOrderLineFieldNetAmount),
			TaxAmount:        decimalOf(line, models.SalesOrderLineFieldTaxAmount),
			TaxRateSnapshot:  decimalOf(line, models.SalesOrderLineFieldTaxRateSnapshot),
		})
	}
	return lines, nil
}

// instructionsDueForIssuance finds what may be issued now.
//
// The status filter is pushed into the query; the rest is checked per row. Eligibility spans the
// order and its bills, which one SearchGraph cannot express, and the alternative — a hand-written
// join — would put SQL in a layer that has none.
func instructionsDueForIssuance(
	ctx corectx.Context, cutoff time.Time, limit int,
) ([]dmodel.DynamicFields, error) {
	engine, err := engineFor(models.SalesBillingInstructionSchemaName)
	if err != nil {
		return nil, err
	}

	size := limit
	if size <= 0 {
		size = model.MODEL_RULE_PAGE_MAX_SIZE
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(
			models.SalesBillingInstructionFieldStatus, dmodel.Equals,
			string(models.SalesBillingInstructionStatusReady)),
	)

	found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		// Over-read, because most candidates are filtered out below and a page sized to the limit
		// would return far fewer than asked for.
		Size: size * 4,
	})
	if err != nil {
		return nil, err
	}
	if found == nil || !found.HasData {
		return nil, nil
	}

	due := make([]dmodel.DynamicFields, 0, size)
	for _, instruction := range found.Data.Items {
		eligible, err := isOrderEligibleForIssuance(ctx,
			stringOf(instruction, models.SalesBillingInstructionFieldSalesOrderId), cutoff)
		if err != nil {
			continue
		}
		if !eligible {
			continue
		}
		due = append(due, instruction)
		if len(due) >= size {
			break
		}
	}
	return due, nil
}

// isOrderEligibleForIssuance applies the commercial conditions.
//
// PAID IS NOT ENOUGH. A sale whose goods have not gone can still be cancelled at the counter, and a
// refund in flight may change what is owed — issuing in either case produces a document that states
// something that stopped being true. The delay on top covers the corrections that happen within
// minutes and would otherwise each become a credit note.
func isOrderEligibleForIssuance(
	ctx corectx.Context, orderId string, cutoff time.Time,
) (bool, error) {
	if orderId == "" {
		return false, nil
	}

	order, err := loadRecord(ctx, models.SalesOrderSchemaName, models.SalesOrderFieldId, orderId)
	if err != nil || order == nil {
		return false, err
	}

	if stringOf(order, models.SalesOrderFieldPaymentStatus) !=
		string(models.SalesOrderPaymentStatusPaid) {
		return false, nil
	}
	if stringOf(order, models.SalesOrderFieldFulfillmentStatus) !=
		string(models.SalesOrderFulfillmentStatusFulfilled) {
		return false, nil
	}

	// A return still in flight may change what is owed. One already completed does not block
	// issuance — the amounts are settled — but an open one means the sale is not final.
	inFlight, err := hasInFlightReturn(ctx, orderId)
	if err != nil || inFlight {
		return false, err
	}

	// Settled long enough ago. The anchor is the bill's settled_at, which is written exactly once
	// when the money lands; the order carries no reliable paid-at of its own.
	settledAt, err := latestSettlementOfOrder(ctx, orderId)
	if err != nil || settledAt == nil {
		return false, err
	}
	return settledAt.Before(cutoff), nil
}

// hasInFlightReturn reports whether a return is open against the order.
func hasInFlightReturn(ctx corectx.Context, orderId string) (bool, error) {
	returns, err := searchBy(ctx,
		models.SalesReturnSchemaName, models.SalesReturnFieldSalesOrderId, orderId)
	if err != nil {
		return false, err
	}
	for _, salesReturn := range returns {
		status := stringOf(salesReturn, models.SalesReturnFieldStatus)
		if status != string(models.SalesReturnStatusCompleted) &&
			status != string(models.SalesReturnStatusCancelled) {
			return true, nil
		}
	}
	return false, nil
}

// latestSettlementOfOrder answers when the sale was fully paid, or nil if any bill is still open.
func latestSettlementOfOrder(ctx corectx.Context, orderId string) (*time.Time, error) {
	bills, err := searchBy(ctx,
		models.SalesBillSchemaName, models.SalesBillFieldSalesOrderId, orderId)
	if err != nil {
		return nil, err
	}
	if len(bills) == 0 {
		return nil, nil
	}

	var latest *time.Time
	for _, bill := range bills {
		if stringOf(bill, models.SalesBillFieldStatus) ==
			string(models.SalesBillStatusCancelled) {
			continue
		}
		settledAt := dateTimeOf(bill, models.SalesBillFieldSettledAt)
		if settledAt == nil {
			// One unsettled bill means the sale is not fully paid, whatever the others say.
			return nil, nil
		}
		at := settledAt.GoTime()
		if latest == nil || at.After(*latest) {
			latest = &at
		}
	}
	return latest, nil
}

// findPrimaryBillOfOrder returns the bill a document is raised against.
//
// The first settled one, which for the overwhelming majority of sales is the only one. A sale split
// across several bills wants a document per bill and that is a larger design; this refuses to guess
// at it rather than silently invoicing one part.
func findPrimaryBillOfOrder(
	ctx corectx.Context, orderId string,
) (dmodel.DynamicFields, error) {
	bills, err := searchBy(ctx,
		models.SalesBillSchemaName, models.SalesBillFieldSalesOrderId, orderId)
	if err != nil {
		return nil, err
	}
	for _, bill := range bills {
		if stringOf(bill, models.SalesBillFieldStatus) ==
			string(models.SalesBillStatusSettled) {
			return bill, nil
		}
	}
	return nil, nil
}

// nextAttemptNumber counts what has already been tried, so attempts read in order.
func nextAttemptNumber(ctx corectx.Context, instructionId string) (int32, error) {
	attempts, err := searchBy(ctx, models.SalesBillingIssuanceAttemptSchemaName,
		models.SalesBillingIssuanceAttemptFieldInstructionId, instructionId)
	if err != nil {
		return 0, err
	}

	highest := int32(0)
	for _, attempt := range attempts {
		if number := int32Of(attempt,
			models.SalesBillingIssuanceAttemptFieldAttemptNo); number > highest {
			highest = number
		}
	}
	return highest + 1, nil
}

// openIssuanceAttempt records that a try has begun, before the provider is called.
//
// Written FIRST on purpose: an attempt that is never recorded because the process died mid-call is
// exactly the case nobody can reconstruct afterwards.
func openIssuanceAttempt(
	ctx corectx.Context,
	instruction dmodel.DynamicFields,
	attemptNo int32,
	providerRequestId string,
	now time.Time,
) (string, error) {
	engine, err := engineFor(models.SalesBillingIssuanceAttemptSchemaName)
	if err != nil {
		return "", err
	}
	id, err := model.NewId()
	if err != nil {
		return "", err
	}

	if _, err := engine.ResourceRepository().Insert(ctx, dmodel.DynamicFields{
		models.SalesBillingIssuanceAttemptFieldId:            string(*id),
		models.SalesBillingIssuanceAttemptFieldInstructionId: stringOf(instruction, models.SalesBillingInstructionFieldId),
		models.SalesBillingIssuanceAttemptFieldAttemptNo:     attemptNo,
		models.SalesBillingIssuanceAttemptFieldStatus: string(
			models.SalesBillingAttemptStatusProcessing),
		models.SalesBillingIssuanceAttemptFieldStartedAt:     model.ModelDateTime(now),
		models.SalesBillingIssuanceAttemptFieldProviderReqId: providerRequestId,
		basemodel.FieldOrgId:                                 stringOf(instruction, basemodel.FieldOrgId),
	}); err != nil {
		return "", err
	}
	return string(*id), nil
}

// updateAttempt closes out one attempt.
func updateAttempt(
	ctx corectx.Context, attemptId string, changes dmodel.DynamicFields,
) error {
	attempt, err := loadRecord(ctx, models.SalesBillingIssuanceAttemptSchemaName,
		models.SalesBillingIssuanceAttemptFieldId, attemptId)
	if err != nil || attempt == nil {
		return err
	}
	return writeChanges(ctx, models.SalesBillingIssuanceAttemptSchemaName, attempt, changes)
}
