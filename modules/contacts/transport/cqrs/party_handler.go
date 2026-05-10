package cqrs

import (
	"context"

	c "github.com/sky-as-code/nikki-erp/modules/contacts/constants"
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
	return cqrs.ServePacket(ctx, string(c.ContactsModuleName), packet, this.PartySvc.CreateParty)
}

func (this *PartyHandler) UpdateParty(ctx context.Context, packet *cqrs.RequestPacket[it.UpdatePartyCommand]) (*cqrs.Reply[it.UpdatePartyResult], error) {
	return cqrs.ServePacket(ctx, string(c.ContactsModuleName), packet, this.PartySvc.UpdateParty)
}

func (this *PartyHandler) DeleteParty(ctx context.Context, packet *cqrs.RequestPacket[it.DeletePartyCommand]) (*cqrs.Reply[it.DeletePartyResult], error) {
	return cqrs.ServePacket(ctx, string(c.ContactsModuleName), packet, this.PartySvc.DeleteParty)
}

func (this *PartyHandler) GetParty(ctx context.Context, packet *cqrs.RequestPacket[it.GetPartyQuery]) (*cqrs.Reply[it.GetPartyResult], error) {
	return cqrs.ServePacket(ctx, string(c.ContactsModuleName), packet, this.PartySvc.GetParty)
}

func (this *PartyHandler) SearchParties(ctx context.Context, packet *cqrs.RequestPacket[it.SearchPartiesQuery]) (*cqrs.Reply[it.SearchPartiesResult], error) {
	return cqrs.ServePacket(ctx, string(c.ContactsModuleName), packet, this.PartySvc.SearchParties)
}

func (this *PartyHandler) PartyExists(ctx context.Context, packet *cqrs.RequestPacket[it.PartyExistsQuery]) (*cqrs.Reply[it.PartyExistsResult], error) {
	return cqrs.ServePacket(ctx, string(c.ContactsModuleName), packet, this.PartySvc.PartyExists)
}
