package commchannel

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	enum "github.com/sky-as-code/nikki-erp/modules/core/enum/interfaces"
)

func init() {
	// Assert interface implementation
	var req cqrs.Request
	req = (*CreateCommChannelCommand)(nil)
	req = (*UpdateCommChannelCommand)(nil)
	req = (*DeleteCommChannelCommand)(nil)
	req = (*GetCommChannelByIdQuery)(nil)
	req = (*GetCommChannelsByPartyQuery)(nil)
	req = (*SearchCommChannelsQuery)(nil)
	util.Unused(req)
}

var createCommChannelCommandType = cqrs.RequestType{
	Module:    "contacts",
	Submodule: "comm_channel",
	Action:    "create",
}

type CreateCommChannelCommand struct {
	OrgId     model.Id              `json:"orgId"`
	Note      *string               `json:"note,omitempty"`
	PartyId   model.Id              `json:"partyId"`
	Type      *string               `json:"type"`
	Value     *string               `json:"value,omitempty"`
	ValueJson *domain.ValueJsonData `json:"valueJson,omitempty"`
}

func (CreateCommChannelCommand) CqrsRequestType() cqrs.RequestType {
	return createCommChannelCommandType
}

func (this CreateCommChannelCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		val.Field(&this.Type,
			val.NotNil,
			val.When(this.Type != nil,
				val.NotEmpty,
				val.OneOf("Phone", "Zalo", "Facebook", "Email", "Post"),
			),
		),
		val.Field(&this.Value,
			val.When(this.Value != nil,
				val.NotEmpty,
				val.Length(1, 255),
			),
		),
		model.IdValidateRule(&this.PartyId, true),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type CreateCommChannelResult = crud.OpResult[*domain.CommChannel]

var updateCommChannelCommandType = cqrs.RequestType{
	Module:    "contacts",
	Submodule: "comm_channel",
	Action:    "update",
}

type UpdateCommChannelCommand struct {
	Id        model.Id              `param:"id" json:"id"`
	Note      *string               `json:"note,omitempty"`
	PartyId   model.Id              `param:"partyId" json:"partyId"`
	Type      *string               `json:"type,omitempty"`
	Value     *string               `json:"value,omitempty"`
	ValueJson *domain.ValueJsonData `json:"valueJson,omitempty"`
	Etag      model.Etag            `json:"etag"`
}

func (UpdateCommChannelCommand) CqrsRequestType() cqrs.RequestType {
	return updateCommChannelCommandType
}

func (this UpdateCommChannelCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
		val.Field(&this.Type,
			val.When(this.Type != nil,
				val.NotEmpty,
				val.OneOf("Phone", "Zalo", "Facebook", "Email", "Post"),
			),
		),
		val.Field(&this.Value,
			val.When(this.Value != nil,
				val.NotEmpty,
				val.Length(1, 255),
			),
		),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type UpdateCommChannelResult = crud.OpResult[*domain.CommChannel]

var deleteCommChannelCommandType = cqrs.RequestType{
	Module:    "contacts",
	Submodule: "comm_channel",
	Action:    "delete",
}

type DeleteCommChannelCommand struct {
	Id model.Id `json:"id" param:"id"`
}

func (DeleteCommChannelCommand) CqrsRequestType() cqrs.RequestType {
	return deleteCommChannelCommandType
}

func (this DeleteCommChannelCommand) ToDomainModel() *domain.CommChannel {
	user := &domain.CommChannel{}
	user.Id = &this.Id
	return user
}

func (this DeleteCommChannelCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type DeleteCommChannelResult = crud.DeletionResult

var getCommChannelByIdQueryType = cqrs.RequestType{
	Module:    "contacts",
	Submodule: "comm_channel",
	Action:    "getCommChannelById",
}

type GetCommChannelByIdQuery struct {
	Id        model.Id `param:"id" json:"id"`
	WithParty bool     `json:"withParty" query:"withParty"`
}

func (GetCommChannelByIdQuery) CqrsRequestType() cqrs.RequestType {
	return getCommChannelByIdQueryType
}

func (this GetCommChannelByIdQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type GetCommChannelByIdResult = crud.OpResult[*domain.CommChannel]

var getCommChannelsByPartyQueryType = cqrs.RequestType{
	Module:    "contacts",
	Submodule: "comm_channel",
	Action:    "getCommChannelsByParty",
}

type GetCommChannelsByPartyQuery struct {
	PartyId   model.Id   `param:"partyId" json:"partyId"`
	Type      *enum.Enum `json:"type,omitempty" query:"type"`
	WithParty bool       `json:"withParty" query:"withParty"`
}

func (GetCommChannelsByPartyQuery) CqrsRequestType() cqrs.RequestType {
	return getCommChannelsByPartyQueryType
}

func (this GetCommChannelsByPartyQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.PartyId, true),
		val.Field(&this.Type,
			val.When(this.Type != nil,
				val.NotEmpty,
				val.OneOf("Phone", "Zalo", "Facebook", "Email", "Post"),
			),
		),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type GetCommChannelsByPartyResult = crud.OpResult[[]*domain.CommChannel]

var searchCommChannelsQueryType = cqrs.RequestType{
	Module:    "contacts",
	Submodule: "comm_channel",
	Action:    "search",
}

type SearchCommChannelsQuery struct {
	crud.SearchQuery
}

func (SearchCommChannelsQuery) CqrsRequestType() cqrs.RequestType {
	return searchCommChannelsQueryType
}

func (this SearchCommChannelsQuery) Validate() ft.ValidationErrors {
	rules := this.SearchQuery.ValidationRules()

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type SearchCommChannelsResultData = crud.PagedResult[domain.CommChannel]
type SearchCommChannelsResult = crud.OpResult[*SearchCommChannelsResultData]
