package v1

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain"
	"github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/comm_channel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
)

type CommChannelDto struct {
	Id        string                `json:"id"`
	PartyId   string                `json:"partyId"`
	Type      string                `json:"type"`
	Value     *string               `json:"value"`
	ValueJson *domain.ValueJsonData `json:"valueJson,omitempty"`
	CreatedAt int64                 `json:"createdAt"`
	UpdatedAt *int64                `json:"updatedAt,omitempty"`
	Etag      string                `json:"etag"`

	Party *PartyDto `json:"party,omitempty"`
}

func (this *CommChannelDto) FromCommChannel(commChannelEntity domain.CommChannel) {
	model.MustCopy(commChannelEntity.AuditableBase, this)
	model.MustCopy(commChannelEntity.ModelBase, this)
	model.MustCopy(commChannelEntity, this)

	// Handle optional related party
	if commChannelEntity.Party != nil {
		this.Party = &PartyDto{}
		this.Party.FromParty(*commChannelEntity.Party)
	}
}

type CreateCommChannelRequest = comm_channel.CreateCommChannelCommand
type CreateCommChannelResponse = httpserver.RestCreateResponse

type UpdateCommChannelRequest = comm_channel.UpdateCommChannelCommand
type UpdateCommChannelResponse = httpserver.RestUpdateResponse

type DeleteCommChannelRequest = comm_channel.DeleteCommChannelCommand
type DeleteCommChannelResponse = httpserver.RestDeleteResponse

type GetCommChannelByIdRequest = comm_channel.GetCommChannelByIdQuery
type GetCommChannelByIdResponse = CommChannelDto

type SearchCommChannelsRequest = comm_channel.SearchCommChannelsQuery

type SearchCommChannelsResponse httpserver.RestSearchResponse[CommChannelDto]

func (this *SearchCommChannelsResponse) FromResult(result *comm_channel.SearchCommChannelsResultData) {
	this.Total = result.Total
	this.Page = result.Page
	this.Size = result.Size
	this.Items = array.Map(result.Items, func(commChannelEntity domain.CommChannel) CommChannelDto {
		item := CommChannelDto{}
		item.FromCommChannel(commChannelEntity)
		return item
	})
}
