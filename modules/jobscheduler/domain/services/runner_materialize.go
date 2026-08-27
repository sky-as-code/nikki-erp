package services

import (
	"time"

	stdErr "errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"

	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/infra/repository"
)

// materializeDueJobs turns every job whose occurrence has arrived into an execution row.
//
// The whole phase runs in one transaction because LockDueJobs takes row locks with SKIP LOCKED,
// and on a pooled connection those locks would be released the instant the statement returned -
// at which point two instances would both see the same job as due and both materialize it.
func (this *Runner) materializeDueJobs(ctx corectx.Context, now time.Time) error {
	_, err := corecrud.ExecInTranx(ctx, this.jobRepo, func(
		tranxCtx corectx.Context,
	) (*int, error) {
		due, err := repository.LockDueJobs(
			tranxCtx, this.jobRepo.GetBaseRepo(), now, this.cfg.ClaimBatchSize,
		)
		if err != nil {
			return nil, err
		}

		materialized := 0
		var errs []error
		for _, job := range due {
			created, err := this.materializeOne(tranxCtx, job.Id, now)
			if err != nil {
				// One job failing must not abandon the others already locked in this
				// transaction: they would stay locked until it commits either way, and rolling
				// back would simply re-lock them on the next tick to fail again.
				errs = append(errs, err)
				continue
			}
			if created {
				materialized++
			}
		}
		return &materialized, stdErr.Join(errs...)
	})
	return err
}

// materializeOne decides and applies what one due job should do.
//
// The read of the job, the open-execution check, the execution insert and the next_run_at update
// all happen inside the caller's transaction. Splitting them would let two ticks both observe
// "no open execution" and both insert; the execution_key unique index is the backstop that
// catches what the transaction does not.
func (this *Runner) materializeOne(
	ctx corectx.Context, jobId model.Id, now time.Time,
) (bool, error) {
	job, err := this.loadJob(ctx, jobId)
	if err != nil || job == nil {
		return false, err
	}

	hasOpen, err := this.hasOpenExecution(ctx, jobId)
	if err != nil {
		return false, err
	}

	snapshot := BuildJobSnapshot(*job, this.cfg)
	result := Materialize(this.materializeInputFor(job, snapshot, hasOpen, now))

	if result.ShouldCreateExecution() {
		if err := this.insertExecution(ctx, *job, snapshot, result, now); err != nil {
			return false, err
		}
	}
	return result.ShouldCreateExecution(), this.advanceNextRunAt(ctx, *job, result.NextRunAt)
}

func (this *Runner) materializeInputFor(
	job *models.Job, snapshot JobSnapshot, hasOpen bool, now time.Time,
) MaterializeInput {
	in := MaterializeInput{
		IsEnabled:         true,
		MisfirePolicy:     snapshot.MisfirePolicy,
		ConcurrencyPolicy: snapshot.ConcurrencyPolicy,
		HasOpenExecution:  hasOpen,
		MisfireThreshold:  this.cfg.MisfireThreshold(),
		Now:               now,
	}
	if expr := job.GetCronExpression(); expr != nil {
		in.CronExpression = *expr
	}
	if enabled := job.GetIsEnabled(); enabled != nil {
		in.IsEnabled = *enabled
	}
	in.NextRunAt = goTimePtr(job.GetNextRunAt())
	in.EffectiveFrom = goTimePtr(job.GetEffectiveFrom())
	in.EffectiveUntil = goTimePtr(job.GetEffectiveUntil())
	return in
}

// insertExecution writes the queued execution.
//
// A duplicate execution_key is not an error. It means another instance materialized the same
// occurrence first, which is the unique index doing exactly what it is there for; the occurrence
// exists once, which is all that was wanted. That case is headed off before the insert rather
// than caught after: a unique-constraint violation aborts the whole enclosing Postgres
// transaction, not just this statement, which took down every other job's advanceNextRunAt in
// the same materializeDueJobs batch along with it.
func (this *Runner) insertExecution(
	ctx corectx.Context, job models.Job, snapshot JobSnapshot,
	result MaterializeResult, now time.Time,
) error {
	executionKey := BuildExecutionKey(
		snapshot.ModuleName, snapshot.JobKey, *result.ExecutionScheduledFor,
	)
	exists, err := this.executionKeyExists(ctx, executionKey)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	frozen, err := snapshotAsMap(snapshot)
	if err != nil {
		return err
	}

	execution := models.NewExecution()
	id, err := model.NewId()
	if err != nil {
		return err
	}
	execution.SetId(id)
	execution.SetJobId(job.GetId())
	execution.SetExecutionKey(strPtrOf(executionKey))
	execution.SetScheduledFor(toModelDateTime(*result.ExecutionScheduledFor))
	execution.SetNextOccurrenceAt(modelDateTimePtr(result.NextOccurrenceAt))
	execution.SetStatus(strPtrOf(models.ExecutionStatusQueued))
	// available_at is the scheduled instant, not now: a misfired occurrence must be claimable
	// immediately rather than waiting for a time that has already passed to come round again.
	execution.SetAvailableAt(toModelDateTime(maxTime(*result.ExecutionScheduledFor, now)))
	execution.SetAttemptCount(int32PtrOf(0))
	execution.SetJobSnapshot(frozen)
	// Set explicitly because these rows are written through baserepo.Insert rather than
	// corecrud.Create, and only the latter runs the schema validation that applies a field's
	// type default. The schemas also decline core.basemodel.base_model to stay out of tenant
	// scope, so there is no mixin filling this in either.
	execution.SetCreatedAt(toModelDateTime(now))

	_, err = this.executionRepo.Insert(ctx, *execution)
	if isDuplicateKey(err) {
		return nil
	}
	return err
}

