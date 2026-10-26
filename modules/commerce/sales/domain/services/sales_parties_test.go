package services

import (
	"testing"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// The party rules, which decide who a sale says bought, who it bills, and who owes the money. Being
// wrong here bills the wrong company.

// A role the caller did not mention is left exactly as it was; one sent as an explicit null is
// cleared. Collapsing the two would mean either that a role can never be cleared, or that omitting
// one silently wipes it — and the second would quietly drop a buyer from a sale.
func TestAnAbsentRoleDiffersFromAClearedOne(t *testing.T) {
	absent := PartyAssignment{}
	if absent.Requested {
		t.Error("a role nobody mentioned must not count as requested")
	}

	cleared := PartyAssignment{Requested: true}
	if !cleared.Requested || cleared.PartyId != "" {
		t.Error("an explicit null must read as requested with no party")
	}

	set := PartyAssignment{Requested: true, PartyId: "01PARTY00000000000000000"}
	if !set.Requested || set.PartyId == "" {
		t.Error("a named party must read as requested with that party")
	}
}

// Bill-to freezes once the document is being raised or has been: at `processing` a worker already
// holds the snapshot, and at `issued` a legal document names that party. Nothing else freezes it —
// a draft or ready instruction is still just an intention.
func TestOnlyProcessingAndIssuedFreezeBillTo(t *testing.T) {
	frozen := map[models.SalesBillingInstructionStatus]bool{
		models.SalesBillingInstructionStatusDraft:      false,
		models.SalesBillingInstructionStatusReady:      false,
		models.SalesBillingInstructionStatusProcessing: true,
		models.SalesBillingInstructionStatusIssued:     true,
		models.SalesBillingInstructionStatusFailed:     false,
		models.SalesBillingInstructionStatusCancelled:  false,
	}

	for status, wantFrozen := range frozen {
		instruction := dmodel.DynamicFields{
			models.SalesBillingInstructionFieldStatus: string(status),
		}
		got := freezesBillTo(stringOf(instruction, models.SalesBillingInstructionFieldStatus))
		if got != wantFrozen {
			t.Errorf("status %q freezes bill-to = %v, want %v", status, got, wantFrozen)
		}
	}
}

// A cancelled instruction does not lock the order. It is not being acted on, and treating it as a
// lock would leave a sale permanently unable to be billed to anyone after one withdrawn request.
func TestACancelledInstructionIsNotActive(t *testing.T) {
	cancelled := models.NewSalesBillingInstructionFrom(dmodel.DynamicFields{
		models.SalesBillingInstructionFieldStatus: string(models.SalesBillingInstructionStatusCancelled),
	})
	if cancelled.IsActive() {
		t.Error("a cancelled instruction must not count as the sale's active arrangement")
	}
}

// Assigning one role writes exactly that column. The three are independent — a subsidiary buys, its
// head office is invoiced, a finance company settles — so a cascade would silently overwrite a
// deliberate choice somebody made about who pays.
func TestAssigningOneRoleTouchesNoOther(t *testing.T) {
	params := AssignPartiesParams{
		SalesOrderId: "01ORDER00000000000000000",
		BillTo:       PartyAssignment{Requested: true, PartyId: "01PARTY00000000000000000"},
	}

	changes := requestedPartyChanges(params)

	if _, ok := changes[models.SalesOrderFieldBillToPartyId]; !ok {
		t.Error("the named role must be written")
	}
	if _, ok := changes[models.SalesOrderFieldSoldToPartyId]; ok {
		t.Error("sold-to was not named and must not be written")
	}
	if _, ok := changes[models.SalesOrderFieldPayerPartyId]; ok {
		t.Error("payer was not named and must not be written")
	}
}

// Clearing writes a null rather than skipping the column, which is what makes an explicit null
// actually remove the party instead of being read as "unchanged".
func TestClearingARoleWritesNull(t *testing.T) {
	params := AssignPartiesParams{
		SalesOrderId: "01ORDER00000000000000000",
		Payer:        PartyAssignment{Requested: true},
	}

	changes := requestedPartyChanges(params)

	value, ok := changes[models.SalesOrderFieldPayerPartyId]
	if !ok {
		t.Fatal("an explicitly cleared role must appear in the changes")
	}
	if value != nil {
		t.Errorf("clearing must write nil, got %v", value)
	}
}

// A request naming no role at all changes nothing. Writing an empty update would bump the etag and
// look like an edit in the audit trail while saying nothing happened.
func TestNamingNoRoleProducesNoChanges(t *testing.T) {
	changes := requestedPartyChanges(AssignPartiesParams{SalesOrderId: "01ORDER00000000000000000"})
	if len(changes) != 0 {
		t.Errorf("a request naming no role must produce no changes, got %d", len(changes))
	}
}
