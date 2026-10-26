package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	db "github.com/sky-as-code/nikki-erp/modules/core/database"
)

// The drain hook, pinned without a broker. The sweep is what guarantees delivery, so every case
// here is about the drain doing NOTHING harmful - a drain that fires at the wrong moment publishes
// an event for a write that may still roll back, which no consumer can take back.

// withNoDrain restores the package-level hook, which other tests in this package share.
func withNoDrain(t *testing.T) {
	t.Helper()

	previous := outboxDrain
	t.Cleanup(func() { outboxDrain = previous })
	outboxDrain = nil
}

type fakeTranx struct{}

func (this *fakeTranx) Commit() error   { return nil }
func (this *fakeTranx) Rollback() error { return nil }

var _ db.DbTransaction = (*fakeTranx)(nil)

func TestDrainOutboxNowIsANoOpWithNoDrainInstalled(t *testing.T) {
	withNoDrain(t)

	assert.NotPanics(t, func() {
		DrainOutboxNow(corectx.NewRequestContext(context.Background()))
	}, "a build that never installs a drain still works, only slower")
}

func TestDrainOutboxNowCallsTheInstalledDrain(t *testing.T) {
	withNoDrain(t)

	var got corectx.Context
	calls := 0
	SetOutboxDrain(func(ctx corectx.Context) {
		calls++
		got = ctx
	})

	ctx := corectx.NewRequestContext(context.Background())
	DrainOutboxNow(ctx)

	assert.Equal(t, 1, calls)
	assert.Same(t, ctx, got, "the drain is scoped by the caller's context, tenant included")
}

// Inside a transaction the rows are not committed, so publishing them would announce a write that
// may still roll back.
func TestDrainOutboxNowDoesNothingInsideATransaction(t *testing.T) {
	withNoDrain(t)

	calls := 0
	SetOutboxDrain(func(corectx.Context) { calls++ })

	ctx := corectx.NewRequestContext(context.Background())
	ctx.SetDbTranx(&fakeTranx{})

	DrainOutboxNow(ctx)

	assert.Zero(t, calls, "those events wait for the sweep, which is the guarantee anyway")
}

func TestSetOutboxDrainIgnoresNil(t *testing.T) {
	withNoDrain(t)

	calls := 0
	SetOutboxDrain(func(corectx.Context) { calls++ })
	SetOutboxDrain(nil)

	DrainOutboxNow(corectx.NewRequestContext(context.Background()))

	assert.Equal(t, 1, calls, "a nil installation must not silently disable an installed drain")
}