// advanceNextRunAt moves the job on, whether or not an execution was produced.
//
// A skipped occurrence still advances. That is what stops an occurrence that was skipped for
// overlap, or for falling outside the effective window, from being rediscovered as due on every
// subsequent tick forever.
func (this *Runner) advanceNextRunAt(
	ctx corectx.Context, job models.Job, nextRunAt *time.Time,
) error {
	fields := dmodel.DynamicFields{
		models.JobFieldId:        job.GetId(),
		models.JobFieldNextRunAt: modelDateTimePtr(nextRunAt),
	}
	// The job schema carries versioned_model, so checkExistenceAndEtag compares this against the
	// stored row on every call. Must go through SetEtag rather than a raw map assignment: it
	// unwraps the *model.Etag job.GetEtag() returns into the plain model.Etag (string) the DB
	// value is compared against with ==, and a stored pointer would never equal that string,
	// tripping "etag mismatch" on every call regardless of the actual value. Omitting the field
	// entirely has the same failure mode via a nil comparison. Either way the update is rejected
	// as a client error, discarded below as if it were a plain error-free no-op, and next_run_at
	// never actually moves - which is what turned into an unbounded 100ms busy loop: the same job
	// stayed "due" forever because nothing ever advanced it past the tick that first found it.
	fields.SetEtag(basemodel.FieldEtag, job.GetEtag())
	result, err := corecrud.UpdateRegardless(ctx, corecrud.UpdateRegardlessParam{
		Action:       "advance job next run",
		DbRepoGetter: this.jobRepo,
		Data:         fields,
	})
	if err != nil {
		return err
	}
	if result != nil && result.ClientErrors.Count() > 0 {
		return result.ClientErrors.ToError()
	}
	return nil
}

func (this *Runner) loadJob(ctx corectx.Context, jobId model.Id) (*models.Job, error) {
	found, err := this.jobRepo.GetOne(ctx, dyn.RepoGetOneParam{
		Filter: dmodel.DynamicFields{models.JobFieldId: jobId},
	})
	if err != nil || found == nil || !found.HasData {
		return nil, err
	}
	return &found.Data, nil
}

// hasOpenExecution reports whether the job already has an execution that has not finished.
//
// waiting_retry counts as open. Its retry chain still belongs to the previous occurrence, and
// starting a new occurrence beside it would put two runs of the same job in flight - which is
// exactly what forbid_overlap exists to prevent.
func (this *Runner) hasOpenExecution(ctx corectx.Context, jobId model.Id) (bool, error) {
	graph := dmodel.NewSearchGraph().And(
		*dmodel.NewSearchGraph().
			NewCondition(models.ExecutionFieldJobId, dmodel.Equals, jobId).ToSearchNode(),
		*dmodel.NewSearchGraph().NewCondition(
			models.ExecutionFieldStatus, dmodel.In,
			models.ExecutionStatusQueued,
			models.ExecutionStatusRunning,
			models.ExecutionStatusWaitingRetry,
		).ToSearchNode(),
	)

	found, err := this.executionRepo.Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  1,
	})
	if err != nil || found == nil || !found.HasData {
		return false, err
	}
	return len(found.Data.Items) > 0, nil
}

// executionKeyExists reports whether an execution for this occurrence was already materialized.
//
// insertExecution used to rely on the unique index to reject the duplicate and treat that as a
// no-op. On Postgres a unique-constraint violation aborts the whole enclosing transaction - not
// just the one statement - so every later statement in the same materializeDueJobs transaction
// (including this job's own advanceNextRunAt) then failed with "current transaction is aborted".
// Checking first avoids the insert - and the abort - ever happening for the row that would have
// collided.
func (this *Runner) executionKeyExists(ctx corectx.Context, executionKey string) (bool, error) {
	graph := dmodel.NewSearchGraph().
		NewCondition(models.ExecutionFieldExecutionKey, dmodel.Equals, executionKey)

	found, err := this.executionRepo.Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  1,
	})
	if err != nil || found == nil || !found.HasData {
		return false, err
	}
	return len(found.Data.Items) > 0, nil
}
