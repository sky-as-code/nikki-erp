package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	JobSchemaName = "jobscheduler_job"

	JobFieldId                   = "id"
	JobFieldName                 = "name"
	JobFieldJobType              = "job_type"
	JobFieldModuleName           = "module_name"
	JobFieldJobKey               = "job_key"
	JobFieldActionType           = "action_type"
	JobFieldActionConfig         = "action_config"
	JobFieldCronExpression       = "cron_expression"
	JobFieldEffectiveFrom        = "effective_from"
	JobFieldEffectiveUntil       = "effective_until"
	JobFieldIsEnabled            = "is_enabled"
	JobFieldMaxAttempts          = "max_attempts"
	JobFieldRetryIntervalSeconds = "retry_interval_seconds"
	JobFieldConcurrencyPolicy    = "concurrency_policy"
	JobFieldMisfirePolicy        = "misfire_policy"
	JobFieldNextRunAt            = "next_run_at"
)

//go:embed job.json
var jobSchemaJson string

func JobSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(jobSchemaJson)
}

type Job struct {
	basemodel.DynamicModelBase
}

func NewJob() *Job {
	return &Job{basemodel.NewDynamicModel()}
}

func NewJobFrom(src dmodel.DynamicFields) *Job {
	return &Job{basemodel.NewDynamicModel(src)}
}

func (this Job) GetId() *model.Id {
	return this.GetFieldData().GetModelId(basemodel.FieldId)
}

func (this *Job) SetId(id *model.Id) {
	this.GetFieldData().SetModelId(basemodel.FieldId, id)
}

func (this Job) GetName() *string {
	return this.GetFieldData().GetString(JobFieldName)
}

func (this *Job) SetName(v *string) {
	this.GetFieldData().SetString(JobFieldName, v)
}

func (this Job) GetJobType() *string {
	return this.GetFieldData().GetString(JobFieldJobType)
}

func (this *Job) SetJobType(v *string) {
	this.GetFieldData().SetString(JobFieldJobType, v)
}

func (this Job) GetModuleName() *string {
	return this.GetFieldData().GetString(JobFieldModuleName)
}

func (this *Job) SetModuleName(v *string) {
	this.GetFieldData().SetString(JobFieldModuleName, v)
}

func (this Job) GetJobKey() *string {
	return this.GetFieldData().GetString(JobFieldJobKey)
}

func (this *Job) SetJobKey(v *string) {
	this.GetFieldData().SetString(JobFieldJobKey, v)
}

func (this Job) GetActionType() *string {
	return this.GetFieldData().GetString(JobFieldActionType)
}

func (this *Job) SetActionType(v *string) {
	this.GetFieldData().SetString(JobFieldActionType, v)
}

func (this Job) GetActionConfig() map[string]any {
	raw := this.GetFieldData().GetAny(JobFieldActionConfig)
	if raw == nil {
		return nil
	}
	cfg, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	return cfg
}

func (this *Job) SetActionConfig(v map[string]any) {
	this.GetFieldData().SetAny(JobFieldActionConfig, v)
}

func (this Job) GetCronExpression() *string {
	return this.GetFieldData().GetString(JobFieldCronExpression)
}

func (this *Job) SetCronExpression(v *string) {
	this.GetFieldData().SetString(JobFieldCronExpression, v)
}

func (this Job) GetEffectiveFrom() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(JobFieldEffectiveFrom)
}

func (this *Job) SetEffectiveFrom(v *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(JobFieldEffectiveFrom, v)
}

func (this Job) GetEffectiveUntil() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(JobFieldEffectiveUntil)
}

func (this *Job) SetEffectiveUntil(v *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(JobFieldEffectiveUntil, v)
}

func (this Job) GetIsEnabled() *bool {
	return this.GetFieldData().GetBool(JobFieldIsEnabled)
}

func (this *Job) SetIsEnabled(v *bool) {
	this.GetFieldData().SetBool(JobFieldIsEnabled, v)
}

func (this Job) GetMaxAttempts() *int32 {
	return this.GetFieldData().GetInt32(JobFieldMaxAttempts)
}

func (this *Job) SetMaxAttempts(v *int32) {
	this.GetFieldData().SetInt32(JobFieldMaxAttempts, v)
}

func (this Job) GetRetryIntervalSeconds() *int32 {
	return this.GetFieldData().GetInt32(JobFieldRetryIntervalSeconds)
}

func (this *Job) SetRetryIntervalSeconds(v *int32) {
	this.GetFieldData().SetInt32(JobFieldRetryIntervalSeconds, v)
}

func (this Job) GetConcurrencyPolicy() *string {
	return this.GetFieldData().GetString(JobFieldConcurrencyPolicy)
}

func (this *Job) SetConcurrencyPolicy(v *string) {
	this.GetFieldData().SetString(JobFieldConcurrencyPolicy, v)
}

func (this Job) GetMisfirePolicy() *string {
	return this.GetFieldData().GetString(JobFieldMisfirePolicy)
}

func (this *Job) SetMisfirePolicy(v *string) {
	this.GetFieldData().SetString(JobFieldMisfirePolicy, v)
}

func (this Job) GetNextRunAt() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(JobFieldNextRunAt)
}

func (this *Job) SetNextRunAt(v *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(JobFieldNextRunAt, v)
}
