package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// The fiscal document request: what Sales asked an eInvoice provider for, and what came back. A
// table rather than columns on sales_bills because one bill may need an original invoice and then
// several adjustments as goods come back. Nothing here names a document type, serial or template:
// Sales states what commercially happened and the provider decides what document that requires.

const (
	SalesFiscalRequestSchemaName = "sales_fiscal_request"

	SalesFiscalRequestFieldId             = basemodel.FieldId
	SalesFiscalRequestFieldOrgId          = basemodel.FieldOrgId
	SalesFiscalRequestFieldSalesBillId    = "sales_bill_id"
	SalesFiscalRequestFieldIntent         = "intent"
	SalesFiscalRequestFieldStatus         = "status"
	SalesFiscalRequestFieldIdempotencyKey = "idempotency_key"
	SalesFiscalRequestFieldProviderRef    = "provider_reference"
	SalesFiscalRequestFieldAttemptCount   = "attempt_count"
	SalesFiscalRequestFieldLastError      = "last_error"
	SalesFiscalRequestFieldBuyerSnapshot  = "buyer_snapshot"
	SalesFiscalRequestFieldOriginalId     = "original_fiscal_request_id"
	SalesFiscalRequestFieldRequestedAt    = "requested_at"
	SalesFiscalRequestFieldIssuedAt       = "issued_at"

	SalesFiscalRequestEdgeSalesBill = "sales_bill"
)

//go:embed sales_fiscal_request.json
var salesFiscalRequestSchemaJson string

func SalesFiscalRequestSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesFiscalRequestSchemaJson)
}

// SalesFiscalRequest is one thing Sales asked the eInvoice provider for.
type SalesFiscalRequest struct {
	basemodel.DynamicModelBase
}

func NewSalesFiscalRequest() *SalesFiscalRequest {
	return &SalesFiscalRequest{basemodel.NewDynamicModel(dmodel.DynamicFields{})}
}

func NewSalesFiscalRequestFrom(src dmodel.DynamicFields) *SalesFiscalRequest {
	return &SalesFiscalRequest{basemodel.NewDynamicModel(src)}
}

func (this SalesFiscalRequest) GetId() *model.Id {
	return this.GetFieldData().GetModelId(SalesFiscalRequestFieldId)
}

func (this SalesFiscalRequest) GetSalesBillId() *model.Id {
	return this.GetFieldData().GetModelId(SalesFiscalRequestFieldSalesBillId)
}

func (this SalesFiscalRequest) GetIntent() *string {
	return this.GetFieldData().GetString(SalesFiscalRequestFieldIntent)
}

func (this SalesFiscalRequest) GetStatus() *string {
	return this.GetFieldData().GetString(SalesFiscalRequestFieldStatus)
}

func (this *SalesFiscalRequest) SetStatus(status *string) {
	this.GetFieldData().SetString(SalesFiscalRequestFieldStatus, status)
}

func (this SalesFiscalRequest) GetIdempotencyKey() *string {
	return this.GetFieldData().GetString(SalesFiscalRequestFieldIdempotencyKey)
}

func (this SalesFiscalRequest) GetProviderReference() *string {
	return this.GetFieldData().GetString(SalesFiscalRequestFieldProviderRef)
}

func (this *SalesFiscalRequest) SetProviderReference(reference *string) {
	this.GetFieldData().SetString(SalesFiscalRequestFieldProviderRef, reference)
}

func (this SalesFiscalRequest) GetLastError() *string {
	return this.GetFieldData().GetString(SalesFiscalRequestFieldLastError)
}

func (this SalesFiscalRequest) GetOriginalFiscalRequestId() *model.Id {
	return this.GetFieldData().GetModelId(SalesFiscalRequestFieldOriginalId)
}

func (this SalesFiscalRequest) GetIssuedAt() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(SalesFiscalRequestFieldIssuedAt)
}

// GetBuyerSnapshot returns the buyer's fiscal identity as supplied at issuance. Typed as `any`
// because domain/models must not import another module, and a stored jsonmap comes back as whatever
// the JSON decoder chose.
func (this SalesFiscalRequest) GetBuyerSnapshot() any {
	return this.GetFieldData().GetAny(SalesFiscalRequestFieldBuyerSnapshot)
}

func (this *SalesFiscalRequest) SetBuyerSnapshot(snapshot any) {
	this.GetFieldData().SetAny(SalesFiscalRequestFieldBuyerSnapshot, snapshot)
}

// IsIssued reports whether the provider confirmed the document exists. A request that has not come
// back is pending, not issued: reporting early would tell a customer they hold a VAT invoice that
// does not exist and cannot be deducted.
func (this SalesFiscalRequest) IsIssued() bool {
	status := this.GetStatus()
	return status != nil && *status == string(SalesFiscalStatusIssued)
}

// IsInFlight reports whether the provider has been asked and has not yet answered. Distinct from
// failed, which Sales may safely resend: an in-flight request may already have issued a document
// Sales has not heard about, and resending it creates a duplicate legal document.
func (this SalesFiscalRequest) IsInFlight() bool {
	status := this.GetStatus()
	return status != nil && *status == string(SalesFiscalStatusPending)
}

// BlocksNewRequest reports whether this request stops another being raised for the same bill. Both
// pending and issued block — issued because a second original invoice is a tax filing to correct,
// pending because Sales does not know whether it became issued.
func (this SalesFiscalRequest) BlocksNewRequest() bool {
	return this.IsIssued() || this.IsInFlight()
}
