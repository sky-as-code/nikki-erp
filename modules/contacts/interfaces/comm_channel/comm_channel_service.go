package comm_channel

import (
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type CommChannelService interface {
	CreateCommChannel(ctx crud.Context, cmd CreateCommChannelCommand) (*CreateCommChannelResult, error)
	DeleteCommChannel(ctx crud.Context, cmd DeleteCommChannelCommand) (*DeleteCommChannelResult, error)
	GetCommChannelById(ctx crud.Context, query GetCommChannelByIdQuery) (*GetCommChannelByIdResult, error)
	SearchCommChannels(ctx crud.Context, query SearchCommChannelsQuery) (*SearchCommChannelsResult, error)
	UpdateCommChannel(ctx crud.Context, cmd UpdateCommChannelCommand) (*UpdateCommChannelResult, error)
}
