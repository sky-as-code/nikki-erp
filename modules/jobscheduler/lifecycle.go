package jobscheduler

import (
	"context"
	"sync"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"

	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/domain/services"
	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/infra/external"
	itexec "github.com/sky-as-code/nikki-erp/modules/jobscheduler/interfaces/execution"
	itjob "github.com/sky-as-code/nikki-erp/modules/jobscheduler/interfaces/job"
)

// The loader discovers these two hooks by type assertion, so a signature that drifts is not a
// compile error there - it is a scheduler that silently never starts. Asserting here makes the
// mismatch a build failure instead.
var (
	_ modules.InCodeModuleAppStarted  = (*JobSchedulerModule)(nil)
	_ modules.InCodeModuleAppStopping = (*JobSchedulerModule)(nil)
)

// running holds what OnAppStopping needs to shut down what OnAppStarted began.
//
// It is package state rather than a field on the module because the module singleton is a
// stateless value shared by both binaries, and the loader hands back the same pointer to
// everyone. A mutex guards it because start and stop are called from different goroutines.
var running struct {
	sync.Mutex
	engine *services.Engine
	pool   *services.WorkerPool
	waker  *services.DeferredWaker
}

// OnAppStarted builds and starts the scheduling engine.
//
// It runs here rather than in Init because from its first tick the engine claims work and writes
// execution rows. Doing that against a container that is still being built would fail on the
// first tick, and doing it before the HTTP server is up would mean a job's REST action could
// fire at an endpoint this very process has not yet begun serving.
func (*JobSchedulerModule) OnAppStarted() error {
	return deps.Invoke(func(
		cfg services.SchedulerConfig,
		jobRepo itjob.JobRepository,
		executionRepo itexec.ExecutionRepository,
		attemptRepo itexec.AttemptRepository,
		dispatcher *external.ActionDispatcher,
		waker *services.DeferredWaker,
		logger logging.LoggerService,
	) error {
		if cfg.IsClaimBatchOversized() {
			// Not fatal, but worth saying once at boot: a batch far larger than the pool means
			// each tick claims work it cannot start, and those executions hold leases while they
			// wait for a slot.
			logger.Warnf(
				"jobscheduler: claim batch size %d is large for a worker pool of %d",
				cfg.ClaimBatchSize, cfg.WorkerConcurrency,
			)
		}

		pool := services.NewWorkerPool(cfg.WorkerConcurrency)
		runner := services.NewRunner(services.RunnerParam{
			Config:        cfg,
			JobRepo:       jobRepo,
			ExecutionRepo: executionRepo,
			AttemptRepo:   attemptRepo,
			Executors:     dispatcher,
			Pool:          pool,
			Logger:        logger,
		})
		engine := services.NewEngine(services.EngineParam{
			Logger:            logger,
			Tick:              runner.Tick,
			ReconcileInterval: cfg.ReconciliationInterval,
		})

		running.Lock()
		running.engine, running.pool, running.waker = engine, pool, waker
		running.Unlock()

		// Attaching before Start means a job created between the two is not missed: the wake is
		// delivered to an engine that has not ticked yet, and its first tick reads everything.
		waker.Attach(engine)
		engine.Start()

		logger.Infof(
			"jobscheduler: engine started as instance %s with %d workers, reconciling every %s",
			services.InstanceId(), cfg.WorkerConcurrency, cfg.ReconciliationInterval,
		)
		return nil
	})
}

// OnAppStopping stops the loop and waits for in-flight attempts to finish.
//
// The order matters. Detaching first stops a request still in flight from waking an engine that
// is about to drain; stopping the loop prevents a new batch being claimed; draining the pool
// then lets the attempts already running record their outcomes.
//
// Work still running when the deadline passes is not lost. Its attempt rows keep live leases,
// and whichever instance reaps them applies the ordinary retry rule - a delayed retry rather
// than a silent disappearance.
func (*JobSchedulerModule) OnAppStopping(ctx context.Context) error {
	running.Lock()
	engine, pool, waker := running.engine, running.pool, running.waker
	running.engine, running.pool, running.waker = nil, nil, nil
	running.Unlock()

	if waker != nil {
		waker.Detach()
	}
	if engine == nil {
		return nil
	}

	engine.Stop(ctx)
	if pool != nil {
		pool.Drain(ctx)
	}
	return nil
}
