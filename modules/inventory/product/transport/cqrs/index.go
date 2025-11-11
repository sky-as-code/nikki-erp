package cqrs

import (
	"context"
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

func InitCqrsHandlers() error {
	err := errors.Join(
		initProductHandler(),
	)
	return err
}

func initProductHandler() error {
	deps.Register(NewProductHandler)

	return deps.Invoke(func(cqrsBus cqrs.CqrsBus, handler *ProductHandler) error {
		ctx := context.Background()
		return cqrsBus.SubscribeRequests(
			ctx,
			cqrs.NewHandler(handler.CreateProduct),
			cqrs.NewHandler(handler.DeleteProduct),
			cqrs.NewHandler(handler.UpdateProduct),
			cqrs.NewHandler(handler.GetProductById),
			cqrs.NewHandler(handler.SearchProducts),
		)
	})
}
