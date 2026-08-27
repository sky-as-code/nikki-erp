package services

import (
	"context"
	"encoding/json"
	"time"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"

	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/infra/repository"
	itext "github.com/sky-as-code/nikki-erp/modules/jobscheduler/interfaces/external"
)

// claimAndDispatch takes a batch of due executions and hands each to the worker pool.
//
// The claim and the attempt insert share one transaction. Claiming, committing, and only then
// writing the attempt would reintroduce exactly the two-transaction window the atomic claim
// exists to close: between the two, the execution is running with no lease naming who holds it,
// and a reaper would see an execution in flight that nothing is responsible for.
func (this *Runner) claimAndDispatch(
	ctx context.Context, reqCtx corectx.Context, now time.Time,
) error {
	claimed, err := corecrud.ExecInTranx(reqCtx, this.executionRepo, func(
		tranxCtx corectx.Context,
	) (*[]claimedWork, error) {
		executions, err := repository.ClaimDueExecutions(
			tranxCtx, this.executionRepo.GetBaseRepo(), now, this.cfg.ClaimBatchSize,
		)
		if err != nil {
			return nil, err
		}

		work := make([]claimedWork, 0, len(executions))
		for _, execution := range executions {
			attempt, err := this.openAttempt(tranxCtx, execution, now)
			if err != nil {
				return nil, err
			}
			work = append(work, claimedWork{Execution: execution, Attempt: attempt})
		}
		return &work, nil
	})
	if err != nil {
		return err
	}

	for _, item := range *claimed {
		this.dispatch(ctx, item)
	}
	return nil
}

// claimedWork is one execution and the attempt row opened for it.
type claimedWork struct {
	Execution repository.ClaimedExecution
	Attempt   attemptRecord
}

type attemptRecord struct {
	Id     model.Id
	Number int32

	// StartedAt is carried in memory rather than re-read when the attempt closes. The duration
	// must measure the attempt, and a value read back from the row would be the instant the
	// database recorded rather than the instant the work began.
	StartedAt time.Time
}

// openAttempt writes the running attempt that carries the lease.
//
// The lease lives on the attempt rather than the execution because it identifies which instance
// is running this try. It expires at the attempt timeout plus a safety margin, so that a worker
// merely running slowly is not reaped out from under itself.
func (this *Runner) openAttempt(
	ctx corectx.Context, execution repository.ClaimedExecution, now time.Time,
) (attemptRecord, error) {
	number := execution.AttemptCount + 1

	attempt := models.NewAttempt()
	id, err := model.NewId()
	if err != nil {
		return attemptRecord{}, err
	}
	attempt.SetId(id)
	attempt.SetExecutionId(&execution.Id)
	attempt.SetAttemptNumber(&number)
	attempt.SetStatus(strPtrOf(models.AttemptStatusRunning))
	attempt.SetInstanceId(strPtrOf(InstanceId()))
	attempt.SetStartedAt(toModelDateTime(now))
	attempt.SetLeaseExpiresAt(toModelDateTime(now.Add(this.cfg.LeaseDuration())))
	// See the note in insertExecution: baserepo.Insert applies no type defaults.
	attempt.SetCreatedAt(toModelDateTime(now))

	if _, err := this.attemptRepo.Insert(ctx, *attempt); err != nil {
		return attemptRecord{}, err
	}

	if err := this.setExecutionAttemptCount(ctx, execution.Id, number); err != nil {
		return attemptRecord{}, err
	}
	return attemptRecord{Id: *id, Number: number, StartedAt: now}, nil
}

// dispatch submits one claimed execution to the pool.
//
// A refused submission is not a failure: the pool refuses only when the context is done, which
// means shutdown. The execution keeps its live lease and its running attempt, and whichever
// instance reaps the lease afterwards decides its fate through the ordinary retry rule - a
// delayed retry rather than lost work.
func (this *Runner) dispatch(ctx context.Context, item claimedWork) {
	submitted := this.pool.Submit(ctx, func() {
		this.runAttempt(ctx, item)
	})
	if !submitted && this.logger != nil {
		this.logger.Warnf(
			"jobscheduler: shutting down before dispatching execution %s; its lease will be reaped",
			string(item.Execution.Id),
		)
	}
}

// runAttempt executes the action and records what happened.
func (this *Runner) runAttempt(ctx context.Context, item claimedWork) {
	snapshot, err := decodeSnapshot(item.Execution.JobSnapshot)
	if err != nil {
		this.logError("jobscheduler: unreadable job snapshot", err)
		// A snapshot that will not decode cannot be retried into working: the configuration the
		// attempt needs is the thing that is broken. Failing it non-retryably ends the execution
		// now rather than spending its whole quota on the same unreadable bytes.
		this.finishAttempt(ctx, item, snapshot, itext.ActionOutcome{
			ErrorCode:    "SNAPSHOT_UNREADABLE",
			ErrorMessage: err.Error(),
		})
		return
	}

	executor := this.executors.ExecutorFor(snapshot.ActionType)
	if executor == nil {
		// The job was registered when an executor for this type existed. Its absence now is a
		// deployment change rather than a caller's mistake, so it is a failed attempt with a
		// named code rather than a panic on a worker goroutine.
		this.finishAttempt(ctx, item, snapshot, itext.ActionOutcome{
			ErrorCode:    "EXECUTOR_MISSING",
			ErrorMessage: "no executor is registered for action type " + snapshot.ActionType,
		})
		return
	}

	outcome := executor.Execute(ctx, itext.ActionInput{
		Config:        snapshot.ActionConfig,
		ExecutionKey:  item.Execution.ExecutionKey,
		JobId:         snapshot.JobId,
		ExecutionId:   string(item.Execution.Id),
		AttemptNumber: int(item.Attempt.Number),
		Timeout:       this.cfg.AttemptTimeout(),
	})
	this.finishAttempt(ctx, item, snapshot, outcome)
}

