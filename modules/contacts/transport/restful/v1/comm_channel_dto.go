package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	itCommChannel "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/commchannel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
)

type CreateCommChannelRequest = itCommChannel.CreateCommChannelCommand
type CreateCommChannelResponse = httpserver.RestCreateResponse

type UpdateCommChannelRequest = itCommChannel.UpdateCommChannelCommand
type UpdateCommChannelResponse = httpserver.RestMutateResponse

type DeleteCommChannelRequest = itCommChannel.DeleteCommChannelCommand
type DeleteCommChannelResponse = httpserver.RestMutateResponse

type GetCommChannelRequest = itCommChannel.GetCommChannelQuery
type GetCommChannelResponse = dmodel.DynamicFields

type CommChannelExistsRequest = itCommChannel.CommChannelExistsQuery
type CommChannelExistsResponse = dynamicmodel.ExistsResultData

type SearchCommChannelsRequest = itCommChannel.SearchCommChannelsQuery
type SearchCommChannelsResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
