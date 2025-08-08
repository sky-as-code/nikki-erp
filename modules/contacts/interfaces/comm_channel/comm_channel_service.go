package comm_channel

import (
	"context"
)

type CommChannelService interface {
	CreateCommChannel(ctx context.Context, cmd CreateCommChannelCommand) (*CreateCommChannelResult, error)
	DeleteCommChannel(ctx context.Context, cmd DeleteCommChannelCommand) (*DeleteCommChannelResult, error)
	GetCommChannelById(ctx context.Context, query GetCommChannelByIdQuery) (*GetCommChannelByIdResult, error)
	GetCommChannelsByParty(ctx context.Context, query GetCommChannelsByPartyQuery) (*GetCommChannelsByPartyResult, error)
	SearchCommChannels(ctx context.Context, query SearchCommChannelsQuery) (*SearchCommChannelsResult, error)
	UpdateCommChannel(ctx context.Context, cmd UpdateCommChannelCommand) (*UpdateCommChannelResult, error)
}
