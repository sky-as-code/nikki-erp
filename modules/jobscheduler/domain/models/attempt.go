package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	AttemptSchemaName = "jobscheduler_attempt"

	AttemptFieldId             = "id"
	AttemptFieldCreatedAt      = "created_at"
	AttemptFieldExecutionId    = "execution_id"
	AttemptFieldAttemptNumber  = "attempt_number"
	AttemptFieldStatus         = "status"
	AttemptFieldInstanceId     = "instance_id"
	AttemptFieldStartedAt      = "started_at"
	AttemptFieldFinishedAt     = "finished_at"
	AttemptFieldDurationMs     = "duration_ms"
	AttemptFieldNextRetryAt    = "next_retry_at"
	AttemptFieldLeaseExpiresAt = "lease_expires_at"
	AttemptFieldErrorCode      = "error_code"
	AttemptFieldErrorMessage   = "error_message"
	AttemptFieldHttpStatusCode = "http_status_code"

	AttemptEdgeExecution = "execution"
)

//go:embed attempt.json
var attemptSchemaJson string

func AttemptSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(attemptSchemaJson)
}

type Attempt struct {
	basemodel.DynamicModelBase
}

func NewAttempt() *Attempt {
	return &Attempt{basemodel.NewDynamicModel()}
}

func NewAttemptFrom(src dmodel.DynamicFields) *Attempt {
	return &Attempt{basemodel.NewDynamicModel(src)}
}

func (this Attempt) GetId() *model.Id {
	return this.GetFieldData().GetModelId(basemodel.FieldId)
}

func (this *Attempt) SetId(id *model.Id) {
	this.GetFieldData().SetModelId(basemodel.FieldId, id)
}

func (this Attempt) GetExecutionId() *model.Id {
	return this.GetFieldData().GetModelId(AttemptFieldExecutionId)
}

func (this *Attempt) SetExecutionId(v *model.Id) {
	this.GetFieldData().SetModelId(AttemptFieldExecutionId, v)
}

func (this Attempt) GetAttemptNumber() *int32 {
	return this.GetFieldData().GetInt32(AttemptFieldAttemptNumber)
}

func (this *Attempt) SetAttemptNumber(v *int32) {
	this.GetFieldData().SetInt32(AttemptFieldAttemptNumber, v)
}

func (this Attempt) GetStatus() *string {
	return this.GetFieldData().GetString(AttemptFieldStatus)
}

func (this *Attempt) SetStatus(v *string) {
	this.GetFieldData().SetString(AttemptFieldStatus, v)
}

func (this Attempt) GetInstanceId() *string {
	return this.GetFieldData().GetString(AttemptFieldInstanceId)
}

func (this *Attempt) SetInstanceId(v *string) {
	this.GetFieldData().SetString(AttemptFieldInstanceId, v)
}

func (this Attempt) GetStartedAt() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(AttemptFieldStartedAt)
}

func (this *Attempt) SetStartedAt(v *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(AttemptFieldStartedAt, v)
}

func (this Attempt) GetFinishedAt() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(AttemptFieldFinishedAt)
}

func (this *Attempt) SetFinishedAt(v *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(AttemptFieldFinishedAt, v)
}

func (this Attempt) GetDurationMs() *int64 {
	return this.GetFieldData().GetInt64(AttemptFieldDurationMs)
}

func (this *Attempt) SetDurationMs(v *int64) {
	this.GetFieldData().SetInt64(AttemptFieldDurationMs, v)
}

func (this Attempt) GetNextRetryAt() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(AttemptFieldNextRetryAt)
}

func (this *Attempt) SetNextRetryAt(v *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(AttemptFieldNextRetryAt, v)
}

func (this Attempt) GetLeaseExpiresAt() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(AttemptFieldLeaseExpiresAt)
}

func (this *Attempt) SetLeaseExpiresAt(v *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(AttemptFieldLeaseExpiresAt, v)
}

func (this Attempt) GetErrorCode() *string {
	return this.GetFieldData().GetString(AttemptFieldErrorCode)
}

func (this *Attempt) SetErrorCode(v *string) {
	this.GetFieldData().SetString(AttemptFieldErrorCode, v)
}

func (this Attempt) GetErrorMessage() *string {
	return this.GetFieldData().GetString(AttemptFieldErrorMessage)
}

func (this *Attempt) SetErrorMessage(v *string) {
	this.GetFieldData().SetString(AttemptFieldErrorMessage, v)
}

func (this Attempt) GetHttpStatusCode() *int32 {
	return this.GetFieldData().GetInt32(AttemptFieldHttpStatusCode)
}

func (this *Attempt) SetHttpStatusCode(v *int32) {
	this.GetFieldData().SetInt32(AttemptFieldHttpStatusCode, v)
}

// GetCreatedAt is when this row was written.
//
// It is declared on the schema rather than inherited from core.basemodel.auditable_model,
// because that mixin also records who acted and a scheduler worker acts as nobody.
func (this Attempt) GetCreatedAt() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(AttemptFieldCreatedAt)
}

// SetCreatedAt must be called explicitly on the engine's write path: those rows go through
// baserepo.Insert, which applies no schema type defaults, so nothing else fills this in.
func (this *Attempt) SetCreatedAt(v *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(AttemptFieldCreatedAt, v)
}
