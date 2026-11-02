package dynamicresource

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/audit"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

// tranxRecorder behaves like the real sink: an entry belongs to the transaction it was recorded
// in and is kept only when that transaction commits.
type tranxRecorder struct {
	fail      error
	committed []audit.Entry
}

func (this *tranxRecorder) Enabled() bool { return true }

func (this *tranxRecorder) Record(ctx corectx.Context, entries []audit.Entry) error {
	if this.fail != nil {
		return this.fail
	}
	tranx := ctx.GetDbTranx()
	if tranx == nil {
		return errors.New("an entry was recorded outside any transaction")
	}
	tranx.OnCommitted(func() { this.committed = append(this.committed, entries...) })
	return nil
}

func insertOne(writes *int) func(corectx.Context, dmodel.DynamicFields) (*dyn.OpResult[int], error) {
	return func(corectx.Context, dmodel.DynamicFields) (*dyn.OpResult[int], error) {
		*writes++
		return &dyn.OpResult[int]{Data: 1, HasData: true}, nil
	}
}

func TestDirectWriteWithoutATransactionCommitsTheWriteAndItsEntryTogether(t *testing.T) {
	recorder := &tranxRecorder{}
	trail, repo := newTestTrail(recorder)
	writes := 0

	_, err := trail.insert(ownerContext(), dmodel.DynamicFields{"id": "l1", "name": "line"}, insertOne(&writes))

	require.NoError(t, err)
	require.NotNil(t, repo.tranx, "the trail opened its own transaction")
	assert.True(t, repo.tranx.committed)
	assert.Equal(t, 1, writes)
	assert.Len(t, recorder.committed, 1)
}

func TestDirectWriteInAnOuterTransactionLeavesNoEntryWhenItRollsBack(t *testing.T) {
	recorder := &tranxRecorder{}
	trail, repo := newTestTrail(recorder)
	repo.record = dmodel.DynamicFields{"id": "w1", "status": "active"}
	repo.found = true
	outer := &stubTransaction{}
	ctx := ownerContext()
	ctx.SetDbTranx(outer)

	_, err := trail.updateRow(ctx, dmodel.DynamicFields{"id": "w1", "status": "suspended"},
		func(corectx.Context, dmodel.DynamicFields) (*MutateResult, error) { return changedOne(), nil })
	require.NoError(t, err)
	require.NoError(t, outer.Rollback())

	assert.Nil(t, repo.tranx, "the write joined the outer transaction")
	assert.Empty(t, recorder.committed)
}

func TestDirectWriteFailsWhenItsEntryCannotBeRecorded(t *testing.T) {
	recorder := &tranxRecorder{fail: errors.New("audit insert failed")}
	trail, repo := newTestTrail(recorder)
	writes := 0

	_, err := trail.insert(ownerContext(), dmodel.DynamicFields{"id": "l1", "name": "line"}, insertOne(&writes))

	require.Error(t, err)
	assert.Equal(t, 1, writes)
	assert.False(t, repo.tranx.committed)
	assert.True(t, repo.tranx.rolledBack, "the write rolls back with its entry")
}
