package tranx

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type fakeTxClient struct {
	orm.DbClient
	commits   int
	callbacks []func()
}

func (this *fakeTxClient) Commit() error {
	this.commits++
	for _, fn := range this.callbacks {
		fn()
	}
	this.callbacks = nil
	return nil
}

func (this *fakeTxClient) Rollback() error {
	this.callbacks = nil
	return nil
}

func (this *fakeTxClient) OnCommitted(fn func()) { this.callbacks = append(this.callbacks, fn) }

type fakeDbClient struct {
	orm.DbClient
	begun []*fakeTxClient
}

func (this *fakeDbClient) BeginTx(context.Context, *sql.TxOptions) (orm.DbTxClient, error) {
	tx := &fakeTxClient{}
	this.begun = append(this.begun, tx)
	return tx, nil
}

// A nested Exec joins the outer transaction, so its callbacks must wait for the outermost
// commit; firing them at the inner return would publish a change the outer body may still undo.
func TestOnCommittedInAJoinedExecFiresOnlyAtTheOutermostCommit(t *testing.T) {
	client := &fakeDbClient{}
	manager := NewManager(ManagerParams{Client: client})
	ctx := corectx.NewRequestContext(context.Background())
	ran := false

	err := manager.Exec(ctx, func(outerCtx corectx.Context) error {
		innerErr := manager.Exec(outerCtx, func(innerCtx corectx.Context) error {
			innerCtx.OnCommitted(func() { ran = true })
			return nil
		})
		assert.False(t, ran, "the joined Exec returned but the transaction has not committed")
		return innerErr
	})

	require.NoError(t, err)
	require.Len(t, client.begun, 1, "the nested Exec must join, not open a second transaction")
	assert.Equal(t, 1, client.begun[0].commits)
	assert.True(t, ran)
}

func TestOnCommittedIsDroppedWhenTheOuterExecRollsBack(t *testing.T) {
	manager := NewManager(ManagerParams{Client: &fakeDbClient{}})
	ctx := corectx.NewRequestContext(context.Background())
	ran := false

	err := manager.Exec(ctx, func(outerCtx corectx.Context) error {
		_ = manager.Exec(outerCtx, func(innerCtx corectx.Context) error {
			innerCtx.OnCommitted(func() { ran = true })
			return nil
		})
		return errors.New("outer body failed")
	})

	require.Error(t, err)
	assert.False(t, ran)
}
