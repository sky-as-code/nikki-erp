package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// The billing instruction: who a sale is to be invoiced to, under which legal identity.
//
// IT IS NOT AN INVOICE, and it is not a request for one either. It records that a buyer asked to be
// billed and agreed the details, which is a different fact from a document existing — and a
// different fact again from the sale having been paid. Issuance reads it; nothing else may.
//
// It supersedes sales_fiscal_request for original issuance. The fiscal request remains for the
// adjustments a return produces, which amend a document that already exists and therefore need the
// reference of the one being amended rather than a buyer's consent.
//
// A sale has AT MOST ONE ACTIVE instruction. Cancelled ones are kept, because a buyer who supplied a
// tax code, withdrew it and supplied another leaves a trail worth having.

const (
	SalesBillingInstructionSchemaName = "sales_billing_instruction"

	SalesBillingInstructionFieldId               = basemodel.FieldId
	SalesBillingInstructionFieldOrgId            = basemodel.FieldOrgId
	SalesBillingInstructionFieldSalesOrderId     = "sales_order_id"
	SalesBillingInstructionFieldBillToPartyId    = "bill_to_party_id"
	SalesBillingInstructionFieldTaxId            = "tax_id"
	SalesBillingInstructionFieldLegalName        = "legal_name"
	SalesBillingInstructionFieldBillingAddress   = "billing_address"
	SalesBillingInstructionFieldBillingEmail     = "billing_email"
	SalesBillingInstructionFieldStatus           = "status"
	SalesBillingInstructionFieldSource           = "source"
	SalesBillingInstructionFieldFetchLatestParty = "fetch_latest_party_details"
	SalesBillingInstructionFieldSnapshotAt       = "snapshot_at"
	SalesBillingInstructionFieldLockedAt         = "locked_at"
	SalesBillingInstructionFieldSubmittedAt      = "submitted_at"
	SalesBillingInstructionFieldIssuedAt         = "issued_at"
	SalesBillingInstructionFieldEinvoiceRef      = "einvoice_reference"
	SalesBillingInstructionFieldLastErrorCode    = "last_error_code"
	SalesBillingInstructionFieldLastErrorMessage = "last_error_message"
)

//go:embed sales_billing_instruction.json
var salesBillingInstructionSchemaJson string

func SalesBillingInstructionSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesBillingInstructionSchemaJson)
}

// SalesBillingInstruction is one sale's billing arrangement.
type SalesBillingInstruction struct {
	basemodel.DynamicModelBase
}

func NewSalesBillingInstruction() *SalesBillingInstruction {
	return &SalesBillingInstruction{basemodel.NewDynamicModel(dmodel.DynamicFields{})}
}

func NewSalesBillingInstructionFrom(src dmodel.DynamicFields) *SalesBillingInstruction {
	return &SalesBillingInstruction{basemodel.NewDynamicModel(src)}
}

func (this SalesBillingInstruction) GetId() *model.Id {
	return this.GetFieldData().GetModelId(SalesBillingInstructionFieldId)
}

func (this SalesBillingInstruction) GetSalesOrderId() *model.Id {
	return this.GetFieldData().GetModelId(SalesBillingInstructionFieldSalesOrderId)
}

func (this SalesBillingInstruction) GetBillToPartyId() *model.Id {
	return this.GetFieldData().GetModelId(SalesBillingInstructionFieldBillToPartyId)
}

func (this SalesBillingInstruction) GetTaxId() *string {
	return this.GetFieldData().GetString(SalesBillingInstructionFieldTaxId)
}

func (this SalesBillingInstruction) GetLegalName() *string {
	return this.GetFieldData().GetString(SalesBillingInstructionFieldLegalName)
}

func (this SalesBillingInstruction) GetBillingAddress() *string {
	return this.GetFieldData().GetString(SalesBillingInstructionFieldBillingAddress)
}

func (this SalesBillingInstruction) GetBillingEmail() *string {
	return this.GetFieldData().GetString(SalesBillingInstructionFieldBillingEmail)
}

func (this SalesBillingInstruction) GetStatus() *string {
	return this.GetFieldData().GetString(SalesBillingInstructionFieldStatus)
}

func (this *SalesBillingInstruction) SetStatus(status *string) {
	this.GetFieldData().SetString(SalesBillingInstructionFieldStatus, status)
}

// WantsLatestPartyDetails reports whether issuance should re-read the buyer rather than reuse the
// snapshot already on this instruction. Absent reads as false: an instruction that came back from a
// failed issuance keeps what somebody reviewed unless they asked for it to be refreshed.
func (this SalesBillingInstruction) WantsLatestPartyDetails() bool {
	value := this.GetFieldData().GetBool(SalesBillingInstructionFieldFetchLatestParty)
	return value != nil && *value
}

func (this SalesBillingInstruction) GetEinvoiceReference() *string {
	return this.GetFieldData().GetString(SalesBillingInstructionFieldEinvoiceRef)
}

// IsEditable reports whether the snapshot may still be changed.
//
// A claimed or issued instruction is frozen: the first because a worker is building a document from
// it right now, the second because the document exists and editing what it was built from would
// leave the record describing something the paper does not say.
func (this SalesBillingInstruction) IsEditable() bool {
	status := this.GetStatus()
	if status == nil {
		return false
	}
	switch *status {
	case string(SalesBillingInstructionStatusDraft),
		string(SalesBillingInstructionStatusReady),
		string(SalesBillingInstructionStatusFailed):
		return true
	}
	return false
}

// IsActive reports whether this instruction still counts as the sale's one arrangement. Only a
// cancelled instruction stops counting — an issued one very much still does.
func (this SalesBillingInstruction) IsActive() bool {
	status := this.GetStatus()
	return status == nil || *status != string(SalesBillingInstructionStatusCancelled)
}

// IsIssuable reports whether the issuance job may claim this instruction. Deliberately narrow: only
// an instruction whose buyer has confirmed the details is consent to bill anyone.
func (this SalesBillingInstruction) IsIssuable() bool {
	status := this.GetStatus()
	return status != nil && *status == string(SalesBillingInstructionStatusReady)
}
