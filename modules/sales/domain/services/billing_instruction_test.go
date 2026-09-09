package services

import (
	"testing"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// The billing instruction's state machine, tested exhaustively because it is what stands between a
// half-filled form and a legal document issued in someone's name.

func instructionWith(fields dmodel.DynamicFields) dmodel.DynamicFields {
	return fields
}

func instructionAt(status models.SalesBillingInstructionStatus) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.SalesBillingInstructionFieldStatus: string(status),
	}
}

// Every status the schema declares must appear in the transition table: canTransition returns false
// for an unknown `from`, so an instruction reaching a missing status could never move again.
func TestEveryBillingStatusAppearsInTheTable(t *testing.T) {
	statuses := []models.SalesBillingInstructionStatus{
		models.SalesBillingInstructionStatusDraft,
		models.SalesBillingInstructionStatusReady,
		models.SalesBillingInstructionStatusProcessing,
		models.SalesBillingInstructionStatusIssued,
		models.SalesBillingInstructionStatusFailed,
		models.SalesBillingInstructionStatusCancelled,
	}
	for _, status := range statuses {
		if _, ok := billingInstructionTransitions[string(status)]; !ok {
			t.Errorf("%q is missing from the transition table, so an instruction "+
				"reaching it could never move again", status)
		}
	}
}

// A claimed instruction is NEVER swept back to ready.
//
// This is the transition whose absence prevents duplicate legal documents: a worker holding the
// instruction may already have created one, and re-releasing it for another worker would issue a
// second invoice for one sale. Recovery means asking the provider, not guessing.
func TestProcessingNeverReturnsToReady(t *testing.T) {
	forbidden := []models.SalesBillingInstructionStatus{
		models.SalesBillingInstructionStatusReady,
		models.SalesBillingInstructionStatusDraft,
		models.SalesBillingInstructionStatusCancelled,
	}
	for _, to := range forbidden {
		if canTransition(billingInstructionTransitions,
			string(models.SalesBillingInstructionStatusProcessing), string(to)) {
			t.Errorf("processing must not be able to become %q: the document may already exist", to)
		}
	}

	// It must reach a verdict, though, or a claimed instruction would be stuck forever.
	for _, to := range []models.SalesBillingInstructionStatus{
		models.SalesBillingInstructionStatusIssued,
		models.SalesBillingInstructionStatusFailed,
	} {
		if !canTransition(billingInstructionTransitions,
			string(models.SalesBillingInstructionStatusProcessing), string(to)) {
			t.Errorf("processing must be able to become %q", to)
		}
	}
}

// An issued document is frozen. Correcting one is the provider's own regulated workflow, and a path
// out of `issued` here would make it reachable from a till.
func TestIssuedAndCancelledAreTerminal(t *testing.T) {
	for _, from := range []models.SalesBillingInstructionStatus{
		models.SalesBillingInstructionStatusIssued,
		models.SalesBillingInstructionStatusCancelled,
	} {
		if len(billingInstructionTransitions[string(from)]) != 0 {
			t.Errorf("%q must be terminal", from)
		}
	}
}

// Only `ready` is issuable. An instruction still being filled in is not the buyer's consent to be
// billed, and issuing from a draft would bill someone on a half-typed tax code.
func TestOnlyReadyIsIssuable(t *testing.T) {
	cases := map[models.SalesBillingInstructionStatus]bool{
		models.SalesBillingInstructionStatusDraft:      false,
		models.SalesBillingInstructionStatusReady:      true,
		models.SalesBillingInstructionStatusProcessing: false,
		models.SalesBillingInstructionStatusIssued:     false,
		models.SalesBillingInstructionStatusFailed:     false,
		models.SalesBillingInstructionStatusCancelled:  false,
	}
	for status, want := range cases {
		got := models.NewSalesBillingInstructionFrom(instructionAt(status)).IsIssuable()
		if got != want {
			t.Errorf("%q issuable = %v, want %v", status, got, want)
		}
	}
}

// The snapshot is editable only while no document depends on it.
func TestEditableOnlyBeforeADocumentDependsOnIt(t *testing.T) {
	cases := map[models.SalesBillingInstructionStatus]bool{
		models.SalesBillingInstructionStatusDraft:      true,
		models.SalesBillingInstructionStatusReady:      true,
		models.SalesBillingInstructionStatusFailed:     true,
		models.SalesBillingInstructionStatusProcessing: false,
		models.SalesBillingInstructionStatusIssued:     false,
		models.SalesBillingInstructionStatusCancelled:  false,
	}
	for status, want := range cases {
		got := models.NewSalesBillingInstructionFrom(instructionAt(status)).IsEditable()
		if got != want {
			t.Errorf("%q editable = %v, want %v", status, got, want)
		}
	}
}

// Only a cancelled instruction stops being the sale's one arrangement. An issued one very much still
// is, which is what stops a second instruction being opened for a sale already invoiced.
func TestOnlyCancelledStopsBeingActive(t *testing.T) {
	for status, want := range map[models.SalesBillingInstructionStatus]bool{
		models.SalesBillingInstructionStatusDraft:      true,
		models.SalesBillingInstructionStatusReady:      true,
		models.SalesBillingInstructionStatusProcessing: true,
		models.SalesBillingInstructionStatusIssued:     true,
		models.SalesBillingInstructionStatusFailed:     true,
		models.SalesBillingInstructionStatusCancelled:  false,
	} {
		got := models.NewSalesBillingInstructionFrom(instructionAt(status)).IsActive()
		if got != want {
			t.Errorf("%q active = %v, want %v", status, got, want)
		}
	}
}

