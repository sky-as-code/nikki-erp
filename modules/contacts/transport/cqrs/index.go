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
		initRelationshipHandlers(),
		initCommChannelHandlers(),
	)
	return err
}

func initPartyHandlers() error {
	deps.Register(NewPartyHandler)

	return deps.Invoke(func(cqrsBus cqrs.CqrsBus, handler *PartyHandler) error {
		ctx := context.Background()
		return cqrsBus.SubscribeRequests(
			ctx,
			cqrs.NewHandler(handler.CreateParty),
			cqrs.NewHandler(handler.UpdateParty),
			cqrs.NewHandler(handler.DeleteParty),
			cqrs.NewHandler(handler.GetPartyById),
			cqrs.NewHandler(handler.SearchParties),
		)
	})
}

func initRelationshipHandlers() error {
	deps.Register(NewRelationshipHandler)

	return deps.Invoke(func(cqrsBus cqrs.CqrsBus, handler *RelationshipHandler) error {
		ctx := context.Background()
		return cqrsBus.SubscribeRequests(
			ctx,
			cqrs.NewHandler(handler.CreateRelationship),
			cqrs.NewHandler(handler.UpdateRelationship),
			cqrs.NewHandler(handler.DeleteRelationship),
			cqrs.NewHandler(handler.GetRelationshipById),
			cqrs.NewHandler(handler.SearchRelationships),
		)
	})
}

func initCommChannelHandlers() error {
	deps.Register(NewCommChannelHandler)

	return deps.Invoke(func(cqrsBus cqrs.CqrsBus, handler *CommChannelHandler) error {
		ctx := context.Background()
		return cqrsBus.SubscribeRequests(
			ctx,
			cqrs.NewHandler(handler.CreateCommChannel),
			cqrs.NewHandler(handler.UpdateCommChannel),
			cqrs.NewHandler(handler.DeleteCommChannel),
			cqrs.NewHandler(handler.GetCommChannelById),
			cqrs.NewHandler(handler.SearchCommChannels),
		)
	})
}
