package services

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// Who is party to a sale.
//
// A sale has three parties and they are INDEPENDENT: whoever buys, whoever the invoice is made out
// to, and whoever pays. A business splits them routinely — a subsidiary buys, its head office is
// invoiced, a finance company settles — so assigning one never touches another. Every one is
// nullable, because the default sale is anonymous and stays valid with none of them set.
//
// The rules here are CR 06-sales-party §7. The one that matters most is §7.2: once a billing
// instruction is processing or issued, bill-to is frozen, because the document either is being
// raised against that party or already names them.

// The refusal reasons assigning a party can produce.
const (
	ReasonPartyOrderNotFound  = "sales_order.not_found"
	ReasonBillToPartyFrozen   = "sales_order.bill_to_party_frozen"
	ReasonPartyAssignmentNone = "sales_order.no_party_change_requested"
)

// PartyAssignment is one role's requested change.
//
// Absent and null are DIFFERENT and the distinction is the whole point of this type: a caller
// sending nothing must leave a role alone, while one sending an explicit null must clear it. A bare
// string cannot say which, so presence is carried separately (CR §24.1).
type PartyAssignment struct {
	// Requested marks that the caller named this role at all. False means "not mentioned", and the
	// role is left exactly as it was.
	Requested bool

	// PartyId is the party to assign. Empty with Requested true means an explicit null: clear it.
	PartyId string
}

// AssignPartiesParams is a request to set some or all of a sale's parties.
type AssignPartiesParams struct {
	SalesOrderId string

	SoldTo PartyAssignment
	BillTo PartyAssignment
	Payer  PartyAssignment
}

// AssignParties sets the parties named on a sale, leaving unnamed roles untouched.
//
// ALL OR NOTHING. Every requested role is validated before any is written, so a request that names
// three roles and is refused on one changes nothing. The alternative — writing what passed — leaves
// the order in a state the caller did not ask for and, worse, is easy to read as success.
func AssignParties(
	ctx corectx.Context, params AssignPartiesParams, parties itExt.PartyExtService,
) (*ft.ClientErrors, error) {
	order, err := loadRecord(ctx,
		models.SalesOrderSchemaName, models.SalesOrderFieldId, params.SalesOrderId)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return refusal("sales_order_id", ReasonPartyOrderNotFound,
			"no sales order exists with id '"+params.SalesOrderId+"'"), nil
	}

	requested := requestedPartyRoles(params)

	changes := requestedPartyChanges(params)
	if len(changes) == 0 {
		return refusal("parties", ReasonPartyAssignmentNone,
			"no party role was named, so there is nothing to assign"), nil
	}

	vErrs := ft.NewClientErrors()

	// Bill-to is frozen once the document is being raised or has been. Checked before the party
	// itself, because a frozen role refuses whatever party was offered.
	if params.BillTo.Requested {
		frozen, err := billToIsFrozen(ctx, params.SalesOrderId)
		if err != nil {
			return nil, err
		}
		if frozen != "" {
			vErrs.Append(*ft.NewBusinessViolation(
				models.SalesOrderFieldBillToPartyId, ReasonBillToPartyFrozen,
				"the billing instruction for this order is "+frozen+
					", so the bill-to party can no longer be changed"))
		}
	}

	// Every party being SET must be assignable. A role being cleared names no party, so there is
	// nothing to validate — clearing is always allowed.
	orgId := stringOf(order, models.SalesOrderFieldOrgId)
	for field, assignment := range requested {
		if !assignment.Requested || assignment.PartyId == "" {
			continue
		}
		partyErrs, err := parties.AssertAssignable(ctx, itExt.AssertPartyAssignableQuery{
			PartyId: assignment.PartyId,
			OrgId:   orgId,
			Field:   field,
		})
		if err != nil {
			return nil, err
		}
		vErrs.ConcatPtr(partyErrs)
	}

	if vErrs.Count() > 0 {
		return vErrs, nil
	}

	// Only the named roles are written. The snapshot on any billing instruction is deliberately NOT
	// refreshed here: an invoice must say what the buyer confirmed, so propagating a new bill-to
	// into it is an explicit decision the caller makes separately (CR §7.2, §10.1).
	return nil, writeChanges(ctx, models.SalesOrderSchemaName, order, changes)
}

// requestedPartyRoles pairs each role's column with what the caller asked for it.
func requestedPartyRoles(params AssignPartiesParams) map[string]PartyAssignment {
	return map[string]PartyAssignment{
		models.SalesOrderFieldSoldToPartyId: params.SoldTo,
		models.SalesOrderFieldBillToPartyId: params.BillTo,
		models.SalesOrderFieldPayerPartyId:  params.Payer,
	}
}

// requestedPartyChanges turns the request into the columns to write.
//
// A role nobody named is absent from the result entirely, which is what leaves it alone; one that
// was named with no party is present as nil, which is what clears it.
func requestedPartyChanges(params AssignPartiesParams) dmodel.DynamicFields {
	changes := dmodel.DynamicFields{}
	for field, assignment := range requestedPartyRoles(params) {
		if !assignment.Requested {
			continue
		}
		changes[field] = nil
		if assignment.PartyId != "" {
			changes[field] = assignment.PartyId
		}
	}
	return changes
}

// billToIsFrozen names the billing-instruction status that forbids changing bill-to, or empty when
// nothing does.
//
// Only the ACTIVE instruction counts. A cancelled one is not being acted on and does not lock the
// order, which is what lets a withdrawn request be replaced by a new one for a different party.
func billToIsFrozen(ctx corectx.Context, salesOrderId string) (string, error) {
	instruction, err := findActiveBillingInstruction(ctx, salesOrderId)
	if err != nil {
		return "", err
	}
	if instruction == nil {
		return "", nil
	}

	status := stringOf(instruction, models.SalesBillingInstructionFieldStatus)
	if freezesBillTo(status) {
		return status, nil
	}
	return "", nil
}

// freezesBillTo reports whether an instruction in this status locks the bill-to party.
//
// `processing` locks because a worker already holds the snapshot and may be mid-issue; `issued`
// because a legal document names that party and cannot be quietly re-pointed at another. Every
// other status is still an intention, and re-aiming an intention is exactly what this permits.
func freezesBillTo(status string) bool {
	switch status {
	case string(models.SalesBillingInstructionStatusProcessing),
		string(models.SalesBillingInstructionStatusIssued):
		return true
	}
	return false
}