// Releasing an instruction whose buyer has no tax code would have the job claim it and the provider
// refuse it hours later. Caught early, it is a message someone can act on at the counter.
//
// Judged on the LIVE party, not on the instruction's own columns: those stay empty until issuance,
// so checking them would refuse every instruction ever created.
func TestAnIncompleteBuyerCannotBeInvoiced(t *testing.T) {
	cases := []struct {
		name     string
		identity itExt.PartyFiscalIdentity
		refused  bool
	}{
		{"nothing at all", itExt.PartyFiscalIdentity{}, true},
		{"tax code only", itExt.PartyFiscalIdentity{TaxCode: "0312345678"}, true},
		{"legal name only", itExt.PartyFiscalIdentity{LegalName: "CÔNG TY TNHH ABC"}, true},
		{"both", itExt.PartyFiscalIdentity{
			TaxCode:   "0312345678",
			LegalName: "CÔNG TY TNHH ABC",
		}, false},
		{"both, no address or email", itExt.PartyFiscalIdentity{
			TaxCode:   "0312345678",
			LegalName: "CÔNG TY TNHH ABC",
		}, false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			vErrs := assertFiscalIdentityComplete(testCase.identity)
			if testCase.refused && (vErrs == nil || vErrs.Count() == 0) {
				t.Error("an incomplete buyer must be refused before release")
			}
			if !testCase.refused && vErrs != nil {
				t.Errorf("a complete buyer must pass, got %v", *vErrs)
			}
		})
	}
}

// An instruction with no snapshot has nothing to reuse, so issuance must go and read the party.
// Presence is judged on the two fields a document cannot go out without: a row carrying only an
// address was half-filled, not captured.
func TestAnUncapturedInstructionHasNoSnapshot(t *testing.T) {
	if existingSnapshotOf(dmodel.DynamicFields{}) != nil {
		t.Error("an instruction that was never captured must report no snapshot")
	}

	addressOnly := dmodel.DynamicFields{
		models.SalesBillingInstructionFieldBillingAddress: "456 Lê Lợi",
	}
	if existingSnapshotOf(addressOnly) != nil {
		t.Error("an address alone is not a captured snapshot")
	}

	captured := dmodel.DynamicFields{
		models.SalesBillingInstructionFieldTaxId:     "0312345678",
		models.SalesBillingInstructionFieldLegalName: "CÔNG TY TNHH ABC",
	}
	if existingSnapshotOf(captured) == nil {
		t.Error("an instruction holding a tax code and legal name has been captured")
	}
}

// The refresh flag is off unless somebody ticked it.
//
// This is the whole safety of the retry path: an instruction that came back from a failed issuance
// carries details a person reviewed, and re-reading the party by default would let an edit made
// meanwhile change what is billed without anyone asking for it.
func TestRefreshIsOffUnlessAskedFor(t *testing.T) {
	unset := models.NewSalesBillingInstructionFrom(dmodel.DynamicFields{})
	if unset.WantsLatestPartyDetails() {
		t.Error("an instruction that never named the flag must not refresh")
	}

	off := models.NewSalesBillingInstructionFrom(dmodel.DynamicFields{
		models.SalesBillingInstructionFieldFetchLatestParty: false,
	})
	if off.WantsLatestPartyDetails() {
		t.Error("an unticked box must not refresh")
	}

	on := models.NewSalesBillingInstructionFrom(dmodel.DynamicFields{
		models.SalesBillingInstructionFieldFetchLatestParty: true,
	})
	if !on.WantsLatestPartyDetails() {
		t.Error("a ticked box must refresh")
	}
}

// A stored snapshot is re-checked before it is billed, not trusted. It may have been left
// incomplete by a correction, and the point of failing is that no document names half a buyer.
func TestAStoredSnapshotIsStillChecked(t *testing.T) {
	if assertSnapshotIsInvoiceable(BillingSnapshot{TaxId: "0312345678"}) == nil {
		t.Error("a snapshot with no legal name must be refused")
	}
	if assertSnapshotIsInvoiceable(BillingSnapshot{
		TaxId:     "0312345678",
		LegalName: "CÔNG TY TNHH ABC",
	}) != nil {
		t.Error("a complete snapshot must be billable")
	}
}

// A partial correction leaves untouched fields alone. Blanking a legal name because this request only
// carried a new address would quietly destroy a confirmed detail.
func TestPartialCorrectionDoesNotBlankOtherFields(t *testing.T) {
	changes := dmodel.DynamicFields{}
	applySnapshotFields(changes, BillingSnapshot{BillingAddress: "456 Lê Lợi"})

	if _, present := changes[models.SalesBillingInstructionFieldLegalName]; present {
		t.Error("a correction that did not mention the legal name must not write it")
	}
	if changes[models.SalesBillingInstructionFieldBillingAddress] != "456 Lê Lợi" {
		t.Error("the field that was supplied must be written")
	}
}

// An attempt whose reply never came back blocks retry. Retrying an indeterminate attempt is how one
// sale acquires two legal invoices.
func TestIndeterminateAttemptIsRecognised(t *testing.T) {
	for status, want := range map[models.SalesBillingAttemptStatus]bool{
		models.SalesBillingAttemptStatusUnknown:    true,
		models.SalesBillingAttemptStatusFailed:     false,
		models.SalesBillingAttemptStatusSucceeded:  false,
		models.SalesBillingAttemptStatusProcessing: false,
	} {
		attempt := models.NewSalesBillingIssuanceAttemptFrom(dmodel.DynamicFields{
			models.SalesBillingIssuanceAttemptFieldStatus: string(status),
		})
		if got := attempt.IsIndeterminate(); got != want {
			t.Errorf("%q indeterminate = %v, want %v", status, got, want)
		}
	}
}