// finishAttempt closes the attempt and settles the execution.
//
// It deliberately does not use the cancelled request context for its writes: the attempt has
// finished and the result must be recorded even when the process is shutting down. Losing the
// outcome would leave a running attempt with an expiring lease, and the work would be repeated.
func (this *Runner) finishAttempt(
	ctx context.Context, item claimedWork, snapshot JobSnapshot, outcome itext.ActionOutcome,
) {
	writeCtx := this.newRequestContext(context.WithoutCancel(ctx))
	now := this.clock.Now()

	if err := this.closeAttemptRow(writeCtx, item, outcome, now); err != nil {
		this.logError("jobscheduler: failed to close attempt", err)
	}
	if err := this.settleExecution(writeCtx, item, snapshot, outcome, now); err != nil {
		this.logError("jobscheduler: failed to settle execution", err)
	}
}

func (this *Runner) closeAttemptRow(
	ctx corectx.Context, item claimedWork, outcome itext.ActionOutcome, now time.Time,
) error {
	status := models.AttemptStatusFailed
	if outcome.Succeeded {
		status = models.AttemptStatusSucceeded
	}

	fields := dmodel.DynamicFields{
		models.AttemptFieldId:         item.Attempt.Id,
		models.AttemptFieldStatus:     status,
		models.AttemptFieldFinishedAt: toModelDateTime(now),
		// Recorded on success as well as failure: how long a job normally takes is what an
		// operator needs in order to choose a sensible attempt timeout for it.
		models.AttemptFieldDurationMs: durationMsSince(item.Attempt.StartedAt, now),
	}
	if !outcome.Succeeded {
		fields[models.AttemptFieldErrorCode] = outcome.ErrorCode
		fields[models.AttemptFieldErrorMessage] = truncate(outcome.ErrorMessage, maxErrorMessageLen)
		if outcome.HttpStatusCode != nil {
			fields[models.AttemptFieldHttpStatusCode] = *outcome.HttpStatusCode
		}
	}

	_, err := corecrud.UpdateRegardless(ctx, corecrud.UpdateRegardlessParam{
		Action:       "close attempt",
		DbRepoGetter: this.attemptRepo,
		Data:         fields,
	})
	return err
}

// settleExecution applies the retry rule and writes the execution's new state.
func (this *Runner) settleExecution(
	ctx corectx.Context, item claimedWork, snapshot JobSnapshot,
	outcome itext.ActionOutcome, now time.Time,
) error {
	if outcome.Succeeded {
		return this.updateExecution(ctx, item.Execution.Id, dmodel.DynamicFields{
			models.ExecutionFieldStatus:     models.ExecutionStatusSucceeded,
			models.ExecutionFieldFinishedAt: toModelDateTime(now),
		})
	}

	decision := DecideRetry(RetryInput{
		AttemptNumber:     int(item.Attempt.Number),
		MaxAttempts:       snapshot.EffectiveMaxAttempts,
		RetryIntervalSecs: snapshot.EffectiveRetryIntervalSeconds,
		MaxIntervalSecs:   this.cfg.ExpBackoffMaxIntervalSecs,
		FinishedAt:        now,
		NextOccurrenceAt:  item.Execution.NextOccurrenceAt,
		Retryable:         outcome.Retryable,
	})

	if !decision.ShouldRetry {
		return this.updateExecution(ctx, item.Execution.Id, dmodel.DynamicFields{
			models.ExecutionFieldStatus:      models.ExecutionStatusFailed,
			models.ExecutionFieldFinishedAt:  toModelDateTime(now),
			models.ExecutionFieldFailureCode: decision.FailureCode,
		})
	}

	// available_at, not a sleep: the retry is a row the next claim finds, so it survives this
	// process dying and can be picked up by any instance rather than only this one.
	if err := this.updateExecution(ctx, item.Execution.Id, dmodel.DynamicFields{
		models.ExecutionFieldStatus:      models.ExecutionStatusWaitingRetry,
		models.ExecutionFieldAvailableAt: toModelDateTime(decision.NextRetryAt),
	}); err != nil {
		return err
	}
	return this.recordNextRetryAt(ctx, item.Attempt.Id, decision.NextRetryAt)
}

func (this *Runner) recordNextRetryAt(
	ctx corectx.Context, attemptId model.Id, nextRetryAt time.Time,
) error {
	_, err := corecrud.UpdateRegardless(ctx, corecrud.UpdateRegardlessParam{
		Action:       "record next retry",
		DbRepoGetter: this.attemptRepo,
		Data: dmodel.DynamicFields{
			models.AttemptFieldId:          attemptId,
			models.AttemptFieldNextRetryAt: toModelDateTime(nextRetryAt),
		},
	})
	return err
}

func (this *Runner) updateExecution(
	ctx corectx.Context, executionId model.Id, fields dmodel.DynamicFields,
) error {
	fields[models.ExecutionFieldId] = executionId
	_, err := corecrud.UpdateRegardless(ctx, corecrud.UpdateRegardlessParam{
		Action:       "update execution",
		DbRepoGetter: this.executionRepo,
		Data:         fields,
	})
	return err
}

func (this *Runner) setExecutionAttemptCount(
	ctx corectx.Context, executionId model.Id, count int32,
) error {
	return this.updateExecution(ctx, executionId, dmodel.DynamicFields{
		models.ExecutionFieldAttemptCount: count,
	})
}

func decodeSnapshot(raw []byte) (JobSnapshot, error) {
	var snapshot JobSnapshot
	if len(raw) == 0 {
		return snapshot, nil
	}
	err := json.Unmarshal(raw, &snapshot)
	return snapshot, err
}
