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
		initPermissionHandlers(),
	)
	return err
}

func initUserHandlers() error {
	deps.Register(NewUserHandler)

	return deps.Invoke(func(cqrsBus cqrs.CqrsBus, handler *UserHandler) error {
		ctx := context.Background()
		return cqrsBus.SubscribeRequests(
			ctx,
			cqrs.NewHandler(handler.GetEnabledUser),
		)
	})
}

func initPermissionHandlers() error {
	deps.Register(NewPermissionHandler)

	return deps.Invoke(func(cqrsBus cqrs.CqrsBus, handler *PermissionHandler) error {
		ctx := context.Background()
		return cqrsBus.SubscribeRequests(
			ctx,
			cqrs.NewHandler(handler.IsAuthorized),
			cqrs.NewHandler(handler.GetUserEntitlements),
		)
	})
}
