package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// One try at issuing a document for one billing instruction.
//
// A table rather than a counter on the instruction, because the interesting cases are the ones a
// counter cannot answer: which attempt produced the document that exists, and whether an attempt
// that never reported back left one behind. The provider_request_id is what makes the second
// question answerable at all.

const (
	SalesBillingIssuanceAttemptSchemaName = "sales_billing_issuance_attempt"

	SalesBillingIssuanceAttemptFieldId            = basemodel.FieldId
	SalesBillingIssuanceAttemptFieldOrgId         = basemodel.FieldOrgId
	SalesBillingIssuanceAttemptFieldInstructionId = "billing_instruction_id"
	SalesBillingIssuanceAttemptFieldAttemptNo     = "attempt_no"
	SalesBillingIssuanceAttemptFieldStatus        = "status"
	SalesBillingIssuanceAttemptFieldStartedAt     = "started_at"
	SalesBillingIssuanceAttemptFieldCompletedAt   = "completed_at"
	SalesBillingIssuanceAttemptFieldProviderReqId = "provider_request_id"
	SalesBillingIssuanceAttemptFieldProviderRef   = "provider_invoice_reference"
	SalesBillingIssuanceAttemptFieldErrorCode     = "error_code"
	SalesBillingIssuanceAttemptFieldErrorMessage  = "error_message"
)

//go:embed sales_billing_issuance_attempt.json
var salesBillingIssuanceAttemptSchemaJson string

func SalesBillingIssuanceAttemptSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesBillingIssuanceAttemptSchemaJson)
}

// SalesBillingIssuanceAttempt is one try at raising a document.
type SalesBillingIssuanceAttempt struct {
	basemodel.DynamicModelBase
}

func NewSalesBillingIssuanceAttempt() *SalesBillingIssuanceAttempt {
	return &SalesBillingIssuanceAttempt{basemodel.NewDynamicModel(dmodel.DynamicFields{})}
}

func NewSalesBillingIssuanceAttemptFrom(src dmodel.DynamicFields) *SalesBillingIssuanceAttempt {
	return &SalesBillingIssuanceAttempt{basemodel.NewDynamicModel(src)}
}

func (this SalesBillingIssuanceAttempt) GetId() *model.Id {
	return this.GetFieldData().GetModelId(SalesBillingIssuanceAttemptFieldId)
}

func (this SalesBillingIssuanceAttempt) GetInstructionId() *model.Id {
	return this.GetFieldData().GetModelId(SalesBillingIssuanceAttemptFieldInstructionId)
}

func (this SalesBillingIssuanceAttempt) GetAttemptNo() *int32 {
	return this.GetFieldData().GetInt32(SalesBillingIssuanceAttemptFieldAttemptNo)
}

func (this SalesBillingIssuanceAttempt) GetStatus() *string {
	return this.GetFieldData().GetString(SalesBillingIssuanceAttemptFieldStatus)
}

func (this SalesBillingIssuanceAttempt) GetProviderRequestId() *string {
	return this.GetFieldData().GetString(SalesBillingIssuanceAttemptFieldProviderReqId)
}

// IsIndeterminate reports whether this attempt left the question open.
//
// It is the one status that must block a retry: the request reached the provider and the reply did
// not come back, so a document may or may not exist. Retrying without reconciling is how one sale
// ends up with two legal invoices.
func (this SalesBillingIssuanceAttempt) IsIndeterminate() bool {
	status := this.GetStatus()
	return status != nil && *status == string(SalesBillingAttemptStatusUnknown)
}
