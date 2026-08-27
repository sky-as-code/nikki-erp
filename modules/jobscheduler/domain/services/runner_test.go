package services

import (
	"encoding/json"
	stdErr "errors"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/domain/models"
)

func assertError(message string) error {
	return stdErr.New(message)
}

func isValidUtf8(value string) bool {
	return utf8.ValidString(value)
}

func runnerTestConfig() SchedulerConfig {
	return SchedulerConfig{
		DefaultMaxAttempts:        3,
		DefaultRetryIntervalSecs:  10,
		ExpBackoffMaxIntervalSecs: 300,
		AttemptTimeoutSecs:        60,
		LeaseSafetyMarginSecs:     30,
		ReconciliationInterval:    time.Minute,
		WorkerConcurrency:         4,
		ClaimBatchSize:            10,
		DefaultConcurrencyPolicy:  models.ConcurrencyForbidOverlap,
		DefaultMisfirePolicy:      models.MisfireRunOnce,
		DefaultJobEnabled:         true,
		HistoryRetentionDays:      30,
		MisfireThresholdSecs:      120,
	}
}

// The lease must outlast the attempt it protects, or a worker still doing its job is reaped and
// its work is run a second time somewhere else while the first is still running.
func TestLeaseOutlastsTheAttemptTimeout(t *testing.T) {
	cfg := runnerTestConfig()

	assert.Greater(t, cfg.LeaseDuration(), cfg.AttemptTimeout(),
		"a lease no longer than the timeout would reap workers that are merely slow")
	assert.Equal(t, 90*time.Second, cfg.LeaseDuration())
}

// The snapshot must survive a round trip through the jsonmap column unchanged. If it did not, an
// execution would run under different settings from the ones frozen when it was created, which
// is the one thing the snapshot exists to prevent.
func TestSnapshotSurvivesTheRoundTrip(t *testing.T) {
	original := JobSnapshot{
		JobId:                         "01M2JBJ0000000001000000000",
		JobKey:                        "rebuild_snapshot",
		ModuleName:                    "inventory",
		CronExpression:                "*/15 * * * *",
		ActionType:                    models.ActionTypeCommandBus,
		ActionConfig:                  map[string]any{"command_name": "inventory_maintenance.rebuild"},
		EffectiveMaxAttempts:          5,
		EffectiveRetryIntervalSeconds: 20,
		ConcurrencyPolicy:             models.ConcurrencyForbidOverlap,
		MisfirePolicy:                 models.MisfireRunOnce,
	}

	asMap, err := snapshotAsMap(original)
	require.NoError(t, err)

	execution := models.NewExecution()
	execution.SetJobSnapshot(asMap)
	restored, err := snapshotOf(*execution)
	require.NoError(t, err)

	assert.Equal(t, original, restored)
}

// The retry chain is driven by the snapshot's frozen numbers, not by current configuration.
// Editing the job after the fact must not lengthen or shorten a chain already under way.
func TestRetryUsesTheFrozenQuotaNotTheCurrentConfig(t *testing.T) {
	snapshot := JobSnapshot{EffectiveMaxAttempts: 5, EffectiveRetryIntervalSeconds: 10}
	cfg := runnerTestConfig() // DefaultMaxAttempts is 3

	decision := DecideRetry(RetryInput{
		AttemptNumber:     4, // beyond the config default, within the frozen quota
		MaxAttempts:       snapshot.EffectiveMaxAttempts,
		RetryIntervalSecs: snapshot.EffectiveRetryIntervalSeconds,
		MaxIntervalSecs:   cfg.ExpBackoffMaxIntervalSecs,
		FinishedAt:        time.Date(2026, 8, 24, 10, 0, 0, 0, time.UTC),
		Retryable:         true,
	})

	assert.True(t, decision.ShouldRetry,
		"attempt 4 of a frozen quota of 5 must retry even though the config default is 3")
}

func TestSnapshotDecodesFromRawBytes(t *testing.T) {
	encoded, err := json.Marshal(JobSnapshot{
		ActionType:           models.ActionTypeRestApi,
		EffectiveMaxAttempts: 7,
	})
	require.NoError(t, err)

	decoded, err := decodeSnapshot(encoded)

	require.NoError(t, err)
	assert.Equal(t, models.ActionTypeRestApi, decoded.ActionType)
	assert.Equal(t, 7, decoded.EffectiveMaxAttempts)
}

