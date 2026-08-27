// Package services holds the scheduling engine and the rules it applies.
//
// This file is where the pure pieces meet the database. Materialize, DecideRetry and
// BuildJobSnapshot are all functions of their inputs and are tested as such; the Runner is what
// reads the rows they decide about and writes back what they decided. Keeping the join in one
// place is what lets those three stay free of persistence.
package services

import (
	"context"
	"time"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"

	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/constants"

	itexec "github.com/sky-as-code/nikki-erp/modules/jobscheduler/interfaces/execution"
	itext "github.com/sky-as-code/nikki-erp/modules/jobscheduler/interfaces/external"
	itjob "github.com/sky-as-code/nikki-erp/modules/jobscheduler/interfaces/job"
)

// ExecutorLookup answers which executor runs a given action type.
//
// It is an interface rather than the concrete dispatcher so that the domain layer does not
// import infra, and so a test can drive the runner with a stub that never leaves the process.
type ExecutorLookup interface {
	ExecutorFor(actionType string) itext.ActionExecutor
}

// RunnerParam is what the runner needs to do a tick.
type RunnerParam struct {
	Config        SchedulerConfig
	JobRepo       itjob.JobRepository
	ExecutionRepo itexec.ExecutionRepository
	AttemptRepo   itexec.AttemptRepository
	Executors     ExecutorLookup
	Pool          *WorkerPool
	Logger        logging.LoggerService
	Clock         Clock
}

// Runner performs one tick: recover, materialize, claim, dispatch, then clean up.
type Runner struct {
	cfg           SchedulerConfig
	jobRepo       itjob.JobRepository
	executionRepo itexec.ExecutionRepository
	attemptRepo   itexec.AttemptRepository
	executors     ExecutorLookup
	pool          *WorkerPool
	logger        logging.LoggerService
	clock         Clock

	// lastCleanupAt throttles history deletion. Retention is measured in days, so running it on
	// every tick would issue a delete a minute that matches nothing for hours at a time.
	lastCleanupAt time.Time
}

func NewRunner(param RunnerParam) *Runner {
	clock := param.Clock
	if clock == nil {
		clock = RealClock{}
	}
	pool := param.Pool
	if pool == nil {
		pool = NewWorkerPool(param.Config.WorkerConcurrency)
	}
	return &Runner{
		cfg:           param.Config,
		jobRepo:       param.JobRepo,
		executionRepo: param.ExecutionRepo,
		attemptRepo:   param.AttemptRepo,
		executors:     param.Executors,
		pool:          pool,
		logger:        param.Logger,
		clock:         clock,
	}
}

// Tick is the engine's TickFunc.
//
// The order of the four phases is the design and must not be rearranged:
//
//  1. Reap expired leases first, so that work recovered from a died instance is claimable in
//     this same tick rather than waiting for the next one.
//  2. Materialize due jobs, turning schedules into execution rows.
//  3. Claim a batch atomically and dispatch it to the bounded pool.
//  4. Recompute the horizon, and occasionally prune history.
//
// It returns the earliest instant at which there is known work, which is what the engine sets
// its single timer from. Nil means nothing is pending and the engine falls back to the
// reconciliation interval.
//
// An error in one phase does not abort the others. A tick that failed to materialize should
// still dispatch what is already queued: the alternative is one bad job row stopping the whole
// scheduler, which is precisely the failure the phases are separated to avoid.
func (this *Runner) Tick(ctx context.Context, reason WakeReason) (*time.Time, error) {
	reqCtx := this.newRequestContext(ctx)
	now := this.clock.Now()

	var errs []error
	if err := this.reapExpiredLeases(reqCtx, now); err != nil {
		errs = append(errs, err)
		this.logError("jobscheduler: lease recovery failed", err)
	}
	if err := this.materializeDueJobs(reqCtx, now); err != nil {
		errs = append(errs, err)
		this.logError("jobscheduler: materialization failed", err)
	}
	if err := this.claimAndDispatch(ctx, reqCtx, now); err != nil {
		errs = append(errs, err)
		this.logError("jobscheduler: claim and dispatch failed", err)
	}
	if err := this.pruneHistoryIfDue(reqCtx, now); err != nil {
		// Retention failing is the least urgent of the four: it wastes disk, it does not stop
		// work running. It is logged and not returned for that reason.
		this.logError("jobscheduler: history cleanup failed", err)
	}

	horizon, err := this.nextHorizon(reqCtx)
	if err != nil {
		errs = append(errs, err)
		this.logError("jobscheduler: horizon lookup failed", err)
	}
	return horizon, joinTickErrors(errs)
}

func (this *Runner) logError(message string, err error) {
	if this.logger == nil || err == nil {
		return
	}
	this.logger.Error(message, err)
}

// newRequestContext builds the ambient context the repositories expect.
//
// The engine runs on its own goroutine with no HTTP request behind it, so there is nothing to
// inherit. The context carries no tenant, which is correct and load-bearing: the scheduler's
// three tables are global, and a tenant-scoped context here would make every write assert a
// tenant that does not exist.
func (this *Runner) newRequestContext(ctx context.Context) corectx.Context {
	return corectx.NewRequestContextM(ctx, constants.JobSchedulerModuleName)
}
