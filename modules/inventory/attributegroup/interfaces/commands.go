package interfaces

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

var CreateAttributeGroupTypes = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "attribute",
	Action:    "create",
}

// Create Command
type CreateAttributeGroupCommand struct {
	ProductId *string         `json:"productId,omitempty" validate:"required"`
	Name      *model.LangJson `json:"name,omitempty" validate:"required"`
	Index     *int            `json:"index,omitempty"`
}

func (CreateAttributeGroupCommand) CqrsRequestType() cqrs.RequestType {
	return CreateAttributeGroupTypes
}

type CreateAttributeGroupResult = GetAttributeGroupByIdResult

// Update Command

var UpdateAttributeGroupCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "attribute",
	Action:    "update",
}

type UpdateAttributeGroupCommand struct {
	Id        model.Id        `json:"id" validate:"required" param:"id"`
	Etag      model.Etag      `json:"etag" validate:"required" header:"If-Match"`
	ProductId *string         `json:"productId,omitempty"`
	Name      *model.LangJson `json:"name,omitempty"`
	Index     *int            `json:"index,omitempty"`
}

func (UpdateAttributeGroupCommand) CqrsRequestType() cqrs.RequestType {
	return UpdateAttributeGroupCommandType
}

type UpdateAttributeGroupResult = GetAttributeGroupByIdResult

// Delete Command
var DeleteAttributeGroupCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "attribute",
	Action:    "delete",
}

type DeleteAttributeGroupCommand struct {
	Id model.Id `json:"id" validate:"required" param:"id"`
}

func (DeleteAttributeGroupCommand) CqrsRequestType() cqrs.RequestType {
	return DeleteAttributeGroupCommandType
}

func (this DeleteAttributeGroupCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}
	return val.ApiBased.ValidateStruct(&this, rules...)
}

type DeleteAttributeGroupResult = crud.DeletionResult

// Get by ID Query
var GetAttributeGroupByIdQueryType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "attribute",
	Action:    "getById",
}

type GetAttributeGroupByIdQuery struct {
	Id model.Id `json:"id" validate:"required" param:"id"`
}

func (GetAttributeGroupByIdQuery) CqrsRequestType() cqrs.RequestType {
	return GetAttributeGroupByIdQueryType
}

func (this GetAttributeGroupByIdQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}
	return val.ApiBased.ValidateStruct(&this, rules...)
}

type GetAttributeGroupByIdResult = crud.OpResult[*AttributeGroup]

// Search Query
var SearchAttributeGroupsQueryType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "attribute",
	Action:    "search",
}

type SearchAttributeGroupsQuery struct {
	crud.SearchQuery
	Criteria  *string `json:"criteria,omitempty" query:"criteria"`
	ProductId *string `json:"productId,omitempty" query:"productId"`
}

func (this SearchAttributeGroupsQuery) Validate() ft.ValidationErrors {
	rules := this.SearchQuery.ValidationRules()

	return val.ApiBased.ValidateStruct(&this, rules...)
}

func (SearchAttributeGroupsQuery) CqrsRequestType() cqrs.RequestType {
	return SearchAttributeGroupsQueryType
}

func (this *SearchAttributeGroupsQuery) SetDefaults() {
	this.SearchQuery.SetDefaults()
}

type SearchAttributeGroupsResultData = crud.PagedResult[AttributeGroup]
type SearchAttributeGroupsResult = crud.OpResult[*SearchAttributeGroupsResultData]
