package services

import (
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/domain/models"
)

// The JSON schemas extend core mixins by name, which the app registers at start-up.
func TestMain(m *testing.M) {
	_ = basemodel.RegisterJsonBaseSchemas()
	os.Exit(m.Run())
}

func newTestJob(t *testing.T) *models.Job {
	t.Helper()
	job := models.NewJob()
	id := model.Id("01M2JBJ0000000001000000000")
	job.SetId(&id)
	job.SetName(strPtr("Rebuild snapshot"))
	job.SetJobType(strPtr(models.JobTypeTechnical))
	job.SetModuleName(strPtr("inventory"))
	job.SetJobKey(strPtr("rebuild_snapshot"))
	job.SetActionType(strPtr(models.ActionTypeCommandBus))
	job.SetActionConfig(map[string]any{
		"type": "command_bus", "command_name": "inventory_maintenance.rebuild",
	})
	job.SetCronExpression(strPtr("*/15 * * * *"))
	return job
}

func strPtr(v string) *string { return &v }
func i32Ptr(v int32) *int32   { return &v }

func testConfig() SchedulerConfig {
	return SchedulerConfig{
		DefaultMaxAttempts:       3,
		DefaultRetryIntervalSecs: 10,
		DefaultConcurrencyPolicy: models.ConcurrencyForbidOverlap,
		DefaultMisfirePolicy:     models.MisfireRunOnce,
	}
}

// AC-8 through AC-11: a job that overrides nothing takes the configured defaults, and the snapshot
// records the resolved values rather than the nulls, so nothing downstream has to re-resolve them.
func TestSnapshotResolvesConfiguredDefaultsForAJobThatOverridesNothing(t *testing.T) {
	job := newTestJob(t)

	snapshot := BuildJobSnapshot(*job, testConfig())

	assert.Equal(t, 3, snapshot.EffectiveMaxAttempts, "AC-8")
	assert.Equal(t, 10, snapshot.EffectiveRetryIntervalSeconds, "AC-9")
	assert.Equal(t, models.ConcurrencyForbidOverlap, snapshot.ConcurrencyPolicy, "AC-10")
	assert.Equal(t, models.MisfireRunOnce, snapshot.MisfirePolicy, "AC-11")
}

func TestSnapshotPrefersTheJobsOwnOverrides(t *testing.T) {
	job := newTestJob(t)
	job.SetMaxAttempts(i32Ptr(7))
	job.SetRetryIntervalSeconds(i32Ptr(45))
	job.SetConcurrencyPolicy(strPtr(models.ConcurrencyAllowOverlap))
	job.SetMisfirePolicy(strPtr(models.MisfireSkip))

	snapshot := BuildJobSnapshot(*job, testConfig())

	assert.Equal(t, 7, snapshot.EffectiveMaxAttempts)
	assert.Equal(t, 45, snapshot.EffectiveRetryIntervalSeconds)
	assert.Equal(t, models.ConcurrencyAllowOverlap, snapshot.ConcurrencyPolicy)
	assert.Equal(t, models.MisfireSkip, snapshot.MisfirePolicy)
}

// AC-24: editing a job must not change work already in flight. This is the whole reason the
// snapshot exists, so it is asserted directly: build, mutate, and confirm the frozen copy did not
// move.
func TestEditingTheJobDoesNotChangeAnExistingSnapshot(t *testing.T) {
	job := newTestJob(t)
	job.SetMaxAttempts(i32Ptr(5))
	job.SetCronExpression(strPtr("*/15 * * * *"))

	snapshot := BuildJobSnapshot(*job, testConfig())

	job.SetMaxAttempts(i32Ptr(1))
	job.SetCronExpression(strPtr("0 0 * * *"))
	job.SetActionConfig(map[string]any{"type": "rest_api", "url": "https://elsewhere.example"})

	assert.Equal(t, 5, snapshot.EffectiveMaxAttempts,
		"a retry chain already under way keeps the budget it started with")
	assert.Equal(t, "*/15 * * * *", snapshot.CronExpression)
	assert.Equal(t, "inventory_maintenance.rebuild", snapshot.ActionConfig["command_name"])
}

// The snapshot carries the job's identity so history stays readable after the job is deleted and
// its job_id is nulled.
func TestSnapshotCarriesIdentityThatSurvivesJobDeletion(t *testing.T) {
	job := newTestJob(t)

	snapshot := BuildJobSnapshot(*job, testConfig())

	assert.Equal(t, "inventory", snapshot.ModuleName)
	assert.Equal(t, "rebuild_snapshot", snapshot.JobKey)
	assert.Equal(t, "01M2JBJ0000000001000000000", snapshot.JobId)
}

// The attempt timeout and the backoff ceiling are runtime configuration, not snapshot fields: an
// operator lowering the ceiling to shed load needs it to reach work already queued.
func TestSnapshotOmitsSystemSafetyLimits(t *testing.T) {
	job := newTestJob(t)

	snapshot := BuildJobSnapshot(*job, testConfig())

	// Expressed against the struct's shape rather than its values: these fields must not exist.
	assert.NotContains(t, snapshotFieldNames(t, snapshot), "AttemptTimeoutSecs")
	assert.NotContains(t, snapshotFieldNames(t, snapshot), "ExpBackoffMaxIntervalSecs")
}

func snapshotFieldNames(t *testing.T, snapshot JobSnapshot) []string {
	t.Helper()
	names := []string{}
	typ := reflect.TypeOf(snapshot)
	for i := 0; i < typ.NumField(); i++ {
		names = append(names, typ.Field(i).Name)
	}
	return names
}

func TestPolicyHelpersReadTheResolvedSnapshot(t *testing.T) {
	job := newTestJob(t)
	job.SetConcurrencyPolicy(strPtr(models.ConcurrencyForbidOverlap))
	job.SetMisfirePolicy(strPtr(models.MisfireSkip))

	snapshot := BuildJobSnapshot(*job, testConfig())

	require.True(t, snapshot.ForbidsOverlap())
	require.True(t, snapshot.SkipsMisfires())
}
