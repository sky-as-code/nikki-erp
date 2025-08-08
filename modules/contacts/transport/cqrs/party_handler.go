package cqrs

import (
	"context"

	it "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/party"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

func NewPartyHandler(partySvc it.PartyService) *PartyHandler {
	return &PartyHandler{
		PartySvc: partySvc,
	}
}

type PartyHandler struct {
	PartySvc it.PartyService
}

func (this *PartyHandler) CreateParty(ctx context.Context, packet *cqrs.RequestPacket[it.CreatePartyCommand]) (*cqrs.Reply[it.CreatePartyResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.PartySvc.CreateParty)
}

func (this *PartyHandler) UpdateParty(ctx context.Context, packet *cqrs.RequestPacket[it.UpdatePartyCommand]) (*cqrs.Reply[it.UpdatePartyResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.PartySvc.UpdateParty)
}

func (this *PartyHandler) DeleteParty(ctx context.Context, packet *cqrs.RequestPacket[it.DeletePartyCommand]) (*cqrs.Reply[it.DeletePartyResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.PartySvc.DeleteParty)
}

func (this *PartyHandler) GetPartyById(ctx context.Context, packet *cqrs.RequestPacket[it.GetPartyByIdQuery]) (*cqrs.Reply[it.GetPartyByIdResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.PartySvc.GetPartyById)
}

func (this *PartyHandler) SearchParties(ctx context.Context, packet *cqrs.RequestPacket[it.SearchPartiesQuery]) (*cqrs.Reply[it.SearchPartiesResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.PartySvc.SearchParties)
}
