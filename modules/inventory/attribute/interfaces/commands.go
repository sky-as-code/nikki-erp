package interfaces

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

// Create

var createAttributeCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "attribute",
	Action:    "create",
}

type CreateAttributeCommand struct {
	ProductId     model.Id        `param:"productId" json:"productId"`
	CodeName      string          `json:"codeName"`
	DisplayName   model.LangJson  `json:"displayName"`
	SortIndex     *int            `json:"sortIndex,omitempty"`
	DataType      string          `json:"dataType"`
	IsRequired    bool            `json:"isRequired"`
	IsEnum        *bool           `json:"isEnum,omitempty"`
	EnumValue     *model.LangJson `json:"enumValue,omitempty"`
	EnumValueSort *bool           `json:"enumValueSort,omitempty"`
	GroupId       *model.Id       `json:"groupId,omitempty"`
}

func (CreateAttributeCommand) CqrsRequestType() cqrs.RequestType {
	return createAttributeCommandType
}

type CreateAttributeResult = GetAttributeByIdResult

// Update

var updateAttributeCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "attribute",
	Action:    "update",
}

type UpdateAttributeCommand struct {
	Id            model.Id        `param:"id" json:"id"`
	Etag          model.Etag      `json:"etag"`
	CodeName      *string         `json:"codeName,omitempty"`
	DisplayName   *model.LangJson `json:"displayName,omitempty"`
	SortIndex     *int            `json:"sortIndex,omitempty"`
	DataType      *string         `json:"dataType,omitempty"`
	IsRequired    *bool           `json:"isRequired,omitempty"`
	IsEnum        *bool           `json:"isEnum,omitempty"`
	EnumValue     *model.LangJson `json:"enumValue,omitempty"`
	EnumValueSort *bool           `json:"enumValueSort,omitempty"`
	GroupId       *model.Id       `json:"groupId,omitempty"`
}

func (UpdateAttributeCommand) CqrsRequestType() cqrs.RequestType {
	return updateAttributeCommandType
}

type UpdateAttributeResult = GetAttributeByIdResult

// Delete

var deleteAttributeCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "attribute",
	Action:    "delete",
}

type DeleteAttributeCommand struct {
	Id model.Id `json:"id" param:"id"`
}

func (this DeleteAttributeCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}
	return val.ApiBased.ValidateStruct(&this, rules...)
}

func (DeleteAttributeCommand) CqrsRequestType() cqrs.RequestType {
	return deleteAttributeCommandType
}

type DeleteAttributeResult = crud.DeletionResult

// Get by ID

var getAttributeByIdQueryType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "attribute",
	Action:    "getById",
}

type GetAttributeByIdQuery struct {
	Id model.Id `param:"id" json:"id"`
}

func (this GetAttributeByIdQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}
	return val.ApiBased.ValidateStruct(&this, rules...)
}

func (GetAttributeByIdQuery) CqrsRequestType() cqrs.RequestType {
	return getAttributeByIdQueryType
}

type GetAttributeByIdResult = crud.OpResult[*Attribute]

var searchAttributesQueryType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "attribute",
	Action:    "search",
}

// Search (advanced)

type SearchAttributesQuery struct {
	// Filled by service from Graph
	crud.SearchQuery
}

func (this SearchAttributesQuery) CqrsRequestType() cqrs.RequestType {
	return searchAttributesQueryType
}

func (this SearchAttributesQuery) Validate() ft.ValidationErrors {
	rules := this.SearchQuery.ValidationRules()

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type SearchAttributesResultData = crud.PagedResult[Attribute]
type SearchAttributesResult = crud.OpResult[*SearchAttributesResultData]