// An empty snapshot is not an error. It decodes to the zero value, whose zero MaxAttempts makes
// the very first attempt terminal - which is the right outcome for an execution whose
// configuration is missing, and is reached without a special case.
func TestEmptySnapshotDecodesToTheZeroValue(t *testing.T) {
	decoded, err := decodeSnapshot(nil)

	require.NoError(t, err)
	assert.Zero(t, decoded.EffectiveMaxAttempts)

	decision := DecideRetry(RetryInput{
		AttemptNumber: 1, MaxAttempts: decoded.EffectiveMaxAttempts, Retryable: true,
	})
	assert.False(t, decision.ShouldRetry)
	assert.Equal(t, models.FailureMaxAttemptsReached, decision.FailureCode)
}

func TestMalformedSnapshotIsAnError(t *testing.T) {
	_, err := decodeSnapshot([]byte("{not json"))

	assert.Error(t, err, "a snapshot that will not parse must not silently become the zero value")
}

// Two instances materializing the same occurrence is expected, not exceptional: the unique index
// on execution_key is what makes the second one a no-op. Recognizing that error is what stops it
// being logged as a tick failure every minute.
func TestDuplicateKeyErrorsAreRecognized(t *testing.T) {
	for _, message := range []string{
		`ERROR: duplicate key value violates unique constraint "jobsched_exec_key_idx"`,
		`pq: duplicate key value violates unique constraint`,
		`failed to insert: SQLSTATE 23505`,
		`UNIQUE constraint failed`,
	} {
		assert.True(t, isDuplicateKey(assertError(message)),
			"should recognize %q as a duplicate key", message)
	}
}

func TestUnrelatedErrorsAreNotMistakenForDuplicates(t *testing.T) {
	for _, message := range []string{
		"connection refused",
		"context deadline exceeded",
		`null value in column "status" violates not-null constraint`,
	} {
		assert.False(t, isDuplicateKey(assertError(message)),
			"should not treat %q as a duplicate key", message)
	}
	assert.False(t, isDuplicateKey(nil))
}

// A verbose error must still produce a recorded attempt. Truncating here rather than letting the
// insert fail keeps the history of what ran, which matters more than the tail of a stack trace.
func TestErrorMessagesAreTruncatedToTheColumn(t *testing.T) {
	long := make([]byte, maxErrorMessageLen*2)
	for i := range long {
		long[i] = 'x'
	}

	truncated := truncate(string(long), maxErrorMessageLen)

	assert.Len(t, truncated, maxErrorMessageLen)
}

// Cutting a UTF-8 string at a byte boundary produces bytes that are not valid text, which some
// drivers reject and others store as an unreadable tail.
func TestTruncationRespectsRuneBoundaries(t *testing.T) {
	// Three bytes per rune, so a byte limit of 10 falls mid-rune.
	value := "日本語日本語日本語日本語"

	truncated := truncate(value, 10)

	assert.LessOrEqual(t, len(truncated), 10)
	assert.True(t, isValidUtf8(truncated), "the result must be valid UTF-8, not a split rune")
}

func TestShortMessagesAreLeftAlone(t *testing.T) {
	assert.Equal(t, "boom", truncate("boom", maxErrorMessageLen))
	assert.Equal(t, "", truncate("", maxErrorMessageLen))
}

// available_at is the scheduled instant, except when that has already passed. A misfired
// occurrence must be claimable now rather than waiting for a past time to come round again.
func TestMisfiredOccurrenceIsAvailableImmediately(t *testing.T) {
	now := time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC)
	scheduledFor := now.Add(-30 * time.Minute)

	assert.Equal(t, now, maxTime(scheduledFor, now))
}

func TestFutureOccurrenceKeepsItsScheduledInstant(t *testing.T) {
	now := time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC)
	scheduledFor := now.Add(30 * time.Minute)

	assert.Equal(t, scheduledFor, maxTime(scheduledFor, now))
}

func TestTickErrorsCollapseToOne(t *testing.T) {
	assert.Nil(t, joinTickErrors(nil))
	assert.Nil(t, joinTickErrors([]error{}))

	joined := joinTickErrors([]error{assertError("first"), assertError("second")})

	require.Error(t, joined)
	assert.Contains(t, joined.Error(), "first")
	assert.Contains(t, joined.Error(), "second")
}
