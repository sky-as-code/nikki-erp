package tranx

import (
	stdErr "errors"

	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// Manager runs a function inside a database transaction shared by every repository
// called during that function, regardless of which model they own.
type Manager interface {
	// Exec begins a transaction, puts it on the context and commits it when fn returns nil,
	// rolling back on error or panic. When the context already carries a transaction, fn joins
	// it and only the outermost Exec commits or rolls back.
	Exec(ctx corectx.Context, fn func(tranxCtx corectx.Context) error) error
}

type ManagerParams struct {
	dig.In

	Client orm.DbClient
}

func NewManager(params ManagerParams) Manager {
	return &ManagerImpl{
		client: params.Client,
	}
}

type ManagerImpl struct {
	client orm.DbClient
}

func (this *ManagerImpl) Exec(ctx corectx.Context, fn func(tranxCtx corectx.Context) error) (err error) {
	if ctx.GetDbTranx() != nil {
		return fn(ctx)
	}

	tranx, err := this.client.BeginTx(ctx.InnerContext(), nil)
	if err != nil {
		return err
	}
	ctx.SetDbTranx(tranx)

	defer func() {
		if e := ft.RecoverPanic(recover(), "Manager.Exec"); e != nil {
			err = e
		}
		ctx.SetDbTranx(nil)

		if err != nil {
			if rbErr := tranx.Rollback(); rbErr != nil {
				err = stdErr.Join(err, rbErr)
			}
			return
		}
		err = tranx.Commit()
	}()

	return fn(ctx)
}

func ExecFor[TResult any](
	ctx corectx.Context,
	manager Manager,
	fn func(tranxCtx corectx.Context) (*TResult, error),
) (*TResult, error) {
	var result *TResult

	err := manager.Exec(ctx, func(tranxCtx corectx.Context) error {
		var err error
		result, err = fn(tranxCtx)
		return err
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}
