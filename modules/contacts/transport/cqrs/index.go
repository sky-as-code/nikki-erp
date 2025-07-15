package cqrs

import (
	"context"
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

func InitCqrsHandlers() error {
	err := errors.Join(
		initPartyHandlers(),
	)
	return err
}

func initPartyHandlers() error {
	deps.Register(NewPartyHandler)

	return deps.Invoke(func(cqrsBus cqrs.CqrsBus, handler *PartyHandler) error {
		ctx := context.Background()
		return cqrsBus.SubscribeRequests(
			ctx,
			cqrs.NewHandler(handler.CreatePartyTag),
			cqrs.NewHandler(handler.UpdatePartyTag),
			cqrs.NewHandler(handler.DeletePartyTag),
			cqrs.NewHandler(handler.PartyTagExistsMulti),
			cqrs.NewHandler(handler.GetPartyTagById),
			cqrs.NewHandler(handler.ListPartyTags),
		)
	})
}
