package models

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// The JSON documents extend the core base mixins by name, which CoreModule.RegisterModels does
// at app start-up. Without it every parse here panics on an unregistered mixin.
func TestMain(m *testing.M) {
	_ = basemodel.RegisterJsonBaseSchemas()
	os.Exit(m.Run())
}

// The schemas are parsed and validated at boot, so a malformed one panics the app rather than
// failing a request. Building them here turns that into a test failure instead: an unknown key,
// a string field missing min/max, or a bad enum shape fails on this line.
func TestSchemasBuild(t *testing.T) {
	jobSchema := JobSchemaBuilder().Build()
	assert.Equal(t, JobSchemaName, jobSchema.Name())
	assert.Equal(t, "jobscheduler_jobs", jobSchema.TableName())

	executionSchema := ExecutionSchemaBuilder().Build()
	assert.Equal(t, ExecutionSchemaName, executionSchema.Name())
	assert.Equal(t, "jobscheduler_executions", executionSchema.TableName())

	attemptSchema := AttemptSchemaBuilder().Build()
	assert.Equal(t, AttemptSchemaName, attemptSchema.Name())
	assert.Equal(t, "jobscheduler_attempts", attemptSchema.TableName())
}

func TestJobDeclaresEveryFieldTheSchedulerReads(t *testing.T) {
	jobSchema := JobSchemaBuilder().Build()

	for _, name := range []string{
		JobFieldName, JobFieldJobType, JobFieldModuleName, JobFieldJobKey,
		JobFieldActionType, JobFieldActionConfig, JobFieldCronExpression,
		JobFieldEffectiveFrom, JobFieldEffectiveUntil, JobFieldIsEnabled,
		JobFieldMaxAttempts, JobFieldRetryIntervalSeconds,
		JobFieldConcurrencyPolicy, JobFieldMisfirePolicy, JobFieldNextRunAt,
	} {
		_, ok := jobSchema.Field(name)
		assert.True(t, ok, "jobscheduler_job must declare %q", name)
	}
}

func TestExecutionAndAttemptDeclareEveryFieldTheSchedulerReads(t *testing.T) {
	executionSchema := ExecutionSchemaBuilder().Build()
	for _, name := range []string{
		ExecutionFieldJobId, ExecutionFieldExecutionKey, ExecutionFieldScheduledFor,
		ExecutionFieldNextOccurrenceAt, ExecutionFieldStatus, ExecutionFieldAvailableAt,
		ExecutionFieldStartedAt, ExecutionFieldFinishedAt, ExecutionFieldAttemptCount,
		ExecutionFieldJobSnapshot, ExecutionFieldFailureCode,
	} {
		_, ok := executionSchema.Field(name)
		assert.True(t, ok, "jobscheduler_execution must declare %q", name)
	}

	attemptSchema := AttemptSchemaBuilder().Build()
	for _, name := range []string{
		AttemptFieldExecutionId, AttemptFieldAttemptNumber, AttemptFieldStatus,
		AttemptFieldInstanceId, AttemptFieldStartedAt, AttemptFieldFinishedAt,
		AttemptFieldDurationMs, AttemptFieldNextRetryAt, AttemptFieldLeaseExpiresAt,
		AttemptFieldErrorCode, AttemptFieldErrorMessage, AttemptFieldHttpStatusCode,
	} {
		_, ok := attemptSchema.Field(name)
		assert.True(t, ok, "jobscheduler_attempt must declare %q", name)
	}
}

// All three are real tables, so all three must carry DB metadata. PrimaryKeys() is populated
// only by populateDbMetadata, which runs only under should_build_db.
func TestSchemasBuildDb(t *testing.T) {
	for _, tc := range []struct {
		name   string
		schema interface{ PrimaryKeys() []string }
	}{
		{"job", JobSchemaBuilder().Build()},
		{"execution", ExecutionSchemaBuilder().Build()},
		{"attempt", AttemptSchemaBuilder().Build()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.NotEmpty(t, tc.schema.PrimaryKeys(),
				"should_build_db must be set, or no table is generated")
		})
	}
}

// The scheduler's tables must never be tenant-scoped, and this guards it in the binary where
// it is possible to get wrong.
//
// A technical job is infrastructure owned by a module, not by a tenant, and its occurrences
// are materialized by a background worker that has no request and therefore no tenant behind
// it. In coremart, extending core.basemodel.base_model would pull in the tenant key, and
// AssertTenantId panics with "tenant ID is required" the first time such a row is written -
// at start-up, not in a test. That is why each schema declares its own id instead.
//
// Without this test, a later edit adding base_model back for convenience looks harmless and
// breaks the boot.
func TestSchemasAreNotTenantScoped(t *testing.T) {
	for _, tc := range []struct {
		name   string
		schema interface{ TenantKey() string }
	}{
		{"job", JobSchemaBuilder().Build()},
		{"execution", ExecutionSchemaBuilder().Build()},
		{"attempt", AttemptSchemaBuilder().Build()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Empty(t, tc.schema.TenantKey(),
				"must not carry a tenant key: scheduler rows are written with no tenant in scope")
		})
	}
}

// The corollary of declining base_model is that each schema owns its own primary key.
func TestSchemasDeclareTheirOwnPrimaryKey(t *testing.T) {
	for _, tc := range []struct {
		name   string
		schema interface{ PrimaryKeys() []string }
	}{
		{"job", JobSchemaBuilder().Build()},
		{"execution", ExecutionSchemaBuilder().Build()},
		{"attempt", AttemptSchemaBuilder().Build()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, []string{"id"}, tc.schema.PrimaryKeys())
		})
	}
}

