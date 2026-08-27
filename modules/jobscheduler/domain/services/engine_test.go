package services

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.bryk.io/pkg/errors"
)

// fakeClock drives the engine without waiting for real time, so a test can advance ten simulated
// minutes instantly and count what the engine did in them.
type fakeClock struct {
	mu     sync.Mutex
	now    time.Time
	timers []*fakeTimer
}

func newFakeClock() *fakeClock {
	return &fakeClock{now: time.Date(2026, time.August, 20, 10, 0, 0, 0, time.UTC)}
}

func (this *fakeClock) Now() time.Time {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.now
}

func (this *fakeClock) NewTimer(d time.Duration) Timer {
	this.mu.Lock()
	defer this.mu.Unlock()
	timer := &fakeTimer{clock: this, fireAt: this.now.Add(d), ch: make(chan time.Time, 1)}
	this.timers = append(this.timers, timer)
	return timer
}

// advance moves simulated time forward and fires whatever became due, letting the engine run
// between steps so its tick and timer reset actually happen.
func (this *fakeClock) advance(d time.Duration) {
	deadline := time.Now().Add(2 * time.Second)
	step := 500 * time.Millisecond
	for elapsed := time.Duration(0); elapsed < d; elapsed += step {
		this.mu.Lock()
		this.now = this.now.Add(step)
		due := []*fakeTimer{}
		for _, timer := range this.timers {
			if !timer.stopped && !timer.fired && !timer.fireAt.After(this.now) {
				timer.fired = true
				due = append(due, timer)
			}
		}
		this.mu.Unlock()

		for _, timer := range due {
			select {
			case timer.ch <- this.Now():
			default:
			}
		}
		// Give the engine goroutine a chance to run its tick and reset the timer.
		time.Sleep(time.Millisecond)
		if time.Now().After(deadline) {
			return
		}
	}
}

type fakeTimer struct {
	clock   *fakeClock
	fireAt  time.Time
	ch      chan time.Time
	stopped bool
	fired   bool
}

func (this *fakeTimer) Chan() <-chan time.Time { return this.ch }

func (this *fakeTimer) Stop() bool {
	this.clock.mu.Lock()
	defer this.clock.mu.Unlock()
	wasActive := !this.fired && !this.stopped
	this.stopped = true
	return wasActive
}

func (this *fakeTimer) Reset(d time.Duration) {
	this.clock.mu.Lock()
	defer this.clock.mu.Unlock()
	select {
	case <-this.ch:
	default:
	}
	this.fireAt = this.clock.now.Add(d)
	this.stopped = false
	this.fired = false
}

// waitForTicks blocks until the counter reaches want, or fails the test. Polling a counter is more
// robust here than a channel handshake, because the engine may legitimately tick more than once.
func waitForTicks(t *testing.T, counter *int32, want int32, within time.Duration) {
	t.Helper()
	require.Eventually(t, func() bool {
		return atomic.LoadInt32(counter) >= want
	}, within, 2*time.Millisecond, "expected at least %d ticks", want)
}

// AC-21: the scheduler does not poll the database every second.
//
// The proof is a count rather than an inspection of the code: over ten simulated minutes with no
// work at all, a per-second poller would tick ~600 times. With a 60s reconciliation interval the
// engine should tick on the order of ten times. The assertion is deliberately loose at the top -
// the point is the order of magnitude, not an exact number.
func TestEngineDoesNotPollEverySecond(t *testing.T) {
	clock := newFakeClock()
	var ticks int32

	engine := NewEngine(EngineParam{
		Clock:             clock,
		ReconcileInterval: 60 * time.Second,
		Tick: func(context.Context, WakeReason) (*time.Time, error) {
			atomic.AddInt32(&ticks, 1)
			return nil, nil // no work at all
		},
	})
	engine.Start()
	defer engine.Stop(newDeadlineCtx(t, 2*time.Second))

	clock.advance(10 * time.Minute)

	observed := atomic.LoadInt32(&ticks)
	assert.Less(t, observed, int32(60),
		"a per-second poller would tick ~600 times in ten minutes; this must be far fewer")
	assert.Greater(t, observed, int32(0), "the engine must still be running")
}

// The reconciliation interval is a ceiling, not a period: work due sooner is not made to wait for
// it. Without this, the interval would become the scheduler's real resolution.
func TestHorizonPrefersWorkDueSoonerThanTheReconcileInterval(t *testing.T) {
	clock := newFakeClock()
	engine := NewEngine(EngineParam{Clock: clock, ReconcileInterval: 60 * time.Second})

	soon := clock.Now().Add(5 * time.Second)
	assert.Equal(t, 5*time.Second, engine.horizon(&soon))

	distant := clock.Now().Add(10 * time.Minute)
	assert.Equal(t, 60*time.Second, engine.horizon(&distant),
		"work further out than the interval must not extend the sleep past it")

	assert.Equal(t, 60*time.Second, engine.horizon(nil),
		"with no known work the engine still reconciles on the interval")
}

// Overdue work must not spin the loop. A negative delay would fire the timer immediately, the tick
// would find the same overdue row, and the engine would busy-wait.
func TestHorizonFloorsOverdueWorkRatherThanSpinning(t *testing.T) {
	clock := newFakeClock()
	engine := NewEngine(EngineParam{Clock: clock, ReconcileInterval: 60 * time.Second})

	overdue := clock.Now().Add(-time.Hour)

	assert.Equal(t, minHorizon, engine.horizon(&overdue))
	assert.Greater(t, engine.horizon(&overdue), time.Duration(0))
}

