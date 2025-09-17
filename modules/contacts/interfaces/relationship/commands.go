package relationship

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/safe"
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
	req = (*CreateRelationshipCommand)(nil)
	req = (*UpdateRelationshipCommand)(nil)
	req = (*DeleteRelationshipCommand)(nil)
	req = (*GetRelationshipByIdQuery)(nil)
	req = (*GetRelationshipsByPartyQuery)(nil)
	req = (*SearchRelationshipsQuery)(nil)
	util.Unused(req)
}

var createRelationshipCommandType = cqrs.RequestType{
	Module:    "contacts",
	Submodule: "relationship",
	Action:    "create",
}

type CreateRelationshipCommand struct {
	PartyId       model.Id `json:"partyId"`
	Note          *string  `json:"note,omitempty"`
	TargetPartyId model.Id `json:"targetPartyId"`
	Type          string   `json:"type"`
}

func (CreateRelationshipCommand) CqrsRequestType() cqrs.RequestType {
	return createRelationshipCommandType
}

func (this CreateRelationshipCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		val.Field(&this.Type,
			val.NotNil,
			val.When(this.Type != "",
				val.NotEmpty,
				val.OneOf("employee", "spouse", "parent", "sibling", "emergency", "subsidiary"),
			),
		),
		model.IdValidateRule(&this.TargetPartyId, true),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type CreateRelationshipResult = crud.OpResult[*domain.Relationship]

var updateRelationshipCommandType = cqrs.RequestType{
	Module:    "contacts",
	Submodule: "relationship",
	Action:    "update",
}

type UpdateRelationshipCommand struct {
	Id            model.Id   `param:"id" json:"id"`
	Note          *string    `json:"note,omitempty"`
	TargetPartyId *model.Id  `json:"targetPartyId,omitempty"`
	Type          *enum.Enum `json:"type,omitempty"`
	Etag          model.Etag `json:"etag"`
}

func (UpdateRelationshipCommand) CqrsRequestType() cqrs.RequestType {
	return updateRelationshipCommandType
}

func (this UpdateRelationshipCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
		val.Field(&this.Type,
			val.When(this.Type != nil,
				val.NotEmpty,
				val.OneOf("employee", "spouse", "parent", "sibling", "emergency", "subsidiary"),
			),
		),
		model.IdPtrValidateRule(&this.TargetPartyId, false),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type UpdateRelationshipResult = crud.OpResult[*domain.Relationship]

var deleteRelationshipCommandType = cqrs.RequestType{
	Module:    "contacts",
	Submodule: "relationship",
	Action:    "delete",
}

type DeleteRelationshipCommand struct {
	Id model.Id `json:"id" param:"id"`
}

func (DeleteRelationshipCommand) CqrsRequestType() cqrs.RequestType {
	return deleteRelationshipCommandType
}

func (this DeleteRelationshipCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type DeleteRelationshipResult = crud.DeletionResult

var getRelationshipByIdQueryType = cqrs.RequestType{
	Module:    "contacts",
	Submodule: "relationship",
	Action:    "getRelationshipById",
}

type GetRelationshipByIdQuery struct {
	Id model.Id `param:"id" json:"id"`
}

func (GetRelationshipByIdQuery) CqrsRequestType() cqrs.RequestType {
	return getRelationshipByIdQueryType
}

func (this GetRelationshipByIdQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type GetRelationshipByIdResult = crud.OpResult[*domain.Relationship]

var getRelationshipsByPartyQueryType = cqrs.RequestType{
	Module:    "contacts",
	Submodule: "relationship",
	Action:    "getRelationshipsByParty",
}

type GetRelationshipsByPartyQuery struct {
	PartyId model.Id   `param:"partyId" json:"partyId"`
	Type    *enum.Enum `json:"type,omitempty" query:"type"`
}

func (GetRelationshipsByPartyQuery) CqrsRequestType() cqrs.RequestType {
	return getRelationshipsByPartyQueryType
}

func (this GetRelationshipsByPartyQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.PartyId, true),
		val.Field(&this.Type,
			val.When(this.Type != nil,
				val.NotEmpty,
				val.OneOf("employee", "spouse", "parent", "sibling", "emergency", "subsidiary"),
			),
		),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type GetRelationshipsByPartyResult = crud.OpResult[[]*domain.Relationship]

var searchRelationshipsQueryType = cqrs.RequestType{
	Module:    "contacts",
	Submodule: "relationship",
	Action:    "search",
}

type SearchRelationshipsQuery struct {
	Page            *int       `json:"page" query:"page"`
	Size            *int       `json:"size" query:"size"`
	Graph           *string    `json:"graph" query:"graph"`
	Type            *enum.Enum `json:"type" query:"type"`
	TargetPartyId   *model.Id  `json:"targetPartyId" query:"targetPartyId"`
	WithTargetParty bool       `json:"withTargetParty" query:"withTargetParty"`
}

func (SearchRelationshipsQuery) CqrsRequestType() cqrs.RequestType {
	return searchRelationshipsQueryType
}

func (this *SearchRelationshipsQuery) SetDefaults() {
	safe.SetDefaultValue(&this.Page, model.MODEL_RULE_PAGE_INDEX_START)
	safe.SetDefaultValue(&this.Size, model.MODEL_RULE_PAGE_DEFAULT_SIZE)
}

func (this SearchRelationshipsQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		crud.PageIndexValidateRule(&this.Page),
		crud.PageSizeValidateRule(&this.Size),
		val.Field(&this.Type,
			val.When(this.Type != nil,
				val.NotEmpty,
				val.OneOf("employee", "spouse", "parent", "sibling", "emergency", "subsidiary"),
			),
		),
		model.IdPtrValidateRule(&this.TargetPartyId, false),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type SearchRelationshipsResultData = crud.PagedResult[domain.Relationship]
type SearchRelationshipsResult = crud.OpResult[*SearchRelationshipsResultData]
