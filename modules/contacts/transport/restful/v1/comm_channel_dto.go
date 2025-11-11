package v1

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain"
	itCommChannel "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/commchannel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
)

type CommChannelDto struct {
	Id        string                `json:"id"`
	OrgId     string                `json:"orgId"`
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

type CreateCommChannelRequest = itCommChannel.CreateCommChannelCommand
type CreateCommChannelResponse = httpserver.RestCreateResponse

type UpdateCommChannelRequest = itCommChannel.UpdateCommChannelCommand
type UpdateCommChannelResponse = httpserver.RestUpdateResponse

type DeleteCommChannelRequest = itCommChannel.DeleteCommChannelCommand
type DeleteCommChannelResponse = httpserver.RestDeleteResponse

type GetCommChannelByIdRequest = itCommChannel.GetCommChannelByIdQuery
type GetCommChannelByIdResponse = CommChannelDto

type SearchCommChannelsRequest = itCommChannel.SearchCommChannelsQuery

type SearchCommChannelsResponse httpserver.RestSearchResponse[CommChannelDto]

func (this *SearchCommChannelsResponse) FromResult(result *itCommChannel.SearchCommChannelsResultData) {
	this.Total = result.Total
	this.Page = result.Page
	this.Size = result.Size
	this.Items = array.Map(result.Items, func(commChannelEntity domain.CommChannel) CommChannelDto {
		item := CommChannelDto{}
		item.FromCommChannel(commChannelEntity)
		return item
	})
}
