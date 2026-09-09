package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	SalesFulfillmentAttemptSchemaName = "sales_fulfillment_attempt"

	SalesFulfillmentAttemptFieldId                    = "id"
	SalesFulfillmentAttemptFieldOrgId                 = "org_id"
	SalesFulfillmentAttemptFieldFulfillmentId         = "fulfillment_id"
	SalesFulfillmentAttemptFieldAttemptNo             = "attempt_no"
	SalesFulfillmentAttemptFieldExecutorOutletId      = "executor_outlet_id"
	SalesFulfillmentAttemptFieldAttemptStatus         = "attempt_status"
	SalesFulfillmentAttemptFieldExternalCorrelationId = "external_correlation_id"
	SalesFulfillmentAttemptFieldResultEventId         = "result_event_id"
	SalesFulfillmentAttemptFieldResultPayloadHash     = "result_payload_hash"
	SalesFulfillmentAttemptFieldInventoryResultRef    = "inventory_result_ref"
	SalesFulfillmentAttemptFieldStartedAt             = "started_at"
	SalesFulfillmentAttemptFieldCompletedAt           = "completed_at"
)

//go:embed sales_fulfillment_attempt.json
var salesFulfillmentAttemptSchemaJson string

func SalesFulfillmentAttemptSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesFulfillmentAttemptSchemaJson)
}

type SalesFulfillmentAttempt struct {
	basemodel.DynamicModelBase
}

func NewSalesFulfillmentAttempt() *SalesFulfillmentAttempt {
	return &SalesFulfillmentAttempt{basemodel.NewDynamicModel()}
}

func NewSalesFulfillmentAttemptFrom(src dmodel.DynamicFields) *SalesFulfillmentAttempt {
	return &SalesFulfillmentAttempt{basemodel.NewDynamicModel(src)}
}

func (this SalesFulfillmentAttempt) GetId() *model.Id {
	return this.GetFieldData().GetModelId(basemodel.FieldId)
}

func (this *SalesFulfillmentAttempt) SetId(id *model.Id) {
	this.GetFieldData().SetModelId(basemodel.FieldId, id)
}

func (this SalesFulfillmentAttempt) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(SalesFulfillmentAttemptFieldOrgId)
}

func (this *SalesFulfillmentAttempt) SetOrgId(value *model.Id) {
	this.GetFieldData().SetModelId(SalesFulfillmentAttemptFieldOrgId, value)
}

func (this SalesFulfillmentAttempt) GetFulfillmentId() *model.Id {
	return this.GetFieldData().GetModelId(SalesFulfillmentAttemptFieldFulfillmentId)
}

func (this *SalesFulfillmentAttempt) SetFulfillmentId(value *model.Id) {
	this.GetFieldData().SetModelId(SalesFulfillmentAttemptFieldFulfillmentId, value)
}

func (this SalesFulfillmentAttempt) GetAttemptNo() *int32 {
	return this.GetFieldData().GetInt32(SalesFulfillmentAttemptFieldAttemptNo)
}

func (this *SalesFulfillmentAttempt) SetAttemptNo(value *int32) {
	this.GetFieldData().SetInt32(SalesFulfillmentAttemptFieldAttemptNo, value)
}

func (this SalesFulfillmentAttempt) GetExecutorOutletId() *model.Id {
	return this.GetFieldData().GetModelId(SalesFulfillmentAttemptFieldExecutorOutletId)
}

func (this *SalesFulfillmentAttempt) SetExecutorOutletId(value *model.Id) {
	this.GetFieldData().SetModelId(SalesFulfillmentAttemptFieldExecutorOutletId, value)
}

func (this SalesFulfillmentAttempt) GetAttemptStatus() *string {
	return this.GetFieldData().GetString(SalesFulfillmentAttemptFieldAttemptStatus)
}

func (this *SalesFulfillmentAttempt) SetAttemptStatus(value *string) {
	this.GetFieldData().SetString(SalesFulfillmentAttemptFieldAttemptStatus, value)
}

