package repository

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/domain/models"
)

func mustExecutionSchema(t *testing.T) *dmodel.ModelSchema {
	t.Helper()
	_ = basemodel.RegisterJsonBaseSchemas()
	return models.ExecutionSchemaBuilder().Build()
}

// The claim must be a single statement. Selecting, committing and then updating would leave a
// window in which another instance sees the same row unclaimed, which is the double-claim the
// whole design exists to prevent.
func TestClaimQueryIsOneAtomicStatement(t *testing.T) {
	query, _ := buildClaimQuery(mustExecutionSchema(t), time.Now(), 20)

	assert.True(t, strings.HasPrefix(query, "UPDATE "),
		"the claim must be an UPDATE, not a SELECT followed by a later write")
	assert.Contains(t, query, "FOR UPDATE SKIP LOCKED")
	assert.Contains(t, query, "RETURNING")
	assert.Equal(t, 1, strings.Count(query, "UPDATE jobscheduler_executions"),
		"one statement, not two")
}

// SKIP LOCKED is what lets a second instance step over rows the first is holding. A plain
// FOR UPDATE would make it block and then claim nothing, serializing the entire fleet on one lock.
func TestClaimQueryUsesSkipLockedRatherThanPlainLocking(t *testing.T) {
	query, _ := buildClaimQuery(mustExecutionSchema(t), time.Now(), 20)

	assert.Contains(t, query, "SKIP LOCKED")
	assert.NotContains(t, query, "NOWAIT")
}

// available_at ties across executions of the same occurrence, so id has to break it. Without a
// total order, two instances can walk overlapping rows in opposite sequences and deadlock.
func TestClaimQueryOrdersDeterministically(t *testing.T) {
	query, _ := buildClaimQuery(mustExecutionSchema(t), time.Now(), 20)

	assert.Contains(t, query,
		"ORDER BY "+models.ExecutionFieldAvailableAt+" ASC, "+models.ExecutionFieldId+" ASC")
}

// AC-6 and AC-23: the configured batch size reaches the statement, and travels as a bound
// parameter rather than being pasted into the SQL text.
func TestClaimQueryBindsTheConfiguredBatchSize(t *testing.T) {
	for _, limit := range []int{1, 17, 250} {
		query, args := buildClaimQuery(mustExecutionSchema(t), time.Now(), limit)

		assert.Contains(t, query, "LIMIT $5")
		assert.NotContains(t, query, "LIMIT "+itoa(limit),
			"the limit must be bound, not interpolated")
		require.Len(t, args, 5)
		assert.Equal(t, limit, args[4])
	}
}

// Only work that is actually waiting may be claimed. Claiming a running execution would hand the
// same occurrence to a second worker while the first is still in it.
func TestClaimQueryTakesOnlyQueuedAndWaitingRetryWork(t *testing.T) {
	query, args := buildClaimQuery(mustExecutionSchema(t), time.Now(), 10)

	assert.Contains(t, query, models.ExecutionFieldStatus+" IN ($3, $4)")
	assert.Equal(t, models.ExecutionStatusQueued, args[2])
	assert.Equal(t, models.ExecutionStatusWaitingRetry, args[3])
	assert.Equal(t, models.ExecutionStatusRunning, args[0], "claimed work is marked running")
}

// An execution being retried started when its first attempt did. Overwriting started_at on each
// claim would lose how long the occurrence as a whole has been in flight.
func TestClaimQueryPreservesTheOriginalStartInstant(t *testing.T) {
	query, _ := buildClaimQuery(mustExecutionSchema(t), time.Now(), 10)

	assert.Contains(t, query,
		models.ExecutionFieldStartedAt+" = COALESCE("+models.ExecutionFieldStartedAt+", $2)")
}

// The scheduler's tables are not tenant-scoped in either binary, so no tenant predicate belongs
// here. This is asserted rather than merely commented because a reviewer familiar with
// stock_quant_lock.go will look for one and might "fix" its absence.
func TestClaimQueryHasNoTenantPredicate(t *testing.T) {
	schema := mustExecutionSchema(t)
	require.Empty(t, schema.TenantKey(), "precondition: executions carry no tenant key")

	query, _ := buildClaimQuery(schema, time.Now(), 10)

	assert.NotContains(t, query, "tenant_id")
}

// The claim is bound to the transaction that took the locks. On a pooled connection the locks
// would release the instant the statement returned, leaving a caller that believes it holds work
// it does not - a failure invisible until two instances collide in production.
func TestClaimingWithoutATransactionIsRefused(t *testing.T) {
	_, err := ClaimDueExecutions(nil, nil, time.Now(), 10)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ambient transaction")
}

