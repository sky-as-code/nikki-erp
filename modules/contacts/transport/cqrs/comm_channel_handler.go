package cqrs

import (
	"context"

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
	return cqrs.HandlePacket(ctx, packet, this.CommChannelSvc.CreateCommChannel)
}

func (this *CommChannelHandler) UpdateCommChannel(ctx context.Context, packet *cqrs.RequestPacket[itCommChannel.UpdateCommChannelCommand]) (*cqrs.Reply[itCommChannel.UpdateCommChannelResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.CommChannelSvc.UpdateCommChannel)
}

func (this *CommChannelHandler) DeleteCommChannel(ctx context.Context, packet *cqrs.RequestPacket[itCommChannel.DeleteCommChannelCommand]) (*cqrs.Reply[itCommChannel.DeleteCommChannelResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.CommChannelSvc.DeleteCommChannel)
}

func (this *CommChannelHandler) GetCommChannelById(ctx context.Context, packet *cqrs.RequestPacket[itCommChannel.GetCommChannelByIdQuery]) (*cqrs.Reply[itCommChannel.GetCommChannelByIdResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.CommChannelSvc.GetCommChannelById)
}

func (this *CommChannelHandler) SearchCommChannels(ctx context.Context, packet *cqrs.RequestPacket[itCommChannel.SearchCommChannelsQuery]) (*cqrs.Reply[itCommChannel.SearchCommChannelsResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.CommChannelSvc.SearchCommChannels)
}
