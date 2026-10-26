package context

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type raceKey struct{}

// database/sql watches every query's context from its own goroutine (Rows.awaitDone) while the
// request goroutine keeps memoizing values with WithValue. Before core-001 was fixed, the inner
// context was a plain embedded field and this test failed under `go test -race`.
func TestContextMethodsAreSafeWhileTheOwnerSetsValues(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx := NewRequestContext(parent)

	stop := make(chan struct{})
	var readers sync.WaitGroup
	for i := 0; i < 4; i++ {
		readers.Add(1)
		go func() {
			defer readers.Done()
			for {
				select {
				case <-stop:
					return
				case <-ctx.Done():
				default:
					_ = ctx.Err()
					_, _ = ctx.Deadline()
					_ = ctx.Value(raceKey{})
				}
			}
		}()
	}
	for i := 0; i < 2000; i++ {
		ctx.WithValue(raceKey{}, i)
		ctx.SetDomainConstraints(nil)
	}
	close(stop)
	readers.Wait()

	assert.Equal(t, 1999, ctx.Value(raceKey{}))
	assert.NoError(t, ctx.Err())
	cancel()
	assert.ErrorIs(t, ctx.Err(), context.Canceled)
}

func TestCloneWithInnerDoesNotTouchTheOriginal(t *testing.T) {
	original := NewRequestContextM(context.Background(), "inventory")
	original.WithValue(raceKey{}, "original")
	inner, cancel := context.WithCancel(original.InnerContext())

	clone := CloneWithInner(original, inner)
	cancel()

	assert.ErrorIs(t, clone.Err(), context.Canceled)
	assert.NoError(t, original.Err())
	assert.Equal(t, "original", clone.Value(raceKey{}))
	assert.Equal(t, "inventory", clone.GetModuleName())
}

// A zero RequestContext has no inner context; it must behave as Background, not dereference nil.
func TestZeroRequestContextBehavesAsBackground(t *testing.T) {
	var ctx RequestContext

	assert.Nil(t, ctx.Done())
	assert.NoError(t, ctx.Err())
	assert.Nil(t, ctx.Value(raceKey{}))
}
