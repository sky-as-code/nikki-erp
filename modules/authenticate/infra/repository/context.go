package repository

import (
	"go.bryk.io/pkg/errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

func asCoreCtx(ctx crud.Context) (corectx.Context, error) {
	c, ok := ctx.(corectx.Context)
	if !ok {
		return nil, errors.New("authenticate repository: context must be corectx.Context")
	}
	return c, nil
}
