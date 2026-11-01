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

type fakeTranx struct{ callbacks []func() }

func (this *fakeTranx) Commit() error {
	for _, fn := range this.callbacks {
		fn()
	}
	this.callbacks = nil
	return nil
}
func (this *fakeTranx) Rollback() error       { this.callbacks = nil; return nil }
func (this *fakeTranx) OnCommitted(fn func()) { this.callbacks = append(this.callbacks, fn) }

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
// may still roll back: the drain waits for the commit, and reads without the finished transaction.
func TestDrainOutboxNowWaitsForTheCommitInsideATransaction(t *testing.T) {
	withNoDrain(t)

	calls := 0
	var got corectx.Context
	SetOutboxDrain(func(ctx corectx.Context) { calls++; got = ctx })

	tranx := &fakeTranx{}
	ctx := corectx.NewRequestContext(context.Background())
	ctx.SetDbTranx(tranx)

	DrainOutboxNow(ctx)
	assert.Zero(t, calls, "nothing is published before the commit")

	assert.NoError(t, tranx.Commit())
	assert.Equal(t, 1, calls)
	assert.Nil(t, got.GetDbTranx(), "the committed transaction cannot be read through any more")
}

func TestDrainOutboxNowIsDroppedOnRollback(t *testing.T) {
	withNoDrain(t)

	calls := 0
	SetOutboxDrain(func(corectx.Context) { calls++ })

	tranx := &fakeTranx{}
	ctx := corectx.NewRequestContext(context.Background())
	ctx.SetDbTranx(tranx)

	DrainOutboxNow(ctx)
	assert.NoError(t, tranx.Rollback())

	assert.Zero(t, calls, "a rolled-back write has no events to announce; the sweep stays the guarantee")
}

func TestSetOutboxDrainIgnoresNil(t *testing.T) {
	withNoDrain(t)

	calls := 0
	SetOutboxDrain(func(corectx.Context) { calls++ })
	SetOutboxDrain(nil)

	DrainOutboxNow(corectx.NewRequestContext(context.Background()))

	assert.Equal(t, 1, calls, "a nil installation must not silently disable an installed drain")
}
