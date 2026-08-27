package services

import (
	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/domain/models"
)

// JobSnapshot is the configuration an execution runs under, frozen when the execution is created.
//
// It exists so that editing or deleting a job does not change how work already in flight behaves,
// and so that history stays readable after the job is deleted: an execution keeps the module name
// and job key it came from even once its job_id has been nulled.
//
// Every value that affects the execution's lifecycle is resolved here, at creation. MaxAttempts
// and RetryIntervalSeconds have already had the configured default applied, so a later change to
// that configuration cannot lengthen or shorten a retry chain that has already begun.
//
// The attempt timeout and the backoff ceiling are deliberately NOT here. They are system safety
// limits rather than per-job behaviour: an operator lowering the ceiling to shed load needs it to
// take effect on work already queued, which is exactly the case where a frozen value would be
// wrong.
type JobSnapshot struct {
	JobId          string         `json:"job_id"`
	JobKey         string         `json:"job_key"`
	ModuleName     string         `json:"module_name"`
	CronExpression string         `json:"cron_expression"`
	EffectiveUntil *string        `json:"effective_until"`
	ActionType     string         `json:"action_type"`
	ActionConfig   map[string]any `json:"action_config"`

	// EffectiveMaxAttempts and EffectiveRetryIntervalSeconds are named "effective" because they
	// are the resolved values rather than the job's own nullable overrides: a reader of the
	// snapshot never has to know which came from the job and which from configuration.
	EffectiveMaxAttempts          int `json:"effective_max_attempts"`
	EffectiveRetryIntervalSeconds int `json:"effective_retry_interval_seconds"`

	ConcurrencyPolicy string `json:"concurrency_policy"`
	MisfirePolicy     string `json:"misfire_policy"`
}

// BuildJobSnapshot resolves a job's configuration against the application defaults and freezes it.
//
// A job leaves its overridable settings null when it does not care, which is the common case; the
// default is applied here rather than at create time on purpose. Applying it at create would bake
// the then-current configuration into the row forever, so a later change to the default would
// never reach existing jobs - the opposite of what a default is for.
func BuildJobSnapshot(job models.Job, cfg SchedulerConfig) JobSnapshot {
	snapshot := JobSnapshot{
		EffectiveMaxAttempts:          cfg.DefaultMaxAttempts,
		EffectiveRetryIntervalSeconds: cfg.DefaultRetryIntervalSecs,
		ConcurrencyPolicy:             cfg.DefaultConcurrencyPolicy,
		MisfirePolicy:                 cfg.DefaultMisfirePolicy,
		ActionConfig:                  job.GetActionConfig(),
	}

	if id := job.GetId(); id != nil {
		snapshot.JobId = string(*id)
	}
	if key := job.GetJobKey(); key != nil {
		snapshot.JobKey = *key
	}
	if module := job.GetModuleName(); module != nil {
		snapshot.ModuleName = *module
	}
	if cron := job.GetCronExpression(); cron != nil {
		snapshot.CronExpression = *cron
	}
	if actionType := job.GetActionType(); actionType != nil {
		snapshot.ActionType = *actionType
	}
	if until := job.GetEffectiveUntil(); until != nil {
		value := until.String()
		snapshot.EffectiveUntil = &value
	}

	if max := job.GetMaxAttempts(); max != nil {
		snapshot.EffectiveMaxAttempts = int(*max)
	}
	if interval := job.GetRetryIntervalSeconds(); interval != nil {
		snapshot.EffectiveRetryIntervalSeconds = int(*interval)
	}
	if policy := job.GetConcurrencyPolicy(); policy != nil {
		snapshot.ConcurrencyPolicy = *policy
	}
	if policy := job.GetMisfirePolicy(); policy != nil {
		snapshot.MisfirePolicy = *policy
	}

	return snapshot
}

// ForbidsOverlap reports whether this execution's job may have only one occurrence open at a time.
func (this JobSnapshot) ForbidsOverlap() bool {
	return this.ConcurrencyPolicy == models.ConcurrencyForbidOverlap
}

// SkipsMisfires reports whether a missed occurrence should be dropped rather than run late.
func (this JobSnapshot) SkipsMisfires() bool {
	return this.MisfirePolicy == models.MisfireSkip
}
