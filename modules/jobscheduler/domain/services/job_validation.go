package services

import (
	"regexp"
	"time"

	"github.com/sky-as-code/nikki-erp/common/cronexpr"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/constants"
	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/domain/models"
)

// ActionConfigValidator is the seam through which the domain service reaches the action
// executors without importing infra. The executors know what a valid action_config looks like
// for their own transport, and nothing else does.
type ActionConfigValidator interface {
	// ValidateActionConfig checks cfg for actionType, returning field-scoped errors or nil. An
	// unknown actionType is itself an error rather than a silent pass: a job whose action nothing
	// can run would be accepted, scheduled, and fail on every occurrence forever.
	ValidateActionConfig(actionType string, cfg map[string]any) *ft.ClientErrors
}

// EngineWaker lets the domain service tell the engine its horizon may have moved, without
// depending on the engine's concrete type. The engine is started after the services are built,
// so a concrete dependency here would be a cycle.
type EngineWaker interface {
	Wake(reason WakeReason)
}

// moduleNamePattern matches the module name exactly as the platform writes it: alphanumeric and
// nothing else. It is deliberately stricter than the column, because module_name is compared for
// equality on every delete-by-module and a name differing only in punctuation would look right
// and delete nothing.
var moduleNamePattern = regexp.MustCompile(`^[A-Za-z0-9]+$`)

// ValidateJobRules applies the rules the model schema cannot express: the cron expression, the
// ordering of the effective period, and the action config.
//
// Everything expressible in the schema - lengths, enums, the retry interval floor of 5 seconds -
// is left to the schema, so that each rule has exactly one home.
func ValidateJobRules(job *models.Job, executors ActionConfigValidator, vErrs *ft.ClientErrors) {
	validateJobType(job, vErrs)
	validateCronExpression(job, vErrs)
	validateEffectivePeriod(job, vErrs)
	validateActionConfig(job, executors, vErrs)
}

// ValidateJobUpdateRules applies the same rules to an update, filling in from the stored row
// whatever the partial update did not send.
//
// A dynamic-model update is a partial merge, so validating the incoming fields alone would let
// an edit that changes only effective_until be checked against no effective_from at all - and
// pass, however wrong the resulting period.
func ValidateJobUpdateRules(
	input *models.Job, found *models.Job, executors ActionConfigValidator, vErrs *ft.ClientErrors,
) {
	merged := mergeJobForValidation(input, found)
	ValidateJobRules(merged, executors, vErrs)
}

// mergeJobForValidation overlays the incoming fields onto the stored ones so the rules see the
// row as it will be after the update, not as it was sent.
func mergeJobForValidation(input *models.Job, found *models.Job) *models.Job {
	if found == nil {
		return input
	}
	fields := dmodel.DynamicFields{}
	for name, value := range found.GetFieldData() {
		fields[name] = value
	}
	for name, value := range input.GetFieldData() {
		fields[name] = value
	}
	return models.NewJobFrom(fields)
}

func validateJobType(job *models.Job, vErrs *ft.ClientErrors) {
	jobType := job.GetJobType()
	if jobType == nil || *jobType == models.JobTypeTechnical {
		return
	}
	// The enum admits "user" because the column will hold it once user scheduling exists.
	// Accepting the value into the table now, with no engine able to run it, would produce jobs
	// that are registered and permanently silent.
	vErrs.Append(*ft.NewValidationError(
		models.JobFieldJobType,
		ft.ErrorKey("err_job_type_user_unsupported", constants.JobSchedulerModuleName),
		"only technical jobs are supported in this scope",
	))
}

func validateCronExpression(job *models.Job, vErrs *ft.ClientErrors) {
	expr := job.GetCronExpression()
	if expr == nil {
		return // The schema reports the omission.
	}
	if _, err := cronexpr.Parse(*expr); err != nil {
		// The parse error names the offending field and why, which is the difference between a
		// caller fixing their expression and a caller guessing at it.
		vErrs.Append(*ft.NewValidationError(
			models.JobFieldCronExpression,
			ft.ErrorKey("err_cron_invalid", constants.JobSchedulerModuleName),
			err.Error(),
		))
	}
}

// validateEffectivePeriod enforces that the period is non-empty.
//
// The interval is half-open, [from, until), so equal bounds describe a period in which nothing
// can ever fire. That is far more likely a mistake than an intention, and a job that silently
// never runs is the hardest kind to notice.
func validateEffectivePeriod(job *models.Job, vErrs *ft.ClientErrors) {
	from, until := job.GetEffectiveFrom(), job.GetEffectiveUntil()
	if from == nil || until == nil {
		return
	}
	if !until.After(*from) {
		vErrs.Append(*ft.NewValidationError(
			models.JobFieldEffectiveUntil,
			ft.ErrorKey("err_effective_period_invalid", constants.JobSchedulerModuleName),
			"effective_until must be strictly after effective_from",
		))
	}
}

func validateActionConfig(
	job *models.Job, executors ActionConfigValidator, vErrs *ft.ClientErrors,
) {
	actionType := job.GetActionType()
	if actionType == nil || executors == nil {
		return
	}
	vErrs.ConcatPtr(executors.ValidateActionConfig(*actionType, job.GetActionConfig()))
}

// validateModuleName guards the delete-by-module command, where the name arrives as a bare query
// parameter rather than through the schema.
//
// An empty name is rejected rather than treated as "all", because the two are one keystroke
// apart and only one of them is recoverable.
func validateModuleName(moduleName string) *ft.ClientErrors {
	if moduleNamePattern.MatchString(moduleName) {
		return nil
	}
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewValidationError(
		models.JobFieldModuleName,
		ft.ErrorKey("err_module_name_required", constants.JobSchedulerModuleName),
		"module_name is required and must be alphanumeric",
	))
	return vErrs
}

// applyNextRunAt computes when the job is next due, so that the engine can find it by index
// rather than by parsing every job's cron on every tick.
//
// A job that is disabled, or whose cron yields no further occurrence, gets a nil next_run_at.
// Nil is what the index treats as "never due", so this is also how a job is taken out of the
// engine's way without deleting it.
func applyNextRunAt(job *models.Job, cfg SchedulerConfig) {
	expr := job.GetCronExpression()
	if expr == nil {
		return
	}
	parsed, err := cronexpr.Parse(*expr)
	if err != nil {
		job.SetNextRunAt(nil)
		return
	}

	enabled := job.GetIsEnabled()
	if enabled != nil && !*enabled {
		job.SetNextRunAt(nil)
		return
	}

	job.SetNextRunAt(nextRunAtFrom(parsed, job, time.Now().UTC()))
}

// nextRunAtFrom starts the search at the later of now and effective_from, so a job registered
// ahead of time does not get a first occurrence that has already passed.
func nextRunAtFrom(
	parsed *cronexpr.CronExpr, job *models.Job, now time.Time,
) *model.ModelDateTime {
	after := now
	if from := job.GetEffectiveFrom(); from != nil && from.AfterT(after) {
		after = from.GoTime()
	}

	next, ok := parsed.Next(after)
	if !ok {
		return nil
	}
	// The period is half-open: an occurrence landing exactly on effective_until belongs to the
	// time after the job stops, not to its last minute.
	if until := job.GetEffectiveUntil(); until != nil && !until.AfterT(next) {
		return nil
	}
	return toModelDateTime(next)
}

func toModelDateTime(t time.Time) *model.ModelDateTime {
	converted := model.ModelDateTime(t.UTC())
	return &converted
}

func nowModelDateTime() model.ModelDateTime {
	return model.ModelDateTime(time.Now().UTC())
}
