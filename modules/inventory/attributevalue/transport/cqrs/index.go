package cqrs

import (
	"context"
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

func InitCqrsHandlers() error {
	err := errors.Join(
		initAttributeValueHandler(),
	)
	return err
}

func initAttributeValueHandler() error {
	deps.Register(NewAttributeValueHandler)

	return deps.Invoke(func(cqrsBus cqrs.CqrsBus, handler *AttributeValueHandler) error {
		ctx := context.Background()
		return cqrsBus.SubscribeRequests(
			ctx,
			cqrs.NewHandler(handler.CreateAttributeValue),
			cqrs.NewHandler(handler.DeleteAttributeValue),
			cqrs.NewHandler(handler.UpdateAttributeValue),
			cqrs.NewHandler(handler.GetAttributeValueById),
			cqrs.NewHandler(handler.SearchAttributeValues),
		)
	})
}
