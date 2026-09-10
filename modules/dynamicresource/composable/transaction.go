package composable

import (
	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// WithTransaction runs body inside one database transaction opened on repo, committing when
// body returns nil and rolling back otherwise. The body receives a cloned request context that
// carries the transaction, so every repository call made through it joins the same unit.
func WithTransaction(
	ctx corectx.Context, repo CrudRepository, body func(tranxCtx corectx.Context) error,
) error {
	tranx, err := repo.BeginTransaction(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to begin a transaction on '%s'", repo.Schema().Name())
	}
	defer tranx.Rollback()

	tranxCtx := corectx.CloneRequestContext(ctx)
	tranxCtx.SetDbTranx(tranx)

	if err := body(tranxCtx); err != nil {
		return err
	}
	return errors.Wrapf(tranx.Commit(), "failed to commit a transaction on '%s'", repo.Schema().Name())
}

// ClientErrorSignal carries a rule violation out through a transaction body, so a rejected
// operation rolls back rather than committing half of itself. It is an error only in the
// plumbing sense: the caller unwraps it back into client errors, which answer 400 rather than
// 500.
type ClientErrorSignal struct {
	Errors ft.ClientErrors
}

func (this ClientErrorSignal) Error() string {
	return "the operation was rejected by a business rule"
}

func AsClientErrorSignal(err error) (ClientErrorSignal, bool) {
	signal, ok := err.(ClientErrorSignal)
	return signal, ok
}
