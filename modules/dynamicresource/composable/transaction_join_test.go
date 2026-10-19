package composable

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

func refusingBody(seen *corectx.Context) func(corectx.Context) (*ft.ClientErrors, error) {
	return func(tranxCtx corectx.Context) (*ft.ClientErrors, error) {
		*seen = tranxCtx
		return &ft.ClientErrors{*ft.NewBusinessViolation("name", "err_test", "refused")}, nil
	}
}

func TestInRefusableTransactionRollsBackItsOwnTransactionOnRefusal(t *testing.T) {
	repo := &stubRepository{schema: newPlainSchema()}
	var seen corectx.Context

	refused, err := InRefusableTransaction(ownerContext(), repo, refusingBody(&seen))

	require.NoError(t, err)
	require.NotNil(t, refused)
	assert.Equal(t, 1, refused.Count())
	assert.Same(t, repo.tranx, seen.GetDbTranx(), "the body must run inside the transaction it opened")
	assert.False(t, repo.tranx.committed)
	assert.True(t, repo.tranx.rolledBack)
}

func TestInRefusableTransactionCommitsItsOwnTransactionOnSuccess(t *testing.T) {
	repo := &stubRepository{schema: newPlainSchema()}

	refused, err := InRefusableTransaction(ownerContext(), repo, func(corectx.Context) (*ft.ClientErrors, error) {
		return &ft.ClientErrors{}, nil
	})

	require.NoError(t, err)
	assert.Nil(t, refused, "an empty error list is not a refusal")
	assert.True(t, repo.tranx.committed)
}

func TestInRefusableTransactionRollsBackOnError(t *testing.T) {
	repo := &stubRepository{schema: newPlainSchema()}

	refused, err := InRefusableTransaction(ownerContext(), repo, func(corectx.Context) (*ft.ClientErrors, error) {
		return nil, errors.New("write failed")
	})

	require.Error(t, err)
	assert.Nil(t, refused)
	assert.False(t, repo.tranx.committed)
	assert.True(t, repo.tranx.rolledBack)
}

// Inside a caller's transaction the refusal is reported, and committing or rolling back stays the
// caller's decision.
func TestInRefusableTransactionInsideACallersTransactionOnlyReports(t *testing.T) {
	repo := &stubRepository{schema: newPlainSchema()}
	callerTranx := &stubTransaction{}
	ctx := ownerContext()
	ctx.SetDbTranx(callerTranx)
	var seen corectx.Context

	refused, err := InRefusableTransaction(ctx, repo, refusingBody(&seen))

	require.NoError(t, err)
	require.NotNil(t, refused)
	assert.Equal(t, 1, refused.Count())
	assert.Nil(t, repo.tranx, "no transaction of its own may be opened inside the caller's")
	assert.Same(t, callerTranx, seen.GetDbTranx())
	assert.False(t, callerTranx.committed)
	assert.False(t, callerTranx.rolledBack)
}
