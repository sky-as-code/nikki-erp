// This file contains the only raw SQL in the Job Scheduler, and it is a deliberate exception to
// the engine-first rule in docs/wiki/07 §6.7 rather than drift.
//
// The reason is narrow and checkable: claiming due work across instances requires
// UPDATE ... WHERE id IN (SELECT ... FOR UPDATE SKIP LOCKED) RETURNING, and the query builder can
// emit none of the three parts it depends on. orm.QueryBuilder has no lock clause, no SKIP LOCKED,
// and no RETURNING, so there is no combination of engine calls that produces an atomic claim.
// Without one, two instances read the same due execution and both run it, which is precisely the
// double-claim the requirement exists to prevent.
//
// Everything else about jobs, executions and attempts still goes through the engine. Only the
// claim, the lease reaper and the timer's horizon query are hand-written, and only because the
// alternative is a race that no amount of application-level care can close.
//
// Note for a reviewer who knows modules/inventory/domain/services/stock_quant_lock.go: there is
// deliberately no tenant predicate anywhere in this file. The scheduler's three tables are not
// tenant-scoped in either binary - they decline the base_model mixin that carries the tenant key -
// because a technical job is infrastructure owned by a module rather than by a tenant, and its
// occurrences are materialized by a worker that has no tenant in scope.
package repository

import (
	"strings"
	"time"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"

	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/domain/models"
)

// ClaimedExecution is one execution this instance now owns.
//
// Its fields are read inside the claiming statement and are therefore current as of the claim,
// which is the only moment at which they can be trusted: a value read before the claim may belong
// to an execution another instance took.
type ClaimedExecution struct {
	Id               model.Id
	JobId            *model.Id
	ExecutionKey     string
	ScheduledFor     time.Time
	NextOccurrenceAt *time.Time
	AttemptCount     int32
	JobSnapshot      []byte
}

// DueJob is an enabled job whose next occurrence has arrived.
type DueJob struct {
	Id         model.Id
	ModuleName string
	JobKey     string
	NextRunAt  time.Time
}

// ReapedAttempt is an attempt whose lease ran out and which has just been marked abandoned.
type ReapedAttempt struct {
	Id            model.Id
	ExecutionId   model.Id
	AttemptNumber int32
	StartedAt     *time.Time
}

// errNoTransaction is returned by every function here that takes locks. On a pooled connection the
// locks would be released the instant the statement returned, and SKIP LOCKED would stop meaning
// anything - so a caller would believe it holds work it does not. Refusing is better than
// degrading, because the degraded mode is invisible until two instances collide in production.
func errNoTransaction(operation string) error {
	return errors.New(operation + " requires an ambient transaction: without one the row locks are " +
		"released as soon as the statement returns, and two instances would claim the same work")
}

// ClaimDueExecutions atomically takes up to `limit` executions that are due, marking each running
// and stamping the moment it started.
//
// Three properties make it correct, and none is optional:
//
//   - It is one statement. Selecting, committing and then updating would leave a window in which
//     another instance sees the same row unclaimed; here the subquery takes the locks and the outer
//     UPDATE writes under those same locks, with no gap between.
//
//   - SKIP LOCKED, not NOWAIT and not a plain FOR UPDATE. A plain FOR UPDATE makes the second
//     instance block until the first commits and then claim nothing, serializing the whole fleet on
//     one lock. SKIP LOCKED lets it step over the locked rows and take the next ones, which is what
//     turns N instances into N times the throughput rather than one instance with N-1 spectators.
//
//   - The order is deterministic. Two instances reaching overlapping rows walk them in the same
//     sequence and therefore queue rather than deadlock. available_at alone ties - many executions
//     share an occurrence - so id breaks it and gives a total order.
//
// The caller must insert the attempt row inside this same transaction. Committing the claim first
// and inserting afterwards reintroduces exactly the two-transaction pattern this function exists to
// avoid: the execution would be marked running with nothing recording who is running it.
func ClaimDueExecutions(
	ctx corectx.Context, repo dyn.BaseDynamicRepository, now time.Time, limit int,
) ([]ClaimedExecution, error) {
	if ctx == nil || ctx.GetDbTranx() == nil {
		return nil, errNoTransaction("ClaimDueExecutions")
	}
	if limit < 1 {
		return nil, errors.Errorf("claim batch size must be at least 1, got %d", limit)
	}

	query, args := buildClaimQuery(repo.Schema(), now, limit)
	rows, err := repo.ExtractClient(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to claim due executions")
	}
	defer rows.Close()

	claimed := make([]ClaimedExecution, 0, limit)
	for rows.Next() {
		execution, err := scanClaimedExecution(rows)
		if err != nil {
			return nil, err
		}
		claimed = append(claimed, execution)
	}
	return claimed, errors.Wrap(rows.Err(), "failed to read claimed executions")
}

