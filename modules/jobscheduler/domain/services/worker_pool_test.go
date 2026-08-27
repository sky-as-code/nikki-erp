package services

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AC-7: the pool never runs more than its configured size at once. The high-water mark is measured
// rather than inferred, because a bound that is merely intended is not a bound.
func TestPoolNeverExceedsItsConfiguredSize(t *testing.T) {
	const size = 4
	const tasks = size + 8

	pool := NewWorkerPool(size)
	var running, highWater int32
	release := make(chan struct{})
	var started sync.WaitGroup
	started.Add(tasks)

	for i := 0; i < tasks; i++ {
		go func() {
			pool.Submit(context.Background(), func() {
				current := atomic.AddInt32(&running, 1)
				for {
					mark := atomic.LoadInt32(&highWater)
					if current <= mark || atomic.CompareAndSwapInt32(&highWater, mark, current) {
						break
					}
				}
				started.Done()
				<-release
				atomic.AddInt32(&running, -1)
			})
		}()
	}

	// Let the first wave occupy every slot, then confirm nothing beyond the bound got in.
	assert.Eventually(t, func() bool {
		return atomic.LoadInt32(&running) == int32(size)
	}, 3*time.Second, 5*time.Millisecond, "the pool should fill to exactly its size")
	assert.LessOrEqual(t, atomic.LoadInt32(&highWater), int32(size),
		"more tasks ran at once than the pool allows")

	close(release)
	started.Wait()
	require.True(t, pool.Drain(newDeadlineCtx(t, 5*time.Second)))
	assert.LessOrEqual(t, highWater, int32(size))
}

// During shutdown the engine stops dispatching rather than queueing forever. The executions it
// does not start keep their leases and are recovered elsewhere, which is a normal path.
func TestSubmitReportsWhenTheContextEndsBeforeASlotFrees(t *testing.T) {
	pool := NewWorkerPool(1)
	release := make(chan struct{})
	defer close(release)

	require.True(t, pool.Submit(context.Background(), func() { <-release }))

	cancelled, cancel := context.WithCancel(context.Background())
	cancel()

	assert.False(t, pool.Submit(cancelled, func() { t.Fatal("must not run") }),
		"a cancelled context must refuse the submission rather than block")
}

// One job with a bug must not stop every other job from running, and must not shrink the pool by
// leaking the slot it was holding.
func TestPanickingTaskReleasesItsSlot(t *testing.T) {
	pool := NewWorkerPool(1)

	require.True(t, pool.Submit(context.Background(), func() {
		panic("a job handler blew up")
	}))
	require.True(t, pool.Drain(newDeadlineCtx(t, 5*time.Second)))

	ran := make(chan struct{})
	require.True(t, pool.Submit(context.Background(), func() { close(ran) }),
		"the slot must be reusable after a panic")

	select {
	case <-ran:
	case <-time.After(3 * time.Second):
		t.Fatal("the pool leaked the panicking task's slot")
	}
	require.True(t, pool.Drain(newDeadlineCtx(t, 5*time.Second)))
}

// A pool that admits nothing accepts work and never runs it, which is indistinguishable from a
// hung scheduler. Raising the floor to one keeps a misconfiguration slow rather than silent.
func TestPoolSizeIsAtLeastOne(t *testing.T) {
	for _, size := range []int{0, -5} {
		pool := NewWorkerPool(size)
		assert.Equal(t, 1, pool.Size())

		ran := make(chan struct{})
		require.True(t, pool.Submit(context.Background(), func() { close(ran) }))
		select {
		case <-ran:
		case <-time.After(3 * time.Second):
			t.Fatalf("a pool created with size %d never ran anything", size)
		}
	}
}

// Drain reports whether the pool emptied, so the shutdown path can tell a clean stop from one that
// ran out of budget with work still in flight.
func TestDrainReportsWhetherWorkFinishedWithinTheBudget(t *testing.T) {
	pool := NewWorkerPool(1)
	release := make(chan struct{})

	require.True(t, pool.Submit(context.Background(), func() { <-release }))

	expired, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	assert.False(t, pool.Drain(expired), "work still running means the drain did not complete")

	close(release)
	assert.True(t, pool.Drain(newDeadlineCtx(t, 5*time.Second)))
}

func newDeadlineCtx(t *testing.T, within time.Duration) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), within)
	t.Cleanup(cancel)
	return ctx
}
