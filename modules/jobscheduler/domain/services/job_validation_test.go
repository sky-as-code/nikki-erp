package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/domain/models"
)

// stubValidator accepts everything unless told otherwise, so a test about cron expressions is
// not also a test about action configs.
type stubValidator struct {
	rejectField string
	seenType    string
	seenConfig  map[string]any
	callCount   int
}

func (this *stubValidator) ValidateActionConfig(
	actionType string, config map[string]any,
) *ft.ClientErrors {
	this.callCount++
	this.seenType = actionType
	this.seenConfig = config
	if this.rejectField == "" {
		return nil
	}
	errs := ft.NewClientErrors()
	errs.Append(*ft.NewValidationError(this.rejectField, "test.rejected", "rejected"))
	return errs
}

func jobWithFields(fields map[string]any) *models.Job {
	data := dmodel.DynamicFields{}
	for name, value := range fields {
		data[name] = value
	}
	return models.NewJobFrom(data)
}

func mustTime(t *testing.T, value string) *model.ModelDateTime {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, value)
	require.NoError(t, err)
	return toModelDateTime(parsed)
}

func validate(job *models.Job, validator ActionConfigValidator) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()
	ValidateJobRules(job, validator, vErrs)
	return vErrs
}

func TestValidRegistrationPassesEveryRule(t *testing.T) {
	job := jobWithFields(map[string]any{
		models.JobFieldJobType:        models.JobTypeTechnical,
		models.JobFieldCronExpression: "*/5 * * * *",
		models.JobFieldActionType:     models.ActionTypeRestApi,
	})

	assert.Zero(t, validate(job, &stubValidator{}).Count())
}

func TestInvalidCronIsReportedOnItsOwnField(t *testing.T) {
	job := jobWithFields(map[string]any{
		models.JobFieldCronExpression: "*/5 * * *", // four fields, not five
	})

	errs := validate(job, &stubValidator{})

	require.Positive(t, errs.Count())
	assert.True(t, errs.Has(models.JobFieldCronExpression),
		"the error must name cron_expression, or the caller has to guess which field was wrong")
}

// Quartz users type "?" reflexively, and a six-field expression is the other common import
// mistake. Both must be refused at registration rather than at the first occurrence.
func TestUnsupportedCronDialectsAreRejected(t *testing.T) {
	for _, expr := range []string{
		"0 0 * * ?",    // Quartz day-of-week placeholder
		"0 0 12 * * *", // six fields: a seconds column
		"@daily",       // a descriptor, not an expression
		"0 0 L * *",    // last-day-of-month
		"0 0 * * MON",  // weekday name
	} {
		job := jobWithFields(map[string]any{models.JobFieldCronExpression: expr})

		errs := validate(job, &stubValidator{})

		assert.Positive(t, errs.Count(), "expression %q should be rejected", expr)
	}
}

// The period is half-open, so equal bounds describe a window in which nothing can ever fire.
// A job that is registered and permanently silent is the hardest kind of failure to notice.
func TestEffectivePeriodMustBeNonEmpty(t *testing.T) {
	job := jobWithFields(map[string]any{
		models.JobFieldCronExpression: "0 * * * *",
		models.JobFieldEffectiveFrom:  *mustTime(t, "2026-08-24T00:00:00Z"),
		models.JobFieldEffectiveUntil: *mustTime(t, "2026-08-24T00:00:00Z"),
	})

	errs := validate(job, &stubValidator{})

	require.Positive(t, errs.Count())
	assert.True(t, errs.Has(models.JobFieldEffectiveUntil))
}

func TestEffectiveUntilBeforeFromIsRejected(t *testing.T) {
	job := jobWithFields(map[string]any{
		models.JobFieldCronExpression: "0 * * * *",
		models.JobFieldEffectiveFrom:  *mustTime(t, "2026-08-24T10:00:00Z"),
		models.JobFieldEffectiveUntil: *mustTime(t, "2026-08-24T09:00:00Z"),
	})

	assert.Positive(t, validate(job, &stubValidator{}).Count())
}

func TestOneSidedEffectivePeriodIsAllowed(t *testing.T) {
	job := jobWithFields(map[string]any{
		models.JobFieldCronExpression: "0 * * * *",
		models.JobFieldEffectiveFrom:  *mustTime(t, "2026-08-24T10:00:00Z"),
	})

	assert.Zero(t, validate(job, &stubValidator{}).Count(),
		"a job with a start and no end runs indefinitely, which is the common case")
}

// The column admits "user" because it will hold it once user scheduling exists. Nothing can run
// such a job today, so accepting one would register a job that is silent forever.
func TestUserJobTypeIsRejectedInThisScope(t *testing.T) {
	job := jobWithFields(map[string]any{
		models.JobFieldJobType:        models.JobTypeUser,
		models.JobFieldCronExpression: "0 * * * *",
	})

	errs := validate(job, &stubValidator{})

	require.Positive(t, errs.Count())
	assert.True(t, errs.Has(models.JobFieldJobType))
}

func TestActionConfigIsDelegatedToTheExecutor(t *testing.T) {
	validator := &stubValidator{}
	config := map[string]any{"url": "https://example.test/hook"}
	job := jobWithFields(map[string]any{
		models.JobFieldCronExpression: "0 * * * *",
		models.JobFieldActionType:     models.ActionTypeRestApi,
		models.JobFieldActionConfig:   config,
	})

	validate(job, validator)

	assert.Equal(t, 1, validator.callCount)
	assert.Equal(t, models.ActionTypeRestApi, validator.seenType)
	assert.Equal(t, config, validator.seenConfig)
}

