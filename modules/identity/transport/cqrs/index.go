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
		initGroupHandlers(),
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
			cqrs.NewHandler(handler.Delete),
			cqrs.NewHandler(handler.GetUserById),
			cqrs.NewHandler(handler.SearchUsers),
			cqrs.NewHandler(handler.Update),
		)
	})
}

func initGroupHandlers() error {
	deps.Register(NewGroupHandler)

	return deps.Invoke(func(cqrsBus cqrs.CqrsBus, handler *GroupHandler) error {
		ctx := context.Background()
		return cqrsBus.SubscribeRequests(
			ctx,
			cqrs.NewHandler(handler.CreateGroup),
			cqrs.NewHandler(handler.DeleteGroup),
			cqrs.NewHandler(handler.GetGroupById),
			cqrs.NewHandler(handler.UpdateGroup),
		)
	})
}