// AC-1: a job carries no timezone. The scheduler is UTC-only by construction, and a per-job
// timezone would reintroduce the ambiguity the requirement removes.
func TestJobHasNoTimezoneField(t *testing.T) {
	jobSchema := JobSchemaBuilder().Build()

	_, ok := jobSchema.Field("timezone")
	assert.False(t, ok, "scheduling is UTC-only; a timezone field must not exist on the job")
}

// The job is explicitly not archivable, so it must not pick up the archivable mixin. If it
// ever did, the REST surface would grow an archive endpoint the requirement forbids.
func TestJobIsNotArchivable(t *testing.T) {
	jobSchema := JobSchemaBuilder().Build()

	_, ok := jobSchema.Field("archived_at")
	assert.False(t, ok, "the job must not be archivable")
}

// The job keeps optimistic concurrency, so two operators editing it cannot silently overwrite
// each other.
func TestJobIsVersioned(t *testing.T) {
	jobSchema := JobSchemaBuilder().Build()

	_, ok := jobSchema.Field("etag")
	assert.True(t, ok, "the job must carry an etag for optimistic concurrency")
}

func TestSchemaNamesMatchTheModuleNamespace(t *testing.T) {
	// The schema name doubles as the REST path segment and the IAM resource code, so a drift
	// here silently breaks authorization rather than failing loudly.
	assert.Equal(t, "jobscheduler_job", JobSchemaName)
	assert.Equal(t, "jobscheduler_execution", ExecutionSchemaName)
	assert.Equal(t, "jobscheduler_attempt", AttemptSchemaName)
}

func TestExecutionIsTerminalOnlyForFinishedStates(t *testing.T) {
	for _, status := range []string{
		ExecutionStatusSucceeded, ExecutionStatusFailed, ExecutionStatusCancelled,
	} {
		assert.True(t, IsExecutionTerminal(status), "%s is terminal", status)
	}

	for _, status := range []string{
		ExecutionStatusQueued, ExecutionStatusRunning, ExecutionStatusWaitingRetry,
	} {
		assert.False(t, IsExecutionTerminal(status),
			"%s is still open; treating it as terminal would let retention delete live work", status)
	}
}

// Every schema must fill created_at on insert without the caller supplying it.
//
// This is a regression test for a bug the whole unit suite missed: execution and attempt
// declared created_at as required_for_create but without use_type_default or auto_generated, so
// validation fell through every branch in ModelField.Validate and left the field empty. Nothing
// errored - the row simply reached Postgres with a NULL and was rejected by the not-null
// constraint, on every single materialization.
//
// The schemas decline core.basemodel.base_model to stay out of tenant scope, which means they
// give up the mixin that would otherwise have declared this correctly. That is exactly why it
// needs a test: the safety net the other modules rely on is one this module opted out of.
func TestCreatedAtIsFilledInWithoutTheCallerSupplyingIt(t *testing.T) {
	for _, tc := range []struct {
		name    string
		schema  *dmodel.ModelSchema
		minimal dmodel.DynamicFields
	}{
		{
			name:   "execution",
			schema: ExecutionSchemaBuilder().Build(),
			minimal: dmodel.DynamicFields{
				ExecutionFieldExecutionKey: "inventory:rebuild:2026-08-24T10:00:00Z",
				ExecutionFieldScheduledFor: "2026-08-24T10:00:00Z",
				ExecutionFieldStatus:       ExecutionStatusQueued,
				ExecutionFieldAvailableAt:  "2026-08-24T10:00:00Z",
				ExecutionFieldJobSnapshot:  map[string]any{"job_key": "rebuild"},
			},
		},
		{
			name:   "attempt",
			schema: AttemptSchemaBuilder().Build(),
			minimal: dmodel.DynamicFields{
				AttemptFieldExecutionId:   "01M2JBE0000000001000000000",
				AttemptFieldAttemptNumber: int32(1),
				AttemptFieldStatus:        AttemptStatusRunning,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sanitized, cErrs := tc.schema.Validate(tc.minimal)

			require.Zero(t, cErrs.Count(), "the minimal insert must validate: %v", cErrs)
			assert.NotEmpty(t, sanitized[basemodel.FieldCreatedAt],
				"created_at must be generated, or every insert fails the not-null constraint")
		})
	}
}

// The engine's rows must carry created_at before they reach the repository.
//
// This is the test that would have caught the real bug. The one above asserts that
// ModelSchema.Validate fills the field in - which is true, and irrelevant on the engine's write
// path: those rows go through baserepo.Insert, which applies no type defaults at all. The
// scheduler therefore has to set the value itself, and this asserts the accessor exists and
// round-trips so that a future refactor cannot quietly drop it again.
func TestEngineRowsCanCarryCreatedAtExplicitly(t *testing.T) {
	stamp := model.ModelDateTime(time.Date(2026, 8, 24, 10, 0, 0, 0, time.UTC))

	execution := NewExecution()
	execution.SetCreatedAt(&stamp)
	require.NotNil(t, execution.GetCreatedAt())
	assert.Equal(t, stamp.String(), execution.GetCreatedAt().String())
	assert.NotEmpty(t, execution.GetFieldData()[ExecutionFieldCreatedAt],
		"the value must land on the field the repository writes")

	attempt := NewAttempt()
	attempt.SetCreatedAt(&stamp)
	require.NotNil(t, attempt.GetCreatedAt())
	assert.NotEmpty(t, attempt.GetFieldData()[AttemptFieldCreatedAt])
}
