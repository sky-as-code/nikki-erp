package services

import (
	stdErr "errors"
	"time"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"

	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/infra/repository"
)

// maxErrorMessageLen matches the error_message column. Truncating here rather than letting the
// insert fail means a verbose error still produces a recorded attempt: the history of what ran
// matters more than the last few hundred characters of a stack trace.
const maxErrorMessageLen = 4000

// reapExpiredLeases recovers work whose worker died holding it.
//
// An abandoned attempt goes through DecideRetry exactly like a failed one. That is the point of
// the design: there is one retry rule, and crash recovery is not an exception to it. An instance
// that died mid-attempt must not get unlimited retries, and must not bypass the retry window.
func (this *Runner) reapExpiredLeases(ctx corectx.Context, now time.Time) error {
	reaped, err := corecrud.ExecInTranx(ctx, this.attemptRepo, func(
		tranxCtx corectx.Context,
	) (*[]repository.ReapedAttempt, error) {
		attempts, err := repository.ReapExpiredLeases(
			tranxCtx, this.attemptRepo.GetBaseRepo(), now, this.cfg.ClaimBatchSize,
		)
		return &attempts, err
	})
	if err != nil {
		return err
	}

	var errs []error
	for _, attempt := range *reaped {
		if err := this.settleReapedAttempt(ctx, attempt, now); err != nil {
			// One unrecoverable attempt must not stop the others being recovered: each is
			// independent work belonging to a different job.
			errs = append(errs, err)
		}
	}
	return stdErr.Join(errs...)
}

// settleReapedAttempt decides what happens to the execution behind an abandoned attempt.
//
// The failure is treated as retryable: the instance died, which says nothing about whether the
// work itself would succeed. Marking it non-retryable would end an execution because a container
// was restarted, which is the opposite of what lease recovery is for.
func (this *Runner) settleReapedAttempt(
	ctx corectx.Context, attempt repository.ReapedAttempt, now time.Time,
) error {
	execution, err := this.loadExecution(ctx, attempt.ExecutionId)
	if err != nil || execution == nil {
		return err
	}

	snapshot, err := snapshotOf(*execution)
	if err != nil {
		return err
	}

	decision := DecideRetry(RetryInput{
		AttemptNumber:     int(attempt.AttemptNumber),
		MaxAttempts:       snapshot.EffectiveMaxAttempts,
		RetryIntervalSecs: snapshot.EffectiveRetryIntervalSeconds,
		MaxIntervalSecs:   this.cfg.ExpBackoffMaxIntervalSecs,
		FinishedAt:        now,
		NextOccurrenceAt:  goTimePtr(execution.GetNextOccurrenceAt()),
		Retryable:         true,
	})

	if !decision.ShouldRetry {
		return this.updateExecution(ctx, attempt.ExecutionId, dmodel.DynamicFields{
			models.ExecutionFieldStatus:      models.ExecutionStatusFailed,
			models.ExecutionFieldFinishedAt:  toModelDateTime(now),
			models.ExecutionFieldFailureCode: decision.FailureCode,
		})
	}
	return this.updateExecution(ctx, attempt.ExecutionId, dmodel.DynamicFields{
		models.ExecutionFieldStatus:      models.ExecutionStatusWaitingRetry,
		models.ExecutionFieldAvailableAt: toModelDateTime(decision.NextRetryAt),
	})
}

func (this *Runner) loadExecution(
	ctx corectx.Context, executionId model.Id,
) (*models.Execution, error) {
	found, err := this.executionRepo.GetOne(ctx, dyn.RepoGetOneParam{
		Filter: dmodel.DynamicFields{models.ExecutionFieldId: executionId},
	})
	if err != nil || found == nil || !found.HasData {
		return nil, err
	}
	return &found.Data, nil
}

// nextHorizon is the earliest instant at which there is known work.
//
// This one query is what makes the whole design possible: without it there would be no
// alternative to polling, because the engine would have no way to know how long it may sleep.
func (this *Runner) nextHorizon(ctx corectx.Context) (*time.Time, error) {
	return repository.EarliestPendingInstant(
		ctx, this.jobRepo.GetBaseRepo(), this.executionRepo.GetBaseRepo(),
	)
}

// pruneHistoryIfDue deletes execution history past its retention, at most once an hour.
//
// Retention is configured in days, so running it on every tick would issue a delete a minute
// that matches nothing for hours at a stretch. The batch is bounded by the claim batch size so
// that one tick cannot try to delete a million rows in a single transaction.
func (this *Runner) pruneHistoryIfDue(ctx corectx.Context, now time.Time) error {
	if !this.lastCleanupAt.IsZero() && now.Sub(this.lastCleanupAt) < historyCleanupInterval {
		return nil
	}
	this.lastCleanupAt = now

	cutoff := now.AddDate(0, 0, -this.cfg.HistoryRetentionDays)
	_, err := corecrud.ExecInTranx(ctx, this.executionRepo, func(
		tranxCtx corectx.Context,
	) (*int64, error) {
		// Only terminal executions are eligible. A running or waiting_retry execution is live
		// work whatever its age, and deleting one would silently drop a job that is mid-retry.
		deleted, err := repository.DeleteExpiredHistory(
			tranxCtx, this.executionRepo.GetBaseRepo(), cutoff, this.cfg.ClaimBatchSize,
		)
		return &deleted, err
	})
	return err
}

const historyCleanupInterval = time.Hour
