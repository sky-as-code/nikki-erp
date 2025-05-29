package cqrs

import (
	"context"
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

func InitCqrsHandlers() error {
	err := errors.Join(
		initUserHandlers(),
	)
	return err
}

func initUserHandlers() error {
	deps.Register(NewUserHandler)

	return deps.Invoke(func(cqrsBus cqrs.CqrsBus, handler *UserHandler) error {
		ctx := context.Background()
		return cqrsBus.SubscribeRequests(
			ctx,
			cqrs.NewHandler(handler.Create),
			cqrs.NewHandler(handler.Update),
		)
	})
}
