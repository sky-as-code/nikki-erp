package job

import (
	"context"
	"testing"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// The scoping seam. What is pinned here is the default and the swap, because a sweep registered
// against the wrong scoper either runs once where it should run per tenant — reading across tenants
// in a multi-tenant build — or panics on every tick, which is what kept the scheduler unstarted.

// The default runs a sweep exactly once. It is the nikkierp behaviour, and the one a build with no
// tenants needs: there is a single dataset, so one run covers it.
func TestTheDefaultScoperRunsASweepOnce(t *testing.T) {
	SetSweepScoper(singleScopeSweep)

	runs := 0
	handler := ScopeSweep(func(ctx corectx.Context) error {
		runs++
		return nil
	}, "test-sweep")

	if err := handler(context.Background(), nil); err != nil {
		t.Fatalf("the sweep returned %v", err)
	}
	if runs != 1 {
		t.Errorf("the sweep ran %d times, want exactly 1", runs)
	}
}

// The sweep is handed a usable context. A nil or empty one would fail at the first repository call,
// which is the whole class of failure this seam exists to prevent.
func TestTheSweepReceivesAContext(t *testing.T) {
	SetSweepScoper(singleScopeSweep)

	var seen corectx.Context
	handler := ScopeSweep(func(ctx corectx.Context) error {
		seen = ctx
		return nil
	}, "test-sweep")

	if err := handler(context.Background(), nil); err != nil {
		t.Fatalf("the sweep returned %v", err)
	}
	if seen == nil {
		t.Fatal("the sweep must be handed a request context, not nil")
	}
}

// A sweep's error reaches the scheduler rather than being swallowed. Handle logs it, and a silently
// dropped error is a sweep that appears to succeed while doing nothing.
func TestASweepErrorIsReturned(t *testing.T) {
	SetSweepScoper(singleScopeSweep)

	wanted := context.DeadlineExceeded
	handler := ScopeSweep(func(ctx corectx.Context) error {
		return wanted
	}, "test-sweep")

	if got := handler(context.Background(), nil); got != wanted {
		t.Errorf("the handler returned %v, want the sweep's own error", got)
	}
}

// A deployment can replace the strategy, which is how coremart installs its per-tenant fanout. The
// swap is what makes one sweep work in both builds without build tags.
func TestAScoperCanBeReplaced(t *testing.T) {
	defer SetSweepScoper(singleScopeSweep)

	fanoutRuns := 0
	SetSweepScoper(func(sweep TenantlessSweep, jobName string) JobHandleFn {
		return func(ctx context.Context, _ *string) error {
			// Stands in for "once per tenant": run it twice.
			for range 2 {
				if err := sweep(corectx.NewRequestContext(ctx)); err != nil {
					return err
				}
			}
			return nil
		}
	})

	handler := ScopeSweep(func(ctx corectx.Context) error {
		fanoutRuns++
		return nil
	}, "test-sweep")

	if err := handler(context.Background(), nil); err != nil {
		t.Fatalf("the sweep returned %v", err)
	}
	if fanoutRuns != 2 {
		t.Errorf("the replaced scoper ran the sweep %d times, want 2", fanoutRuns)
	}
}

// A nil scoper is ignored rather than installed. Installing one would make ScopeSweep panic at
// registration, taking down boot — and the caller that passed nil almost certainly meant "leave it".
func TestANilScoperIsIgnored(t *testing.T) {
	defer SetSweepScoper(singleScopeSweep)

	SetSweepScoper(singleScopeSweep)
	SetSweepScoper(nil)

	runs := 0
	handler := ScopeSweep(func(ctx corectx.Context) error {
		runs++
		return nil
	}, "test-sweep")

	if err := handler(context.Background(), nil); err != nil {
		t.Fatalf("the sweep returned %v", err)
	}
	if runs != 1 {
		t.Errorf("a nil scoper must leave the previous one in place; the sweep ran %d times", runs)
	}
}
