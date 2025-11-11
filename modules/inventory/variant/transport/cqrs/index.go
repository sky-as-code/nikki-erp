package cqrs

import (
	"context"
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

func InitCqrsHandlers() error {
	err := errors.Join(
		initVariantHandler(),
	)
	return err
}

func initVariantHandler() error {
	deps.Register(NewVariantHandler)

	return deps.Invoke(func(cqrsBus cqrs.CqrsBus, handler *VariantHandler) error {
		ctx := context.Background()
		return cqrsBus.SubscribeRequests(
			ctx,
			cqrs.NewHandler(handler.CreateVariant),
			cqrs.NewHandler(handler.DeleteVariant),
			cqrs.NewHandler(handler.UpdateVariant),
			cqrs.NewHandler(handler.GetVariantById),
			cqrs.NewHandler(handler.SearchVariants),
		)
	})
}
