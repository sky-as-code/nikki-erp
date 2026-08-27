package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	ExecutionSchemaName = "jobscheduler_execution"

	ExecutionFieldId               = "id"
	ExecutionFieldCreatedAt        = "created_at"
	ExecutionFieldJobId            = "job_id"
	ExecutionFieldExecutionKey     = "execution_key"
	ExecutionFieldScheduledFor     = "scheduled_for"
	ExecutionFieldNextOccurrenceAt = "next_occurrence_at"
	ExecutionFieldStatus           = "status"
	ExecutionFieldAvailableAt      = "available_at"
	ExecutionFieldStartedAt        = "started_at"
	ExecutionFieldFinishedAt       = "finished_at"
	ExecutionFieldAttemptCount     = "attempt_count"
	ExecutionFieldJobSnapshot      = "job_snapshot"
	ExecutionFieldFailureCode      = "failure_code"

	ExecutionEdgeJob = "job"
)

//go:embed execution.json
var executionSchemaJson string

func ExecutionSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(executionSchemaJson)
}

type Execution struct {
	basemodel.DynamicModelBase
}

func NewExecution() *Execution {
	return &Execution{basemodel.NewDynamicModel()}
}

func NewExecutionFrom(src dmodel.DynamicFields) *Execution {
	return &Execution{basemodel.NewDynamicModel(src)}
}

func (this Execution) GetId() *model.Id {
	return this.GetFieldData().GetModelId(basemodel.FieldId)
}

func (this *Execution) SetId(id *model.Id) {
	this.GetFieldData().SetModelId(basemodel.FieldId, id)
}

func (this Execution) GetJobId() *model.Id {
	return this.GetFieldData().GetModelId(ExecutionFieldJobId)
}

func (this *Execution) SetJobId(v *model.Id) {
	this.GetFieldData().SetModelId(ExecutionFieldJobId, v)
}

func (this Execution) GetExecutionKey() *string {
	return this.GetFieldData().GetString(ExecutionFieldExecutionKey)
}

func (this *Execution) SetExecutionKey(v *string) {
	this.GetFieldData().SetString(ExecutionFieldExecutionKey, v)
}

func (this Execution) GetScheduledFor() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(ExecutionFieldScheduledFor)
}

func (this *Execution) SetScheduledFor(v *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(ExecutionFieldScheduledFor, v)
}

func (this Execution) GetNextOccurrenceAt() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(ExecutionFieldNextOccurrenceAt)
}

func (this *Execution) SetNextOccurrenceAt(v *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(ExecutionFieldNextOccurrenceAt, v)
}

func (this Execution) GetStatus() *string {
	return this.GetFieldData().GetString(ExecutionFieldStatus)
}

func (this *Execution) SetStatus(v *string) {
	this.GetFieldData().SetString(ExecutionFieldStatus, v)
}

func (this Execution) GetAvailableAt() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(ExecutionFieldAvailableAt)
}

func (this *Execution) SetAvailableAt(v *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(ExecutionFieldAvailableAt, v)
}

func (this Execution) GetStartedAt() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(ExecutionFieldStartedAt)
}

func (this *Execution) SetStartedAt(v *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(ExecutionFieldStartedAt, v)
}

func (this Execution) GetFinishedAt() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(ExecutionFieldFinishedAt)
}

func (this *Execution) SetFinishedAt(v *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(ExecutionFieldFinishedAt, v)
}

func (this Execution) GetAttemptCount() *int32 {
	return this.GetFieldData().GetInt32(ExecutionFieldAttemptCount)
}

func (this *Execution) SetAttemptCount(v *int32) {
	this.GetFieldData().SetInt32(ExecutionFieldAttemptCount, v)
}

func (this Execution) GetJobSnapshot() map[string]any {
	raw := this.GetFieldData().GetAny(ExecutionFieldJobSnapshot)
	if raw == nil {
		return nil
	}
	snapshot, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	return snapshot
}

func (this *Execution) SetJobSnapshot(v map[string]any) {
	this.GetFieldData().SetAny(ExecutionFieldJobSnapshot, v)
}

func (this Execution) GetFailureCode() *string {
	return this.GetFieldData().GetString(ExecutionFieldFailureCode)
}

func (this *Execution) SetFailureCode(v *string) {
	this.GetFieldData().SetString(ExecutionFieldFailureCode, v)
}

// GetCreatedAt is when this row was written.
//
// It is declared on the schema rather than inherited from core.basemodel.auditable_model,
// because that mixin also records who acted and a scheduler worker acts as nobody.
func (this Execution) GetCreatedAt() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(ExecutionFieldCreatedAt)
}

// SetCreatedAt must be called explicitly on the engine's write path: those rows go through
// baserepo.Insert, which applies no schema type defaults, so nothing else fills this in.
func (this *Execution) SetCreatedAt(v *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(ExecutionFieldCreatedAt, v)
}