// buildClaimQuery assembles the claiming UPDATE with positional placeholders.
//
// Table and column names come from the schema and this module's own constants, never from caller
// input, so there is nothing here an untrusted value can reach: every value travels as a bound
// argument. The limit is bound rather than interpolated for the same reason, even though it is an
// int - and so that the configured batch size is observably the one the statement uses.
func buildClaimQuery(schema *dmodel.ModelSchema, now time.Time, limit int) (string, []any) {
	table := schema.TableName()

	// started_at is preserved when already set: an execution being retried started when its first
	// attempt did, and overwriting it would lose how long the whole occurrence has been in flight.
	query := "UPDATE " + table + " SET " +
		models.ExecutionFieldStatus + " = $1, " +
		models.ExecutionFieldStartedAt + " = COALESCE(" + models.ExecutionFieldStartedAt + ", $2)" +
		" WHERE " + models.ExecutionFieldId + " IN (" +
		"SELECT " + models.ExecutionFieldId + " FROM " + table +
		" WHERE " + models.ExecutionFieldStatus + " IN ($3, $4)" +
		" AND " + models.ExecutionFieldAvailableAt + " <= $2" +
		" ORDER BY " + models.ExecutionFieldAvailableAt + " ASC, " + models.ExecutionFieldId + " ASC" +
		" LIMIT $5 FOR UPDATE SKIP LOCKED)" +
		" RETURNING " + strings.Join(claimedExecutionColumns(), ", ")

	args := []any{
		models.ExecutionStatusRunning,
		now.UTC(),
		models.ExecutionStatusQueued,
		models.ExecutionStatusWaitingRetry,
		limit,
	}
	return query, args
}

func claimedExecutionColumns() []string {
	return []string{
		models.ExecutionFieldId,
		models.ExecutionFieldJobId,
		models.ExecutionFieldExecutionKey,
		models.ExecutionFieldScheduledFor,
		models.ExecutionFieldNextOccurrenceAt,
		models.ExecutionFieldAttemptCount,
		models.ExecutionFieldJobSnapshot,
	}
}

// ReapExpiredLeases marks attempts whose lease has run out as abandoned and returns them, so the
// caller can apply the retry policy to each.
//
// It deliberately does not decide the retry itself. Putting the retry-window rule inside a SQL
// statement would make the single most important business rule in this module untestable without a
// database, and would leave the next person to touch the query having to rediscover it.
//
// An abandoned attempt is not the same as a failed one: nobody reported an outcome, the reaper
// inferred it from the silence. Recording that difference is what lets an operator tell a job that
// fails from an instance that died.
func ReapExpiredLeases(
	ctx corectx.Context, repo dyn.BaseDynamicRepository, now time.Time, limit int,
) ([]ReapedAttempt, error) {
	if ctx == nil || ctx.GetDbTranx() == nil {
		return nil, errNoTransaction("ReapExpiredLeases")
	}
	if limit < 1 {
		return nil, errors.Errorf("reap batch size must be at least 1, got %d", limit)
	}

	table := repo.Schema().TableName()
	query := "UPDATE " + table + " SET " +
		models.AttemptFieldStatus + " = $1, " +
		models.AttemptFieldFinishedAt + " = $2" +
		" WHERE " + models.AttemptFieldId + " IN (" +
		"SELECT " + models.AttemptFieldId + " FROM " + table +
		" WHERE " + models.AttemptFieldStatus + " = $3" +
		" AND " + models.AttemptFieldLeaseExpiresAt + " < $2" +
		" ORDER BY " + models.AttemptFieldLeaseExpiresAt + " ASC, " + models.AttemptFieldId + " ASC" +
		" LIMIT $4 FOR UPDATE SKIP LOCKED)" +
		" RETURNING " + models.AttemptFieldId + ", " + models.AttemptFieldExecutionId + ", " +
		models.AttemptFieldAttemptNumber + ", " + models.AttemptFieldStartedAt

	args := []any{
		models.AttemptStatusAbandoned,
		now.UTC(),
		models.AttemptStatusRunning,
		limit,
	}

	rows, err := repo.ExtractClient(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to reap expired leases")
	}
	defer rows.Close()

	reaped := make([]ReapedAttempt, 0, limit)
	for rows.Next() {
		attempt, err := scanReapedAttempt(rows)
		if err != nil {
			return nil, err
		}
		reaped = append(reaped, attempt)
	}
	return reaped, errors.Wrap(rows.Err(), "failed to read reaped attempts")
}

