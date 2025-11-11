package cqrs

import (
	"context"
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

func InitCqrsHandlers() error {
	err := errors.Join(
		initAttributeGroupHandler(),
	)
	return err
}

func initAttributeGroupHandler() error {
	deps.Register(NewAttributeGroupHandler)

	return deps.Invoke(func(cqrsBus cqrs.CqrsBus, handler *AttributeGroupHandler) error {
		ctx := context.Background()
		return cqrsBus.SubscribeRequests(
			ctx,
			cqrs.NewHandler(handler.CreateAttributeGroup),
			cqrs.NewHandler(handler.UpdateAttributeGroup),
			cqrs.NewHandler(handler.DeleteAttributeGroup),
			cqrs.NewHandler(handler.GetAttributeGroupById),
			cqrs.NewHandler(handler.SearchAttributeGroups),
		)
	})
}
