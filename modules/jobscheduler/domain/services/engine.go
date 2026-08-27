package services

import (
	"context"
	"sync"
	"time"

	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

// WakeReason names why the engine was woken.
//
// Every reason is named rather than folded into one signal so that a lost wake-up shows in the log
// as a specific missing reason instead of as a job that was merely late - which is otherwise very
// hard to tell from a job that was scheduled late on purpose.
type WakeReason string

const (
	WakeJobCreated     WakeReason = "job_created"
	WakeJobUpdated     WakeReason = "job_updated"
	WakeJobEnabled     WakeReason = "job_enabled"
	WakeJobDeleted     WakeReason = "job_deleted"
	WakeRetryScheduled WakeReason = "retry_scheduled"
	WakeExecutionDone  WakeReason = "execution_finished"
	WakeLeaseRecovery  WakeReason = "lease_recovery"
	WakeNotification   WakeReason = "notification"
	WakeReconciliation WakeReason = "reconciliation"
	WakeStartup        WakeReason = "startup"
)

// minHorizon keeps the loop from spinning when work is already overdue.
//
// Without a floor, a row whose instant has passed yields a zero or negative delay, the timer fires
// immediately, the tick finds the same row still overdue, and the loop becomes a busy wait. A
// short floor turns that into a rapid but bounded retry.
const minHorizon = 100 * time.Millisecond

// TickFunc does one round of scheduler work: recover expired leases, materialize due jobs, claim
// and dispatch. It returns the next instant the engine has work for, or nil when it has none.
//
// The engine owns the timing; this owns the database. Splitting them is what lets the loop be
// tested against a fake clock and a counting tick, without a database anywhere near it.
type TickFunc func(ctx context.Context, reason WakeReason) (*time.Time, error)

// Engine is the scheduling loop.
//
// It is deliberately not a poller. The design is: persist next_run_at, sleep on a single timer
// until the earliest of it and the reconciliation interval, wake, claim atomically, and repeat.
// Querying every second would be simpler and is what the requirement forbids - it puts load
// proportional to time rather than to work, and it still does not react promptly to a job created
// a moment after a poll.
type Engine struct {
	clock             Clock
	logger            logging.LoggerService
	tick              TickFunc
	reconcileInterval time.Duration

	// wakeChan is capacity 1 and written non-blockingly. A wake-up is a hint that the horizon may
	// have moved, not a message to be delivered: two arriving before the loop runs are
	// indistinguishable from one, because the tick recomputes the horizon from the database
	// either way. Blocking an HTTP handler that just created a job on a scheduler tick would be
	// strictly worse than coalescing.
	wakeChan chan WakeReason

	startOnce sync.Once
	stopOnce  sync.Once
	done      chan struct{}
	finished  chan struct{}
}

type EngineParam struct {
	Clock             Clock
	Logger            logging.LoggerService
	Tick              TickFunc
	ReconcileInterval time.Duration
}

func NewEngine(param EngineParam) *Engine {
	clock := param.Clock
	if clock == nil {
		clock = RealClock{}
	}
	interval := param.ReconcileInterval
	if interval <= 0 {
		interval = time.Minute
	}
	return &Engine{
		clock:             clock,
		logger:            param.Logger,
		tick:              param.Tick,
		reconcileInterval: interval,
		wakeChan:          make(chan WakeReason, 1),
		done:              make(chan struct{}),
		finished:          make(chan struct{}),
	}
}

// Wake asks the engine to run a tick soon.
//
// It never blocks and never fails. A caller reporting that something changed should not have to
// care whether the scheduler is mid-tick, and must not be delayed by it.
func (this *Engine) Wake(reason WakeReason) {
	select {
	case this.wakeChan <- reason:
	default:
		// A wake-up is already pending and will pick this change up too, because the tick reads
		// the current state rather than a queued description of it.
	}
}

// Start runs the loop until Stop is called. It returns immediately.
func (this *Engine) Start() {
	this.startOnce.Do(func() {
		go this.run()
	})
}

// Stop ends the loop and waits for the tick in flight, or until ctx expires.
//
// It reports whether the loop finished within the budget. False means a tick was still running,
// in which case its claimed work keeps its leases and is recovered by whichever instance reaps it
// - a delayed retry rather than lost work.
func (this *Engine) Stop(ctx context.Context) bool {
	this.stopOnce.Do(func() {
		close(this.done)
	})

	select {
	case <-this.finished:
		return true
	case <-ctx.Done():
		return false
	}
}

func (this *Engine) run() {
	defer close(this.finished)

	// The first tick is immediate rather than one interval away: on boot there may be work that
	// fell due while the process was down, and waiting a full interval to look would make every
	// restart cost an interval of latency.
	timer := this.clock.NewTimer(minHorizon)
	defer timer.Stop()

	for {
		select {
		case <-this.done:
			return
		case reason := <-this.wakeChan:
			this.runTick(reason, timer)
		case <-timer.Chan():
			this.runTick(WakeReconciliation, timer)
		}
	}
}

func (this *Engine) runTick(reason WakeReason, timer Timer) {
	ctx, cancel := this.tickContext()
	defer cancel()

	next, err := this.tick(ctx, reason)
	if err != nil {
		// A failed tick is logged and the loop continues. The alternative - stopping - would turn
		// one bad query into a scheduler that never runs again, and the reconciliation timer means
		// the next tick retries whatever this one could not do.
		if this.logger != nil {
			this.logger.Error("scheduler tick failed", err)
		}
	}

	timer.Reset(this.horizon(next))
}

// tickContext bounds a tick to the reconciliation interval and cancels it when the engine stops.
//
// The bound matters because a tick that hangs on the database would otherwise hold the loop
// forever, and the timer that would have woken it is only reset once the tick returns.
func (this *Engine) tickContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), this.reconcileInterval)
	go func() {
		select {
		case <-this.done:
			cancel()
		case <-ctx.Done():
		}
	}()
	return ctx, cancel
}

// horizon is how long to sleep before the next tick: until the next known work, but never longer
// than the reconciliation interval and never shorter than the floor.
//
// The reconciliation interval is a ceiling, not a period. It guarantees the loop runs at least
// that often even if every wake-up is lost and every timer is wrong, and it never delays work that
// is due sooner. Raising it does not make jobs less punctual; it only widens the window in which a
// lost notification goes unnoticed.
func (this *Engine) horizon(next *time.Time) time.Duration {
	delay := this.reconcileInterval
	if next != nil {
		until := next.UTC().Sub(this.clock.Now())
		if until < delay {
			delay = until
		}
	}
	if delay < minHorizon {
		delay = minHorizon
	}
	return delay
}