// LockDueJobs takes a row lock on enabled jobs whose next occurrence has arrived, so that only one
// instance materializes each occurrence.
//
// The lock is what keeps two instances from creating the same occurrence at the same moment. It is
// belt and braces rather than the only defence: execution_key is unique and derived from the job
// and the occurrence, so a duplicate insert would fail even without the lock. The lock turns that
// failure into an absence of contention, which is cheaper than an exception on every tick.
func LockDueJobs(
	ctx corectx.Context, repo dyn.BaseDynamicRepository, now time.Time, limit int,
) ([]DueJob, error) {
	if ctx == nil || ctx.GetDbTranx() == nil {
		return nil, errNoTransaction("LockDueJobs")
	}
	if limit < 1 {
		return nil, errors.Errorf("due-job batch size must be at least 1, got %d", limit)
	}

	table := repo.Schema().TableName()
	query := "SELECT " + models.JobFieldId + ", " + models.JobFieldModuleName + ", " +
		models.JobFieldJobKey + ", " + models.JobFieldNextRunAt +
		" FROM " + table +
		" WHERE " + models.JobFieldIsEnabled + " = true" +
		" AND " + models.JobFieldNextRunAt + " IS NOT NULL" +
		" AND " + models.JobFieldNextRunAt + " <= $1" +
		" ORDER BY " + models.JobFieldNextRunAt + " ASC, " + models.JobFieldId + " ASC" +
		" LIMIT $2 FOR UPDATE SKIP LOCKED"

	rows, err := repo.ExtractClient(ctx).Query(ctx, query, now.UTC(), limit)
	if err != nil {
		return nil, errors.Wrap(err, "failed to lock due jobs")
	}
	defer rows.Close()

	due := make([]DueJob, 0, limit)
	for rows.Next() {
		job, err := scanDueJob(rows)
		if err != nil {
			return nil, err
		}
		due = append(due, job)
	}
	return due, errors.Wrap(rows.Err(), "failed to read due jobs")
}

// EarliestPendingInstant returns the soonest moment at which the scheduler has anything to do:
// the earliest next_run_at among enabled jobs, or the earliest available_at among executions
// waiting to be claimed, whichever comes first. Nil means there is no scheduled work at all.
//
// This is the query that makes event-driven scheduling possible. The timer is set from its result,
// so the scheduler sleeps until something is actually due instead of waking every second to ask.
// Without it there is no alternative to polling.
//
// It takes no lock and needs no transaction: it reads a horizon, and a horizon that is a moment
// stale merely means the next tick recomputes it.
func EarliestPendingInstant(
	ctx corectx.Context, jobRepo dyn.BaseDynamicRepository, executionRepo dyn.BaseDynamicRepository,
) (*time.Time, error) {
	query := "SELECT MIN(instant) FROM (" +
		"SELECT MIN(" + models.JobFieldNextRunAt + ") AS instant FROM " + jobRepo.Schema().TableName() +
		" WHERE " + models.JobFieldIsEnabled + " = true AND " + models.JobFieldNextRunAt + " IS NOT NULL" +
		" UNION ALL " +
		"SELECT MIN(" + models.ExecutionFieldAvailableAt + ") AS instant FROM " +
		executionRepo.Schema().TableName() +
		" WHERE " + models.ExecutionFieldStatus + " IN ($1, $2)" +
		") AS horizons"

	rows, err := jobRepo.ExtractClient(ctx).Query(ctx, query,
		models.ExecutionStatusQueued, models.ExecutionStatusWaitingRetry)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read the scheduler horizon")
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, errors.Wrap(rows.Err(), "failed to read the scheduler horizon")
	}

	var earliest *time.Time
	if err := rows.Scan(&earliest); err != nil {
		return nil, errors.Wrap(err, "failed to scan the scheduler horizon")
	}
	if earliest != nil {
		utc := earliest.UTC()
		earliest = &utc
	}
	return earliest, errors.Wrap(rows.Err(), "failed to read the scheduler horizon")
}

