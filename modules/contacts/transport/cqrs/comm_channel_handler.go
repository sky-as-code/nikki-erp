package cqrs

import (
	"context"

	c "github.com/sky-as-code/nikki-erp/modules/contacts/constants"
	itCommChannel "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/commchannel"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

func NewCommChannelHandler(commChannelSvc itCommChannel.CommChannelService) *CommChannelHandler {
	return &CommChannelHandler{
		CommChannelSvc: commChannelSvc,
	}
}

type CommChannelHandler struct {
	CommChannelSvc itCommChannel.CommChannelService
}

func (this *CommChannelHandler) CreateCommChannel(ctx context.Context, packet *cqrs.RequestPacket[itCommChannel.CreateCommChannelCommand]) (*cqrs.Reply[itCommChannel.CreateCommChannelResult], error) {
	return cqrs.ServePacket(ctx, string(c.ContactsModuleName), packet, this.CommChannelSvc.CreateCommChannel)
}

func (this *CommChannelHandler) UpdateCommChannel(ctx context.Context, packet *cqrs.RequestPacket[itCommChannel.UpdateCommChannelCommand]) (*cqrs.Reply[itCommChannel.UpdateCommChannelResult], error) {
	return cqrs.ServePacket(ctx, string(c.ContactsModuleName), packet, this.CommChannelSvc.UpdateCommChannel)
}

func (this *CommChannelHandler) DeleteCommChannel(ctx context.Context, packet *cqrs.RequestPacket[itCommChannel.DeleteCommChannelCommand]) (*cqrs.Reply[itCommChannel.DeleteCommChannelResult], error) {
	return cqrs.ServePacket(ctx, string(c.ContactsModuleName), packet, this.CommChannelSvc.DeleteCommChannel)
}

func (this *CommChannelHandler) GetCommChannel(ctx context.Context, packet *cqrs.RequestPacket[itCommChannel.GetCommChannelQuery]) (*cqrs.Reply[itCommChannel.GetCommChannelResult], error) {
	return cqrs.ServePacket(ctx, string(c.ContactsModuleName), packet, this.CommChannelSvc.GetCommChannel)
}

func (this *CommChannelHandler) SearchCommChannels(ctx context.Context, packet *cqrs.RequestPacket[itCommChannel.SearchCommChannelsQuery]) (*cqrs.Reply[itCommChannel.SearchCommChannelsResult], error) {
	return cqrs.ServePacket(ctx, string(c.ContactsModuleName), packet, this.CommChannelSvc.SearchCommChannels)
}

func (this *CommChannelHandler) CommChannelExists(ctx context.Context, packet *cqrs.RequestPacket[itCommChannel.CommChannelExistsQuery]) (*cqrs.Reply[itCommChannel.CommChannelExistsResult], error) {
	return cqrs.ServePacket(ctx, string(c.ContactsModuleName), packet, this.CommChannelSvc.CommChannelExists)
}
