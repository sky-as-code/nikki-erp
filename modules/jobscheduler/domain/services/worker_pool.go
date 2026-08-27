package services

import (
	"context"
	"sync"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
)

// WorkerPool bounds how many attempts one instance runs at a time.
//
// The bound is the point. An unbounded pool claims a batch, spawns a goroutine for each item, and
// then one slow endpoint holds every one of them open until their leases expire and the work is
// reaped as abandoned - which the retry window may well refuse to retry. The scheduler would have
// converted a slow dependency into lost occurrences.
//
// It is a counted semaphore rather than a fixed set of long-lived worker goroutines because the
// acquire is what makes the bound provable: a caller cannot start work without taking a slot, so
// there is no path that bypasses the limit.
type WorkerPool struct {
	slots chan struct{}
	wg    sync.WaitGroup
}

// NewWorkerPool returns a pool admitting at most `size` concurrent tasks. A size below one is
// raised to one: a pool that admits nothing would accept work and never run it, which looks
// exactly like a hung scheduler.
func NewWorkerPool(size int) *WorkerPool {
	if size < 1 {
		size = 1
	}
	return &WorkerPool{slots: make(chan struct{}, size)}
}

// Size is the maximum number of tasks that may run at once.
func (this *WorkerPool) Size() int {
	return cap(this.slots)
}

// Submit blocks until a slot is free or ctx is done, then runs fn on its own goroutine.
//
// It returns false when ctx was cancelled before a slot came free. That is how the engine learns
// to stop dispatching the rest of a claimed batch during shutdown: the executions it does not
// start keep a live lease and are recovered by whichever instance reaps them, which is a normal
// path rather than an error.
//
// A panic inside fn releases the slot and is logged rather than taking down the process. One job
// with a bug must not stop every other job from running.
func (this *WorkerPool) Submit(ctx context.Context, fn func()) bool {
	select {
	case this.slots <- struct{}{}:
	case <-ctx.Done():
		return false
	}

	this.wg.Add(1)
	go func() {
		defer func() {
			// Recover before releasing, so a panicking task cannot leak its slot and shrink the
			// pool by one every time it happens.
			_ = ft.RecoverPanicFailedTo(recover(), "run a scheduled job attempt")
			<-this.slots
			this.wg.Done()
		}()
		fn()
	}()
	return true
}

// Drain waits for the tasks already running to finish, or until ctx expires.
//
// It reports whether the pool emptied. A false return means work was still in flight when the
// shutdown budget ran out; those attempts keep their leases and are reaped by another instance,
// so the outcome is a delayed retry rather than lost work.
func (this *WorkerPool) Drain(ctx context.Context) bool {
	done := make(chan struct{})
	go func() {
		this.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return true
	case <-ctx.Done():
		return false
	}
}
