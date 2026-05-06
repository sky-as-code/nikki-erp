package commchannel

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type CommChannelService interface {
	CreateCommChannel(ctx corectx.Context, cmd CreateCommChannelCommand) (*CreateCommChannelResult, error)
	DeleteCommChannel(ctx corectx.Context, cmd DeleteCommChannelCommand) (*DeleteCommChannelResult, error)
	GetCommChannel(ctx corectx.Context, query GetCommChannelQuery) (*GetCommChannelResult, error)
	SearchCommChannels(ctx corectx.Context, query SearchCommChannelsQuery) (*SearchCommChannelsResult, error)
	CommChannelExists(ctx corectx.Context, query CommChannelExistsQuery) (*CommChannelExistsResult, error)
	UpdateCommChannel(ctx corectx.Context, cmd UpdateCommChannelCommand) (*UpdateCommChannelResult, error)
}