func (this SalesFulfillmentAttempt) GetExternalCorrelationId() *string {
	return this.GetFieldData().GetString(SalesFulfillmentAttemptFieldExternalCorrelationId)
}

func (this *SalesFulfillmentAttempt) SetExternalCorrelationId(value *string) {
	this.GetFieldData().SetString(SalesFulfillmentAttemptFieldExternalCorrelationId, value)
}

func (this SalesFulfillmentAttempt) GetResultEventId() *string {
	return this.GetFieldData().GetString(SalesFulfillmentAttemptFieldResultEventId)
}

func (this *SalesFulfillmentAttempt) SetResultEventId(value *string) {
	this.GetFieldData().SetString(SalesFulfillmentAttemptFieldResultEventId, value)
}

func (this SalesFulfillmentAttempt) GetResultPayloadHash() *string {
	return this.GetFieldData().GetString(SalesFulfillmentAttemptFieldResultPayloadHash)
}

func (this *SalesFulfillmentAttempt) SetResultPayloadHash(value *string) {
	this.GetFieldData().SetString(SalesFulfillmentAttemptFieldResultPayloadHash, value)
}

func (this SalesFulfillmentAttempt) GetInventoryResultRef() *string {
	return this.GetFieldData().GetString(SalesFulfillmentAttemptFieldInventoryResultRef)
}

func (this *SalesFulfillmentAttempt) SetInventoryResultRef(value *string) {
	this.GetFieldData().SetString(SalesFulfillmentAttemptFieldInventoryResultRef, value)
}

func (this SalesFulfillmentAttempt) GetStartedAt() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(SalesFulfillmentAttemptFieldStartedAt)
}

func (this *SalesFulfillmentAttempt) SetStartedAt(value *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(SalesFulfillmentAttemptFieldStartedAt, value)
}

func (this SalesFulfillmentAttempt) GetCompletedAt() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(SalesFulfillmentAttemptFieldCompletedAt)
}

func (this *SalesFulfillmentAttempt) SetCompletedAt(value *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(SalesFulfillmentAttemptFieldCompletedAt, value)
}

// DeriveAttemptStatus rolls one try's items up into its own outcome, on the same principle as the
// item result: derived, never asserted. Any dispensed quantity at all means goods left the machine,
// which is why a mixed outcome is `partially_succeeded` rather than `failed` — calling it a failure
// would invite a retry for quantity the customer is already holding.
func DeriveAttemptStatus(results []FulfillmentAttemptItemResult) FulfillmentAttemptStatus {
	if len(results) == 0 {
		return FulfillmentAttemptStatusPending
	}

	anyDelivered := false
	anyOutstanding := false
	for _, result := range results {
		switch result {
		case FulfillmentAttemptItemResultSuccess:
			anyDelivered = true
		case FulfillmentAttemptItemResultPartial:
			anyDelivered = true
			anyOutstanding = true
		case FulfillmentAttemptItemResultFailure:
			anyOutstanding = true
		default:
			// A pending item means the reporter did not answer for it. The attempt is not finished,
			// and reporting it as succeeded would close a try that is still owed an answer.
			return FulfillmentAttemptStatusPending
		}
	}

	switch {
	case anyDelivered && anyOutstanding:
		return FulfillmentAttemptStatusPartiallySucceeded
	case anyDelivered:
		return FulfillmentAttemptStatusSucceeded
	}
	return FulfillmentAttemptStatusFailed
}

// IsOutstanding reports that this attempt is still owed a result. It is the guard that stops a second
// attempt starting: two executors acting on one reservation would hand over goods paid for once.
func (this SalesFulfillmentAttempt) IsOutstanding() bool {
	status := FulfillmentAttemptStatus(derefString(this.GetAttemptStatus()))
	return status == FulfillmentAttemptStatusPending
}

// HasResult reports that a result has already been recorded against this attempt, which is what the
// idempotency check asks before deciding whether an arriving event is new or a redelivery.
func (this SalesFulfillmentAttempt) HasResult() bool {
	return derefString(this.GetResultEventId()) != ""
}
