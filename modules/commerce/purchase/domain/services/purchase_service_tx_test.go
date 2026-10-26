package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/database"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

type countingTransaction struct {
	commits   int
	rollbacks int
}

func (this *countingTransaction) Commit() error   { this.commits++; return nil }
func (this *countingTransaction) Rollback() error { this.rollbacks++; return nil }

type transactionProbeRepo struct {
	composable.CrudRepository
	begun []*countingTransaction
}

func (this *transactionProbeRepo) Schema() *dmodel.ModelSchema { return &dmodel.ModelSchema{} }

func (this *transactionProbeRepo) BeginTransaction(corectx.Context) (database.DbTransaction, error) {
	tranx := &countingTransaction{}
	this.begun = append(this.begun, tranx)
	return tranx, nil
}

func stubOrderRepo(t *testing.T) *transactionProbeRepo {
	probe := &transactionProbeRepo{}
	original := repoFor
	repoFor = func(string) (composable.CrudRepository, error) { return probe, nil }
	t.Cleanup(func() { repoFor = original })
	return probe
}

// create_rfq creates its order lines through the line service, whose Create opens an order
// transaction of its own. The inner call must join the outer transaction: the database client
// refuses a nested one, which surfaced as a 500.
func TestOrderTransactionJoinsTheOneCtxCarries(t *testing.T) {
	probe := stubOrderRepo(t)
	ctx := corectx.NewRequestContext(context.Background())

	var innerTranx database.DbTransaction
	err := withOrderTransaction(ctx, func(outerCtx corectx.Context) error {
		return withOrderTransaction(outerCtx, func(innerCtx corectx.Context) error {
			innerTranx = innerCtx.GetDbTranx()
			return nil
		})
	})

	require.NoError(t, err)
	require.Len(t, probe.begun, 1, "the nested call must not begin a second transaction")
	assert.Same(t, probe.begun[0], innerTranx)
	assert.Equal(t, 1, probe.begun[0].commits, "only the outer call commits")
}

// A refusal from the inner call aborts the outer transaction rather than committing it.
func TestOrderTransactionRollsBackOnAnInnerRefusal(t *testing.T) {
	probe := stubOrderRepo(t)
	ctx := corectx.NewRequestContext(context.Background())

	err := withOrderTransaction(ctx, func(outerCtx corectx.Context) error {
		return withOrderTransaction(outerCtx, func(corectx.Context) error {
			return composable.ClientErrorSignal{}
		})
	})

	_, isSignal := composable.AsClientErrorSignal(err)
	assert.True(t, isSignal)
	require.Len(t, probe.begun, 1)
	assert.Zero(t, probe.begun[0].commits)
	assert.Equal(t, 1, probe.begun[0].rollbacks)
}
