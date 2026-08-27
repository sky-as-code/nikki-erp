package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// The fiscal document request: what Sales asked an eInvoice provider for, and what came back.
//
// **A bill is not a VAT invoice** (BR 33). This table is the join between the two, and the reason it
// exists rather than a set of columns on sales_bills: a bill may need an original invoice and then
// several adjustments as goods come back, so one bill has MANY fiscal requests. Columns on the bill
// would have to be rewritten by the first adjustment, losing the original.
//
// Nothing here names a document type, a serial or a template. Sales states what commercially
// happened and the provider decides what document that requires (BR 50) - see the package comment on
// interfaces/external/invoicing.

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

// GetBuyerSnapshot returns the buyer's fiscal identity as supplied at issuance.
//
// Typed as `any` for the same reason the tax snapshot is: domain/models must not import another
// module, and a stored jsonmap comes back as whatever the JSON decoder chose. A caller needing the
// structured shape unmarshals it.
func (this SalesFiscalRequest) GetBuyerSnapshot() any {
	return this.GetFieldData().GetAny(SalesFiscalRequestFieldBuyerSnapshot)
}

func (this *SalesFiscalRequest) SetBuyerSnapshot(snapshot any) {
	this.GetFieldData().SetAny(SalesFiscalRequestFieldBuyerSnapshot, snapshot)
}

// IsIssued reports whether the provider confirmed the document exists.
//
// THE question this table answers, and the one BR 77 turns on. A request that has not come back is
// pending, not issued: reporting a document as issued before confirmation would tell a customer they
// hold a VAT invoice that does not exist, and a VAT invoice that does not exist cannot be deducted.
func (this SalesFiscalRequest) IsIssued() bool {
	status := this.GetStatus()
	return status != nil && *status == string(SalesFiscalStatusIssued)
}

// IsInFlight reports whether the provider has been asked and has not yet answered.
//
// Distinct from failed, and the distinction matters for retry: a failed request is one the provider
// refused and Sales may safely resend, while an in-flight one may already have issued a document
// Sales has not heard about. Resending that is how a duplicate legal document gets created.
func (this SalesFiscalRequest) IsInFlight() bool {
	status := this.GetStatus()
	return status != nil && *status == string(SalesFiscalStatusPending)
}

// BlocksNewRequest reports whether this request stops another being raised for the same bill.
//
// Both pending and issued block: issued because a second original invoice for one sale is a tax
// filing to correct, and pending because Sales does not know whether it became issued. Only failed
// and cancelled leave the bill free to ask again.
func (this SalesFiscalRequest) BlocksNewRequest() bool {
	return this.IsIssued() || this.IsInFlight()
}
