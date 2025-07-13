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

// func (this *PartyHandler) CreateParty(ctx context.Context, packet *cqrs.RequestPacket[it.CreatePartyCommand]) (*cqrs.Reply[it.CreatePartyResult], error) {
// 	return cqrs.HandlePacket[it.CreatePartyCommand, it.CreatePartyResult](ctx, packet, this.PartySvc.CreateParty)
// }

// func (this *PartyHandler) UpdateParty(ctx context.Context, packet *cqrs.RequestPacket[it.UpdatePartyCommand]) (*cqrs.Reply[it.UpdatePartyResult], error) {
// 	return cqrs.HandlePacket[it.UpdatePartyCommand, it.UpdatePartyResult](ctx, packet, this.PartySvc.UpdateParty)
// }

// func (this *PartyHandler) DeleteParty(ctx context.Context, packet *cqrs.RequestPacket[it.DeletePartyCommand]) (*cqrs.Reply[it.DeletePartyResult], error) {
// 	return cqrs.HandlePacket[it.DeletePartyCommand, it.DeletePartyResult](ctx, packet, this.PartySvc.DeleteParty)
// }

// func (this *PartyHandler) PartyExists(ctx context.Context, packet *cqrs.RequestPacket[it.PartyExistsQuery]) (*cqrs.Reply[it.PartyExistsResult], error) {
// 	return cqrs.HandlePacket[it.PartyExistsQuery, it.PartyExistsResult](ctx, packet, this.PartySvc.PartyExists)
// }

// func (this *PartyHandler) PartyExistsMulti(ctx context.Context, packet *cqrs.RequestPacket[it.PartyExistsMultiQuery]) (*cqrs.Reply[it.PartyExistsMultiResult], error) {
// 	return cqrs.HandlePacket[it.PartyExistsMultiQuery, it.PartyExistsMultiResult](ctx, packet, this.PartySvc.PartyExistsMulti)
// }

// func (this *PartyHandler) GetParty(ctx context.Context, packet *cqrs.RequestPacket[it.GetPartyQuery]) (*cqrs.Reply[it.GetPartyResult], error) {
// 	return cqrs.HandlePacket[it.GetPartyQuery, it.GetPartyResult](ctx, packet, this.PartySvc.GetParty)
// }

// func (this *PartyHandler) ListPartys(ctx context.Context, packet *cqrs.RequestPacket[it.ListPartysQuery]) (*cqrs.Reply[it.ListPartysResult], error) {
// 	return cqrs.HandlePacket[it.ListPartysQuery, it.ListPartysResult](ctx, packet, this.PartySvc.ListPartys)
// }

func (this *PartyHandler) CreatePartyTag(ctx context.Context, packet *cqrs.RequestPacket[it.CreatePartyTagCommand]) (*cqrs.Reply[it.CreatePartyTagResult], error) {
	return cqrs.HandlePacket[it.CreatePartyTagCommand, it.CreatePartyTagResult](ctx, packet, this.PartySvc.CreatePartyTag)
}

func (this *PartyHandler) UpdatePartyTag(ctx context.Context, packet *cqrs.RequestPacket[it.UpdatePartyTagCommand]) (*cqrs.Reply[it.UpdatePartyTagResult], error) {
	return cqrs.HandlePacket[it.UpdatePartyTagCommand, it.UpdatePartyTagResult](ctx, packet, this.PartySvc.UpdatePartyTag)
}

func (this *PartyHandler) DeletePartyTag(ctx context.Context, packet *cqrs.RequestPacket[it.DeletePartyTagCommand]) (*cqrs.Reply[it.DeletePartyTagResult], error) {
	return cqrs.HandlePacket[it.DeletePartyTagCommand, it.DeletePartyTagResult](ctx, packet, this.PartySvc.DeletePartyTag)
}

func (this *PartyHandler) PartyTagExistsMulti(ctx context.Context, packet *cqrs.RequestPacket[it.PartyTagExistsMultiQuery]) (*cqrs.Reply[it.PartyTagExistsMultiResult], error) {
	return cqrs.HandlePacket[it.PartyTagExistsMultiQuery, it.PartyTagExistsMultiResult](ctx, packet, this.PartySvc.PartyTagExistsMulti)
}

func (this *PartyHandler) GetPartyTagById(ctx context.Context, packet *cqrs.RequestPacket[it.GetPartyByIdTagQuery]) (*cqrs.Reply[it.GetPartyTagByIdResult], error) {
	return cqrs.HandlePacket[it.GetPartyByIdTagQuery, it.GetPartyTagByIdResult](ctx, packet, this.PartySvc.GetPartyTagById)
}

func (this *PartyHandler) ListPartyTags(ctx context.Context, packet *cqrs.RequestPacket[it.ListPartyTagsQuery]) (*cqrs.Reply[it.ListPartyTagsResult], error) {
	return cqrs.HandlePacket[it.ListPartyTagsQuery, it.ListPartyTagsResult](ctx, packet, this.PartySvc.ListPartyTags)
}