func TestReapingWithoutATransactionIsRefused(t *testing.T) {
	_, err := ReapExpiredLeases(nil, nil, time.Now(), 10)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ambient transaction")
}

func TestLockingDueJobsWithoutATransactionIsRefused(t *testing.T) {
	_, err := LockDueJobs(nil, nil, time.Now(), 10)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ambient transaction")
}

// A batch size of zero would claim nothing on every tick and the scheduler would appear hung, so
// it is rejected at the boundary rather than silently producing an empty claim.
//
// The context here carries a transaction, so the batch-size check is what fails rather than the
// transaction guard - otherwise this test would pass for the wrong reason.
func TestNonPositiveBatchSizesAreRejected(t *testing.T) {
	ctx := corectx.NewRequestContext(context.Background())
	ctx.SetDbTranx(&fakeTransaction{})

	_, err := ClaimDueExecutions(ctx, nil, time.Now(), 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "batch size")

	_, err = ReapExpiredLeases(ctx, nil, time.Now(), -1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "batch size")
}

// fakeTransaction stands in for an ambient transaction so the guard sees one. Nothing calls
// through it: the functions under test fail on their batch size before reaching the database.
type fakeTransaction struct{}

func (this *fakeTransaction) Commit() error   { return nil }
func (this *fakeTransaction) Rollback() error { return nil }

// fakeRow feeds scan targets without a database, using the rowScanner seam.
type fakeRow struct {
	values []any
}

func (this *fakeRow) Scan(dest ...any) error {
	for i := range dest {
		switch target := dest[i].(type) {
		case *string:
			*target = this.values[i].(string)
		case **string:
			*target = this.values[i].(*string)
		case *time.Time:
			*target = this.values[i].(time.Time)
		case **time.Time:
			*target = this.values[i].(*time.Time)
		case **int32:
			*target = this.values[i].(*int32)
		case *int32:
			*target = this.values[i].(int32)
		case *[]byte:
			*target = this.values[i].([]byte)
		}
	}
	return nil
}

func TestScanClaimedExecutionNormalizesToUtc(t *testing.T) {
	jakarta := time.FixedZone("WIB", 7*60*60)
	scheduled := time.Date(2026, 8, 20, 17, 0, 0, 0, jakarta)
	nextOcc := scheduled.Add(time.Hour)
	jobId := "01M2JBJ0000000001000000000"
	count := int32(2)

	claimed, err := scanClaimedExecution(&fakeRow{values: []any{
		"01M2JBE0000000001000000000", &jobId, "mod:key:2026-08-20T10:00:00Z",
		scheduled, &nextOcc, &count, []byte(`{"job_key":"key"}`),
	}})

	require.NoError(t, err)
	assert.Equal(t, time.UTC, claimed.ScheduledFor.Location(),
		"a timestamp read in another zone must not leak that zone into the scheduler")
	require.NotNil(t, claimed.NextOccurrenceAt)
	assert.Equal(t, time.UTC, claimed.NextOccurrenceAt.Location())
	assert.Equal(t, int32(2), claimed.AttemptCount)
	require.NotNil(t, claimed.JobId)
	assert.Equal(t, jobId, string(*claimed.JobId))
}

// A deleted job leaves its executions behind with a null job_id, so the scan has to tolerate it.
// Anything downstream that needs the job's identity reads job_snapshot instead.
func TestScanClaimedExecutionAcceptsAnOrphanedExecution(t *testing.T) {
	scheduled := time.Date(2026, 8, 20, 10, 0, 0, 0, time.UTC)

	claimed, err := scanClaimedExecution(&fakeRow{values: []any{
		"01M2JBE0000000001000000000", (*string)(nil), "mod:key:2026-08-20T10:00:00Z",
		scheduled, (*time.Time)(nil), (*int32)(nil), []byte(`{}`),
	}})

	require.NoError(t, err)
	assert.Nil(t, claimed.JobId, "history outlives the job it came from")
	assert.Nil(t, claimed.NextOccurrenceAt, "no further occurrence means no retry ceiling")
	assert.Equal(t, int32(0), claimed.AttemptCount)
}

func itoa(value int) string {
	digits := ""
	if value == 0 {
		return "0"
	}
	for value > 0 {
		digits = string(rune('0'+value%10)) + digits
		value /= 10
	}
	return digits
}