// A wake-up must reach the engine promptly - that is the whole point of having one - and must
// carry the reason it was raised.
func TestWakeTriggersATickWithItsReason(t *testing.T) {
	clock := newFakeClock()
	reasons := make(chan WakeReason, 8)

	engine := NewEngine(EngineParam{
		Clock:             clock,
		ReconcileInterval: time.Hour, // long, so only the wake-up can cause a tick
		Tick: func(_ context.Context, reason WakeReason) (*time.Time, error) {
			select {
			case reasons <- reason:
			default:
			}
			return nil, nil
		},
	})
	engine.Start()
	defer engine.Stop(newDeadlineCtx(t, 2*time.Second))

	engine.Wake(WakeJobCreated)

	// The startup timer is set to minHorizon, and the fake clock is not advanced here, so the only
	// thing that can produce a tick is the wake-up itself.
	deadline := time.After(2 * time.Second)
	for {
		select {
		case reason := <-reasons:
			if reason == WakeJobCreated {
				return
			}
		case <-deadline:
			t.Fatal("a wake-up did not cause a tick carrying its reason")
		}
	}
}

// Wake must never block its caller. An HTTP handler that has just created a job should not be made
// to wait on a scheduler tick.
func TestWakeNeverBlocksEvenWhenNobodyIsListening(t *testing.T) {
	engine := NewEngine(EngineParam{Clock: newFakeClock()})

	done := make(chan struct{})
	go func() {
		for i := 0; i < 1000; i++ {
			engine.Wake(WakeJobUpdated)
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Wake blocked; it must coalesce rather than queue")
	}
}

// A failing tick must not stop the scheduler. One bad query would otherwise mean no job ever runs
// again, where the reconciliation timer would have retried it moments later.
func TestEngineKeepsRunningAfterAFailedTick(t *testing.T) {
	clock := newFakeClock()
	var ticks int32

	engine := NewEngine(EngineParam{
		Clock:             clock,
		ReconcileInterval: 60 * time.Second,
		Tick: func(context.Context, WakeReason) (*time.Time, error) {
			atomic.AddInt32(&ticks, 1)
			return nil, errors.New("database is unreachable")
		},
	})
	engine.Start()
	defer engine.Stop(newDeadlineCtx(t, 2*time.Second))

	clock.advance(5 * time.Minute)

	assert.Greater(t, atomic.LoadInt32(&ticks), int32(1),
		"the engine must keep ticking after an error, not give up")
}

// Stop must end the loop and be safe to call more than once, since both the module hook and a test
// cleanup may reach it.
func TestStopIsIdempotentAndEndsTheLoop(t *testing.T) {
	engine := NewEngine(EngineParam{
		Clock:             newFakeClock(),
		ReconcileInterval: time.Hour,
		Tick: func(context.Context, WakeReason) (*time.Time, error) {
			return nil, nil
		},
	})
	engine.Start()

	assert.True(t, engine.Stop(newDeadlineCtx(t, 2*time.Second)))
	assert.True(t, engine.Stop(newDeadlineCtx(t, 2*time.Second)),
		"stopping an already stopped engine must not hang or panic")
}

// Stop reports when a tick is still running, so the shutdown path can tell a clean stop from one
// that ran out of budget.
func TestStopReportsWhenATickOverrunsTheBudget(t *testing.T) {
	release := make(chan struct{})
	defer close(release)
	entered := make(chan struct{})
	var once sync.Once

	engine := NewEngine(EngineParam{
		Clock:             newFakeClock(),
		ReconcileInterval: time.Hour,
		Tick: func(context.Context, WakeReason) (*time.Time, error) {
			once.Do(func() { close(entered) })
			<-release
			return nil, nil
		},
	})
	engine.Start()

	// Get a tick genuinely stuck first, or Stop would find an idle loop and legitimately report a
	// clean shutdown - which would pass this test for entirely the wrong reason.
	engine.Wake(WakeJobCreated)
	select {
	case <-entered:
	case <-time.After(2 * time.Second):
		t.Fatal("the tick never started, so there is nothing to overrun")
	}

	expired, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	assert.False(t, engine.Stop(expired),
		"a tick still in flight means the drain did not complete")
}

// Start must be safe to call twice: two loops would double every claim attempt.
func TestStartOnlyEverRunsOneLoop(t *testing.T) {
	clock := newFakeClock()
	var ticks int32

	engine := NewEngine(EngineParam{
		Clock:             clock,
		ReconcileInterval: time.Hour,
		Tick: func(context.Context, WakeReason) (*time.Time, error) {
			atomic.AddInt32(&ticks, 1)
			return nil, nil
		},
	})
	engine.Start()
	engine.Start()
	engine.Start()
	defer engine.Stop(newDeadlineCtx(t, 2*time.Second))

	// Let the startup timer fire. Three loops would each run their own startup tick.
	clock.advance(time.Second)
	waitForTicks(t, &ticks, 1, 2*time.Second)

	assert.LessOrEqual(t, atomic.LoadInt32(&ticks), int32(2),
		"repeated Start calls must not spawn additional loops")
}
