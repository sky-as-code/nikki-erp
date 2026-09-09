package services

import (
	"time"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// The billing instruction: who a sale is to be invoiced to, and under which legal identity.
//
// EVERY RULE HERE PROTECTS ONE THING — that a legal document is never issued from information the
// buyer did not confirm, and never changes after it has been. That is why marking ready is its own
// operation rather than a field update, why a claimed instruction is frozen, and why an issued one
// cannot be edited at all.
//
// THE SNAPSHOT IS TAKEN AT ISSUANCE, NOT BEFORE. Until then the instruction holds only
// bill_to_party_id and the snapshot columns are empty: a screen showing the instruction reads
// through to the party, so what staff see is always current and there is no copy in between to go
// stale. At issuance the party is read one last time, frozen onto these columns, and the instruction
// locks — so a company that renames itself afterwards does not alter an invoice already issued.
//
// This is why `mark_ready` validates the LIVE party rather than these columns, and why the only
// writer of them is captureBillingSnapshot in issue_einvoices.go.

// The refusal reasons the billing instruction operations produce.
const (
	ReasonBillingInstructionNotFound      = "sales_billing_instruction.not_found"
	ReasonBillingInstructionExists        = "sales_billing_instruction.already_exists"
	ReasonBillingInstructionNotEditable   = "sales_billing_instruction.not_editable"
	ReasonBillingInstructionNotReady      = "sales_billing_instruction.not_ready"
	ReasonBillingInstructionIncomplete    = "sales_billing_instruction.incomplete"
	ReasonBillingInstructionOrderNotFound = "sales_billing_instruction.order_not_found"
)

// billingInstructionTransitions is the whole state machine, as a table.
//
// `processing` leads only to a verdict: an instruction a worker claimed is never swept back to
// `ready`, because the document may already exist and re-issuing is the one mistake that cannot be
// undone. Recovering it means asking the provider, not guessing.
//
// `issued` and `cancelled` lead nowhere. A document that exists is corrected through the provider's
// own regulated workflow, and a withdrawn request is replaced by a new instruction rather than
// reopened.
var billingInstructionTransitions = map[string][]string{
	string(models.SalesBillingInstructionStatusDraft): {
		string(models.SalesBillingInstructionStatusReady),
		string(models.SalesBillingInstructionStatusCancelled),
	},
	string(models.SalesBillingInstructionStatusReady): {
		string(models.SalesBillingInstructionStatusDraft),
		string(models.SalesBillingInstructionStatusProcessing),
		string(models.SalesBillingInstructionStatusCancelled),
	},
	string(models.SalesBillingInstructionStatusProcessing): {
		string(models.SalesBillingInstructionStatusIssued),
		string(models.SalesBillingInstructionStatusFailed),
	},
	string(models.SalesBillingInstructionStatusFailed): {
		string(models.SalesBillingInstructionStatusDraft),
		string(models.SalesBillingInstructionStatusReady),
		string(models.SalesBillingInstructionStatusCancelled),
	},
	string(models.SalesBillingInstructionStatusIssued):    {},
	string(models.SalesBillingInstructionStatusCancelled): {},
}

// BillingSnapshot is the buyer's fiscal identity as they confirmed it.
type BillingSnapshot struct {
	TaxId          string
	LegalName      string
	BillingAddress string
	BillingEmail   string
}

// CreateBillingInstructionParams opens a billing arrangement for one sale.
type CreateBillingInstructionParams struct {
	SalesOrderId  string
	BillToPartyId string
	Snapshot      BillingSnapshot

	// Source records who supplied this. Empty defaults to back office, which is the conservative
	// reading: it claims staff entered it rather than that the buyer self-served.
	Source string
}

// CreateBillingInstruction records that a sale is to be invoiced, and to whom.
func CreateBillingInstruction(
	ctx corectx.Context, params CreateBillingInstructionParams,
) (string, *ft.ClientErrors, error) {
	order, err := loadRecord(ctx,
		models.SalesOrderSchemaName, models.SalesOrderFieldId, params.SalesOrderId)
	if err != nil {
		return "", nil, err
	}
	if order == nil {
		return "", refusal("sales_order_id", ReasonBillingInstructionOrderNotFound,
			"no sales order exists with id '"+params.SalesOrderId+"'"), nil
	}

	// Checked here as well as by the partial unique index. The index is what makes it correct under
	// concurrency; this is what turns the common case into a message a till can show.
	existing, err := findActiveBillingInstruction(ctx, params.SalesOrderId)
	if err != nil {
		return "", nil, err
	}
	if existing != nil {
		return "", refusal("sales_order_id", ReasonBillingInstructionExists,
			"this sale already has a billing instruction; correct it rather than adding another"), nil
	}

	source := params.Source
	if source == "" {
		source = string(models.SalesBillingInstructionSourceBackOffice)
	}

	engine, err := engineFor(models.SalesBillingInstructionSchemaName)
	if err != nil {
		return "", nil, err
	}
	id, err := model.NewId()
	if err != nil {
		return "", nil, err
	}

	fields := dmodel.DynamicFields{
		models.SalesBillingInstructionFieldId:           string(*id),
		models.SalesBillingInstructionFieldSalesOrderId: params.SalesOrderId,
		models.SalesBillingInstructionFieldStatus: string(
			models.SalesBillingInstructionStatusDraft),
		models.SalesBillingInstructionFieldSource: source,
		basemodel.FieldOrgId:                      stringOf(order, basemodel.FieldOrgId),
	}
	if params.BillToPartyId != "" {
		fields[models.SalesBillingInstructionFieldBillToPartyId] = params.BillToPartyId
	}
	applySnapshotFields(fields, params.Snapshot)

	if _, err := engine.ResourceRepository().Insert(ctx, fields); err != nil {
		return "", nil, err
	}
	return string(*id), nil, nil
}

// UpdateBillingInstructionSnapshot corrects the fiscal details before a document is issued.
//
// It NEVER writes back to the business partner. The snapshot is allowed to diverge — a buyer may
// give an address for this invoice that is not the company's registered one — and propagating an
// edit outward would let a till quietly rewrite master data.
func UpdateBillingInstructionSnapshot(
	ctx corectx.Context, instructionId string, snapshot BillingSnapshot,
) (*ft.ClientErrors, error) {
	instruction, err := loadBillingInstruction(ctx, instructionId)
	if err != nil {
		return nil, err
	}
	if instruction == nil {
		return refusal("id", ReasonBillingInstructionNotFound,
			"no billing instruction exists with id '"+instructionId+"'"), nil
	}

	if !models.NewSalesBillingInstructionFrom(instruction).IsEditable() {
		return refusal("status", ReasonBillingInstructionNotEditable,
			"a billing instruction that is '"+
				stringOf(instruction, models.SalesBillingInstructionFieldStatus)+
				"' can no longer be corrected"), nil
	}

	changes := dmodel.DynamicFields{}
	applySnapshotFields(changes, snapshot)

	return nil, writeChanges(ctx, models.SalesBillingInstructionSchemaName, instruction, changes)
}

// MarkBillingInstructionReady is the buyer's consent to be billed.
//
// It is the gate the issuance job reads, so the completeness check lives here: an instruction that
// reaches `ready` missing a tax code would be claimed by the job and then refused by the provider,
// turning a correctable mistake into a failed document hours later.
func MarkBillingInstructionReady(
	ctx corectx.Context,
	instructionId string,
	fetchLatestPartyDetails bool,
	parties itExt.PartyExtService,
) (*ft.ClientErrors, error) {
	instruction, err := loadBillingInstruction(ctx, instructionId)
	if err != nil {
		return nil, err
	}
	if instruction == nil {
		return refusal("id", ReasonBillingInstructionNotFound,
			"no billing instruction exists with id '"+instructionId+"'"), nil
	}

	if vErrs := assertBillingTransition(instruction,
		models.SalesBillingInstructionStatusReady); vErrs != nil {
		return vErrs, nil
	}

	// Validated against the LIVE party, not against this row: the snapshot columns are empty until
	// the document is raised, so checking them here would refuse every instruction ever created.
	//
	// The check is repeated at issuance because the party can change in between — but doing it here
	// too is what tells the person marking it ready, while they are looking at it, rather than
	// letting a scheduled job fail two hours later where nobody is watching.
	vErrs, err := assertPartyCanBeInvoiced(ctx, instruction, parties)
	if err != nil || vErrs != nil {
		return vErrs, err
	}

	// snapshot_at is deliberately NOT stamped here. It records when the buyer's details were
	// captured, and nothing is captured until issuance; setting it now would date a snapshot that
	// does not exist yet.
	//
	// The refresh flag is carried on the row rather than acted on now, because the capture it
	// governs happens later, in the issuance job, with nobody watching.
	return nil, writeChanges(ctx, models.SalesBillingInstructionSchemaName, instruction,
		dmodel.DynamicFields{
			models.SalesBillingInstructionFieldStatus: string(
				models.SalesBillingInstructionStatusReady),
			models.SalesBillingInstructionFieldSubmittedAt:      model.ModelDateTime(time.Now().UTC()),
			models.SalesBillingInstructionFieldFetchLatestParty: fetchLatestPartyDetails,
		})
}

// assertPartyCanBeInvoiced refuses an instruction whose party cannot carry a tax document.
//
// A tax code and a legal name are what a VAT invoice must name. An address and an email are not
// required: a provider accepts a document without them, and demanding them here would refuse
// invoices that would have been issued perfectly well.
func assertPartyCanBeInvoiced(
	ctx corectx.Context, instruction dmodel.DynamicFields, parties itExt.PartyExtService,
) (*ft.ClientErrors, error) {
	partyId := stringOf(instruction, models.SalesBillingInstructionFieldBillToPartyId)
	if partyId == "" {
		return refusal(models.SalesBillingInstructionFieldBillToPartyId,
			ReasonBillingInstructionIncomplete,
			"the sale must name who is to be invoiced before it can be billed"), nil
	}
	if parties == nil {
		// No port bound. Refusing is the conservative reading: permitting would let an instruction
		// reach the issuance job, which would then raise a document naming nobody.
		return refusal(models.SalesBillingInstructionFieldBillToPartyId,
			ReasonBillingInstructionIncomplete,
			"the buyer's details cannot be read, so this sale cannot be billed"), nil
	}

	identity, err := parties.GetFiscalIdentity(ctx, itExt.GetPartyFiscalIdentityQuery{
		PartyId: partyId,
		OrgId:   stringOf(instruction, models.SalesBillingInstructionFieldOrgId),
	})
	if err != nil {
		return nil, err
	}
	if identity == nil {
		return refusal(models.SalesBillingInstructionFieldBillToPartyId,
			ReasonBillingInstructionIncomplete,
			"the party this sale is to be invoiced to no longer exists"), nil
	}

	return assertFiscalIdentityComplete(*identity), nil
}

// assertFiscalIdentityComplete names what the buyer's record lacks, or nil when it is invoiceable.
func assertFiscalIdentityComplete(identity itExt.PartyFiscalIdentity) *ft.ClientErrors {
	missing := []string{}
	if identity.TaxCode == "" {
		missing = append(missing, "tax code")
	}
	if identity.LegalName == "" {
		missing = append(missing, "legal name")
	}
	if len(missing) == 0 {
		return nil
	}

	detail := missing[0]
	if len(missing) > 1 {
		detail = missing[0] + " and " + missing[1]
	}
	return refusal("tax_id", ReasonBillingInstructionIncomplete,
		"the buyer's "+detail+" is needed before this sale can be invoiced")
}

// RevertBillingInstructionToDraft takes a released instruction back for correction.
//
// submitted_at is deliberately left alone: it records that consent was given once, which remains
// true, rather than that it is currently in force. The status is what says the latter.
func RevertBillingInstructionToDraft(
	ctx corectx.Context, instructionId string,
) (*ft.ClientErrors, error) {
	return moveBillingInstruction(ctx, instructionId,
		models.SalesBillingInstructionStatusDraft, nil)
}

// CancelBillingInstruction withdraws a buyer's request to be invoiced.
//
// Only before a document exists. An issued instruction is refused rather than cancelled, because
// cancelling a legal document is the provider's own regulated workflow and must not be reachable
// from a till.
func CancelBillingInstruction(
	ctx corectx.Context, instructionId string,
) (*ft.ClientErrors, error) {
	return moveBillingInstruction(ctx, instructionId,
		models.SalesBillingInstructionStatusCancelled, nil)
}

// moveBillingInstruction applies one transition, checked against the table.
func moveBillingInstruction(
	ctx corectx.Context,
	instructionId string,
	to models.SalesBillingInstructionStatus,
	extra dmodel.DynamicFields,
) (*ft.ClientErrors, error) {
	instruction, err := loadBillingInstruction(ctx, instructionId)
	if err != nil {
		return nil, err
	}
	if instruction == nil {
		return refusal("id", ReasonBillingInstructionNotFound,
			"no billing instruction exists with id '"+instructionId+"'"), nil
	}
	if vErrs := assertBillingTransition(instruction, to); vErrs != nil {
		return vErrs, nil
	}

	changes := dmodel.DynamicFields{
		models.SalesBillingInstructionFieldStatus: string(to),
	}
	for key, value := range extra {
		changes[key] = value
	}
	return nil, writeChanges(ctx, models.SalesBillingInstructionSchemaName, instruction, changes)
}

// assertBillingTransition refuses a move the state machine does not allow.
func assertBillingTransition(
	instruction dmodel.DynamicFields, to models.SalesBillingInstructionStatus,
) *ft.ClientErrors {
	from := stringOf(instruction, models.SalesBillingInstructionFieldStatus)
	if canTransition(billingInstructionTransitions, from, string(to)) {
		return nil
	}
	return refusal("status", ReasonBillingInstructionNotReady,
		"a billing instruction that is '"+from+"' cannot become '"+string(to)+"'")
}

// applySnapshotFields writes the four snapshot columns, leaving an empty one untouched so a partial
// correction does not blank the fields it did not mention.
func applySnapshotFields(target dmodel.DynamicFields, snapshot BillingSnapshot) {
	if snapshot.TaxId != "" {
		target[models.SalesBillingInstructionFieldTaxId] = snapshot.TaxId
	}
	if snapshot.LegalName != "" {
		target[models.SalesBillingInstructionFieldLegalName] = snapshot.LegalName
	}
	if snapshot.BillingAddress != "" {
		target[models.SalesBillingInstructionFieldBillingAddress] = snapshot.BillingAddress
	}
	if snapshot.BillingEmail != "" {
		target[models.SalesBillingInstructionFieldBillingEmail] = snapshot.BillingEmail
	}
}

func loadBillingInstruction(
	ctx corectx.Context, instructionId string,
) (dmodel.DynamicFields, error) {
	return loadRecord(ctx, models.SalesBillingInstructionSchemaName,
		models.SalesBillingInstructionFieldId, instructionId)
}

// findActiveBillingInstruction returns the sale's one live arrangement, if it has one. Cancelled
// instructions are skipped: they are kept for audit and do not stand in the way of a new one.
func findActiveBillingInstruction(
	ctx corectx.Context, salesOrderId string,
) (dmodel.DynamicFields, error) {
	found, err := searchBy(ctx, models.SalesBillingInstructionSchemaName,
		models.SalesBillingInstructionFieldSalesOrderId, salesOrderId)
	if err != nil {
		return nil, err
	}
	for _, instruction := range found {
		if models.NewSalesBillingInstructionFrom(instruction).IsActive() {
			return instruction, nil
		}
	}
	return nil, nil
}