func TestExecutorRejectionSurfacesOnItsOwnField(t *testing.T) {
	job := jobWithFields(map[string]any{
		models.JobFieldCronExpression: "0 * * * *",
		models.JobFieldActionType:     models.ActionTypeCommandBus,
	})

	errs := validate(job, &stubValidator{rejectField: "action_config.command_name"})

	assert.True(t, errs.Has("action_config.command_name"))
}

// A partial update sending only effective_until must be checked against the stored
// effective_from. Validating the incoming fields alone would let an edit that inverts the period
// pass, because from its own point of view there is no period at all.
func TestUpdateValidatesAgainstTheMergedRow(t *testing.T) {
	stored := jobWithFields(map[string]any{
		models.JobFieldJobType:        models.JobTypeTechnical,
		models.JobFieldCronExpression: "0 * * * *",
		models.JobFieldEffectiveFrom:  *mustTime(t, "2026-08-24T10:00:00Z"),
	})
	update := jobWithFields(map[string]any{
		models.JobFieldEffectiveUntil: *mustTime(t, "2026-08-24T09:00:00Z"),
	})

	vErrs := ft.NewClientErrors()
	ValidateJobUpdateRules(update, stored, &stubValidator{}, vErrs)

	require.Positive(t, vErrs.Count(),
		"the inverted period is only visible once the update is merged onto the stored row")
	assert.True(t, vErrs.Has(models.JobFieldEffectiveUntil))
}

func TestUpdateDoesNotMutateTheStoredRow(t *testing.T) {
	stored := jobWithFields(map[string]any{
		models.JobFieldCronExpression: "0 * * * *",
	})
	update := jobWithFields(map[string]any{
		models.JobFieldCronExpression: "*/5 * * * *",
	})

	ValidateJobUpdateRules(update, stored, &stubValidator{}, ft.NewClientErrors())

	assert.Equal(t, "0 * * * *", *stored.GetCronExpression(),
		"merging for validation must not write back into the row that was read")
}

func TestModuleNameMustBeAlphanumeric(t *testing.T) {
	for _, name := range []string{"", " ", "my-module", "my_module", "my.module", "my module"} {
		assert.NotNil(t, validateModuleName(name), "module name %q should be rejected", name)
	}
	for _, name := range []string{"inventory", "paymentInvoice", "iam2"} {
		assert.Nil(t, validateModuleName(name), "module name %q should be accepted", name)
	}
}

// The empty name is the one that matters: it is one keystroke from a real name, and treating it
// as "all modules" would make a typo delete every registration in the system.
func TestEmptyModuleNameIsNeverTreatedAsAll(t *testing.T) {
	errs := validateModuleName("")

	require.NotNil(t, errs)
	assert.True(t, errs.Has(models.JobFieldModuleName))
}

func TestNextRunAtIsComputedFromTheCron(t *testing.T) {
	job := jobWithFields(map[string]any{
		models.JobFieldCronExpression: "*/5 * * * *",
		models.JobFieldIsEnabled:      true,
	})

	applyNextRunAt(job, SchedulerConfig{})

	next := job.GetNextRunAt()
	require.NotNil(t, next)
	assert.Zero(t, next.GoTime().Second(), "occurrences are truncated to the minute")
	assert.Zero(t, next.GoTime().Minute()%5)
	assert.True(t, next.GoTime().After(time.Now().UTC().Add(-time.Minute)))
}

// A disabled job holds no next_run_at, which is how it leaves the engine's index rather than
// being woken for work that will never be created.
func TestDisabledJobHasNoNextRun(t *testing.T) {
	job := jobWithFields(map[string]any{
		models.JobFieldCronExpression: "*/5 * * * *",
		models.JobFieldIsEnabled:      false,
	})

	applyNextRunAt(job, SchedulerConfig{})

	assert.Nil(t, job.GetNextRunAt())
}

// A registration made ahead of time must not get a first occurrence in the past, or the misfire
// path would fire it immediately - the opposite of what scheduling it ahead was for.
func TestNextRunStartsAtEffectiveFromWhenItIsInTheFuture(t *testing.T) {
	from := time.Now().UTC().Add(48 * time.Hour).Truncate(time.Hour)
	job := jobWithFields(map[string]any{
		models.JobFieldCronExpression: "0 * * * *",
		models.JobFieldEffectiveFrom:  model.ModelDateTime(from),
	})

	applyNextRunAt(job, SchedulerConfig{})

	next := job.GetNextRunAt()
	require.NotNil(t, next)
	assert.False(t, next.GoTime().Before(from))
}

// The period is half-open, so a job whose window has already closed parks rather than firing one
// last time on the boundary.
func TestNextRunIsNilWhenTheWindowHasClosed(t *testing.T) {
	job := jobWithFields(map[string]any{
		models.JobFieldCronExpression: "0 * * * *",
		models.JobFieldEffectiveUntil: model.ModelDateTime(time.Now().UTC().Add(-time.Hour)),
	})

	applyNextRunAt(job, SchedulerConfig{})

	assert.Nil(t, job.GetNextRunAt())
}

// A row predating a parser change must park rather than crash the caller. One unparseable job
// must not be able to stop every other job from being scheduled.
func TestUnparseableCronParksTheJobInsteadOfPanicking(t *testing.T) {
	job := jobWithFields(map[string]any{
		models.JobFieldCronExpression: "not a cron",
		models.JobFieldNextRunAt:      model.ModelDateTime(time.Now().UTC()),
	})

	assert.NotPanics(t, func() { applyNextRunAt(job, SchedulerConfig{}) })
	assert.Nil(t, job.GetNextRunAt())
}
