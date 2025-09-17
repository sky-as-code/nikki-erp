package party

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

func init() {
	// Assert interface implementation
	var req cqrs.Request
	req = (*CreatePartyCommand)(nil)
	req = (*UpdatePartyCommand)(nil)
	req = (*DeletePartyCommand)(nil)
	req = (*GetPartyByIdQuery)(nil)
	req = (*GetPartyByDisplayNameQuery)(nil)
	req = (*SearchPartiesQuery)(nil)
	util.Unused(req)
}

var createPartyCommandType = cqrs.RequestType{
	Module:    "contacts",
	Submodule: "party",
	Action:    "create",
}

type CreatePartyCommand struct {
	OrgId        model.Id  `json:"orgId" param:"orgId"`
	AvatarUrl    *string   `json:"avatarUrl,omitempty"`
	DisplayName  string    `json:"displayName"`
	LegalName    *string   `json:"legalName,omitempty"`
	LegalAddress *string   `json:"legalAddress,omitempty"`
	TaxId        *string   `json:"taxId,omitempty"`
	JobPosition  *string   `json:"jobPosition,omitempty"`
	Title        *string   `json:"title,omitempty"`
	Type         string    `json:"type"`
	Note         *string   `json:"note,omitempty"`
	Nationality  *model.Id `json:"nationality,omitempty"`
	Language     *model.Id `json:"language,omitempty"`
	Website      *string   `json:"website,omitempty"`
}

func (CreatePartyCommand) CqrsRequestType() cqrs.RequestType {
	return createPartyCommandType
}

func (this CreatePartyCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		val.Field(&this.DisplayName,
			val.NotEmpty,
			val.Length(1, 50),
		),
		val.Field(&this.Type,
			val.NotEmpty,
			val.OneOf("individual", "company"),
		),

		model.IdPtrValidateRule(&this.Nationality, true),
		model.IdPtrValidateRule(&this.Language, true),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type CreatePartyResult = crud.OpResult[*domain.Party]

var updatePartyCommandType = cqrs.RequestType{
	Module:    "contacts",
	Submodule: "party",
	Action:    "update",
}

type UpdatePartyCommand struct {
	Id           model.Id   `param:"id" json:"id"`
	AvatarUrl    *string    `json:"avatarUrl,omitempty"`
	DisplayName  *string    `json:"displayName,omitempty"`
	LegalName    *string    `json:"legalName,omitempty"`
	LegalAddress *string    `json:"legalAddress,omitempty"`
	TaxId        *string    `json:"taxId,omitempty"`
	JobPosition  *string    `json:"jobPosition,omitempty"`
	Title        *string    `json:"title,omitempty"`
	Type         *string    `json:"type,omitempty"`
	Note         *string    `json:"note,omitempty"`
	Nationality  *model.Id  `json:"nationality,omitempty"`
	Org          *model.Id  `json:"org,omitempty" param:"org"`
	Language     *model.Id  `json:"language,omitempty"`
	Website      *string    `json:"website,omitempty"`
	Etag         model.Etag `json:"etag"`
}

func (UpdatePartyCommand) CqrsRequestType() cqrs.RequestType {
	return updatePartyCommandType
}

func (this UpdatePartyCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
		val.Field(&this.DisplayName,
			val.When(this.DisplayName != nil,
				val.NotEmpty,
				val.Length(1, 50),
			),
		),
		val.Field(&this.Type,
			val.When(this.Type != nil,
				val.NotEmpty,
				val.OneOf("individual", "company"),
			),
		),

		model.IdPtrValidateRule(&this.Nationality, true),
		model.IdPtrValidateRule(&this.Language, true),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type UpdatePartyResult = crud.OpResult[*domain.Party]

var deletePartyCommandType = cqrs.RequestType{
	Module:    "contacts",
	Submodule: "party",
	Action:    "delete",
}

type DeletePartyCommand struct {
	Id model.Id `json:"id" param:"id"`
}

func (DeletePartyCommand) CqrsRequestType() cqrs.RequestType {
	return deletePartyCommandType
}

func (this DeletePartyCommand) ToDomainModel() *domain.Party {
	party := &domain.Party{}
	party.Id = &this.Id
	return party
}

func (this DeletePartyCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type DeletePartyResult = crud.DeletionResult

var getPartyByIdQueryType = cqrs.RequestType{
	Module:    "contacts",
	Submodule: "party",
	Action:    "getPartyById",
}

type GetPartyByIdQuery struct {
	Id                model.Id `param:"id" json:"id"`
	WithCommChannels  bool     `json:"withCommChannels" query:"withCommChannels"`
	WithRelationships bool     `json:"withRelationships" query:"withRelationships"`
}

func (GetPartyByIdQuery) CqrsRequestType() cqrs.RequestType {
	return getPartyByIdQueryType
}

func (this GetPartyByIdQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type GetPartyByIdResult = crud.OpResult[*domain.Party]

var getPartyByDisplayNameQueryType = cqrs.RequestType{
	Module:    "contacts",
	Submodule: "party",
	Action:    "getPartyByDisplayName",
}

type GetPartyByDisplayNameQuery struct {
	DisplayName       string `param:"displayName" json:"displayName"`
	WithCommChannels  bool   `json:"withCommChannels" query:"withCommChannels"`
	WithRelationships bool   `json:"withRelationships" query:"withRelationships"`
}

func (GetPartyByDisplayNameQuery) CqrsRequestType() cqrs.RequestType {
	return getPartyByDisplayNameQueryType
}

func (this GetPartyByDisplayNameQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		val.Field(&this.DisplayName,
			val.NotEmpty,
			val.Length(1, 50),
		),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type GetPartyByDisplayNameResult = crud.OpResult[*domain.Party]

var searchPartiesQueryType = cqrs.RequestType{
	Module:    "contacts",
	Submodule: "party",
	Action:    "search",
}

type SearchPartiesQuery struct {
	crud.SearchQuery
	Type              *string `json:"type" query:"type"`
	WithCommChannels  bool    `json:"withCommChannels" query:"withCommChannels"`
	WithRelationships bool    `json:"withRelationships" query:"withRelationships"`
}

func (SearchPartiesQuery) CqrsRequestType() cqrs.RequestType {
	return searchPartiesQueryType
}

func (this SearchPartiesQuery) Validate() ft.ValidationErrors {
	rules := this.SearchQuery.ValidationRules()

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type SearchPartiesResultData = crud.PagedResult[domain.Party]
type SearchPartiesResult = crud.OpResult[*SearchPartiesResultData]
