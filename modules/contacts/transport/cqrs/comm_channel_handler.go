package cqrs

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/comm_channel"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

func NewCommChannelHandler(commChannelSvc comm_channel.CommChannelService) *CommChannelHandler {
	return &CommChannelHandler{
		CommChannelSvc: commChannelSvc,
	}
}

type CommChannelHandler struct {
	CommChannelSvc comm_channel.CommChannelService
}

func (this *CommChannelHandler) CreateCommChannel(ctx context.Context, packet *cqrs.RequestPacket[comm_channel.CreateCommChannelCommand]) (*cqrs.Reply[comm_channel.CreateCommChannelResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.CommChannelSvc.CreateCommChannel)
}

func (this *CommChannelHandler) UpdateCommChannel(ctx context.Context, packet *cqrs.RequestPacket[comm_channel.UpdateCommChannelCommand]) (*cqrs.Reply[comm_channel.UpdateCommChannelResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.CommChannelSvc.UpdateCommChannel)
}

func (this *CommChannelHandler) DeleteCommChannel(ctx context.Context, packet *cqrs.RequestPacket[comm_channel.DeleteCommChannelCommand]) (*cqrs.Reply[comm_channel.DeleteCommChannelResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.CommChannelSvc.DeleteCommChannel)
}

func (this *CommChannelHandler) GetCommChannelById(ctx context.Context, packet *cqrs.RequestPacket[comm_channel.GetCommChannelByIdQuery]) (*cqrs.Reply[comm_channel.GetCommChannelByIdResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.CommChannelSvc.GetCommChannelById)
}

func (this *CommChannelHandler) SearchCommChannels(ctx context.Context, packet *cqrs.RequestPacket[comm_channel.SearchCommChannelsQuery]) (*cqrs.Reply[comm_channel.SearchCommChannelsResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.CommChannelSvc.SearchCommChannels)
}
