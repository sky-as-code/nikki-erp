package cqrs

import (
	"context"
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

func InitCqrsHandlers() error {
	err := errors.Join(
		initAttributeHandler(),
	)
	return err
}

func initAttributeHandler() error {
	deps.Register(NewAttributeHandler)

	return deps.Invoke(func(cqrsBus cqrs.CqrsBus, handler *AttributeHandler) error {
		ctx := context.Background()
		return cqrsBus.SubscribeRequests(
			ctx,
			cqrs.NewHandler(handler.CreateAttribute),
			cqrs.NewHandler(handler.DeleteAttribute),
			cqrs.NewHandler(handler.UpdateAttribute),
			cqrs.NewHandler(handler.GetAttributeById),
			cqrs.NewHandler(handler.SearchAttributes),
		)
	})
}