// DeleteExpiredHistory removes finished executions older than the cutoff. Their attempts go with
// them through the foreign key.
//
// Only terminal executions are eligible, whatever their age: a job that has been running for
// longer than the retention period is still running, and deleting it would strand a worker holding
// a lease on a row that no longer exists.
func DeleteExpiredHistory(
	ctx corectx.Context, repo dyn.BaseDynamicRepository, cutoff time.Time, limit int,
) (int64, error) {
	if limit < 1 {
		return 0, errors.Errorf("retention batch size must be at least 1, got %d", limit)
	}

	table := repo.Schema().TableName()
	query := "DELETE FROM " + table +
		" WHERE " + models.ExecutionFieldId + " IN (" +
		"SELECT " + models.ExecutionFieldId + " FROM " + table +
		" WHERE " + models.ExecutionFieldStatus + " IN ($1, $2, $3)" +
		" AND " + models.ExecutionFieldFinishedAt + " IS NOT NULL" +
		" AND " + models.ExecutionFieldFinishedAt + " < $4" +
		" ORDER BY " + models.ExecutionFieldFinishedAt + " ASC" +
		" LIMIT $5)"

	result, err := repo.ExtractClient(ctx).Exec(ctx, query,
		models.ExecutionStatusSucceeded, models.ExecutionStatusFailed, models.ExecutionStatusCancelled,
		cutoff.UTC(), limit)
	if err != nil {
		return 0, errors.Wrap(err, "failed to delete expired execution history")
	}

	deleted, err := result.RowsAffected()
	return deleted, errors.Wrap(err, "failed to count deleted execution history")
}

// rowScanner is the part of the row cursor this file uses, named so the scans can be tested
// without a database.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanClaimedExecution(row rowScanner) (ClaimedExecution, error) {
	var (
		id, executionKey string
		jobId            *string
		scheduledFor     time.Time
		nextOccurrenceAt *time.Time
		attemptCount     *int32
		snapshot         []byte
	)
	err := row.Scan(&id, &jobId, &executionKey, &scheduledFor, &nextOccurrenceAt, &attemptCount, &snapshot)
	if err != nil {
		return ClaimedExecution{}, errors.Wrap(err, "failed to scan a claimed execution")
	}

	claimed := ClaimedExecution{
		Id:               model.Id(id),
		ExecutionKey:     executionKey,
		ScheduledFor:     scheduledFor.UTC(),
		NextOccurrenceAt: utcOrNil(nextOccurrenceAt),
		JobSnapshot:      snapshot,
	}
	if jobId != nil {
		claimed.JobId = ptrId(*jobId)
	}
	if attemptCount != nil {
		claimed.AttemptCount = *attemptCount
	}
	return claimed, nil
}

func scanReapedAttempt(row rowScanner) (ReapedAttempt, error) {
	var (
		id, executionId string
		attemptNumber   int32
		startedAt       *time.Time
	)
	if err := row.Scan(&id, &executionId, &attemptNumber, &startedAt); err != nil {
		return ReapedAttempt{}, errors.Wrap(err, "failed to scan a reaped attempt")
	}
	return ReapedAttempt{
		Id:            model.Id(id),
		ExecutionId:   model.Id(executionId),
		AttemptNumber: attemptNumber,
		StartedAt:     utcOrNil(startedAt),
	}, nil
}

func scanDueJob(row rowScanner) (DueJob, error) {
	var (
		id, moduleName, jobKey string
		nextRunAt              time.Time
	)
	if err := row.Scan(&id, &moduleName, &jobKey, &nextRunAt); err != nil {
		return DueJob{}, errors.Wrap(err, "failed to scan a due job")
	}
	return DueJob{
		Id:         model.Id(id),
		ModuleName: moduleName,
		JobKey:     jobKey,
		NextRunAt:  nextRunAt.UTC(),
	}, nil
}

func utcOrNil(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	utc := value.UTC()
	return &utc
}

func ptrId(value string) *model.Id {
	id := model.Id(value)
	return &id
}
