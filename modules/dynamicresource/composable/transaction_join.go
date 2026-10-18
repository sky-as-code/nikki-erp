package composable

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// InTransaction joins the transaction ctx already carries - a bulk create, an import, or a caller
// that opened one - since the database client refuses a nested one, and otherwise opens its own
// through WithTransaction.
func InTransaction(
	ctx corectx.Context, repo CrudRepository, body func(tranxCtx corectx.Context) error,
) error {
	if ctx.GetDbTranx() != nil {
		return body(ctx)
	}
	return WithTransaction(ctx, repo, body)
}

// InRefusableTransaction is InTransaction for a body that may refuse the request with client
// errors after it has already written something. The refusal is carried out as a
// ClientErrorSignal so the transaction rolls back, then handed back as client errors.
//
// When ctx carries the caller's transaction, rolling it back is the caller's decision: the
// refusal is still reported, and the caller sees the client errors before it commits.
func InRefusableTransaction(
	ctx corectx.Context, repo CrudRepository, body func(tranxCtx corectx.Context) (*ft.ClientErrors, error),
) (*ft.ClientErrors, error) {
	err := InTransaction(ctx, repo, func(tranxCtx corectx.Context) error {
		cErrs, err := body(tranxCtx)
		if err != nil {
			return err
		}
		if cErrs != nil && cErrs.Count() > 0 {
			return ClientErrorSignal{Errors: *cErrs}
		}
		return nil
	})
	if signal, ok := AsClientErrorSignal(err); ok {
		return &signal.Errors, nil
	}
	return nil, err
}
