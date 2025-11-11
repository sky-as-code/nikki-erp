package interfaces

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

// Create Command
var createUnitCategoryCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "unit_category",
	Action:    "create",
}

type CreateUnitCategoryCommand struct {
	OrgId        *string         `json:"orgId,omitempty" validate:"required"`
	Name         *model.LangJson `json:"name,omitempty" validate:"required"`
	Description  *model.LangJson `json:"description,omitempty"`
	Status       *string         `json:"status,omitempty"`
	ThumbnailUrl *string         `json:"thumbnailURL,omitempty"`
}

func (CreateUnitCategoryCommand) CqrsRequestType() cqrs.RequestType {
	return createUnitCategoryCommandType
}

type CreateUnitCategoryResult = GetUnitCategoryByIdResult

// Update Command
var updateUnitCategoryCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "unit_category",
	Action:    "update",
}

type UpdateUnitCategoryCommand struct {
	Id           model.Id        `json:"id" validate:"required" param:"id"`
	Etag         model.Etag      `json:"etag" validate:"required" header:"If-Match"`
	Name         *model.LangJson `json:"name,omitempty"`
	Description  *model.LangJson `json:"description,omitempty"`
	Status       *string         `json:"status,omitempty"`
	ThumbnailUrl *string         `json:"thumbnailURL,omitempty"`
}

func (UpdateUnitCategoryCommand) CqrsRequestType() cqrs.RequestType {
	return updateUnitCategoryCommandType
}

type UpdateUnitCategoryResult = GetUnitCategoryByIdResult

// Delete Command
var deleteUnitCategoryCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "unit_category",
	Action:    "delete",
}

type DeleteUnitCategoryCommand struct {
	Id model.Id `json:"id" validate:"required" param:"id"`
}

func (DeleteUnitCategoryCommand) CqrsRequestType() cqrs.RequestType {
	return deleteUnitCategoryCommandType
}

func (this DeleteUnitCategoryCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}
	return val.ApiBased.ValidateStruct(&this, rules...)
}

type DeleteUnitCategoryResult = crud.DeletionResult

// Get by ID Quer
var getUnitCategoryByIdQueryType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "unit_category",
	Action:    "get_by_id",
}

type GetUnitCategoryByIdQuery struct {
	Id model.Id `json:"id" validate:"required" param:"id"`
}

func (this GetUnitCategoryByIdQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}
	return val.ApiBased.ValidateStruct(&this, rules...)
}

func (GetUnitCategoryByIdQuery) CqrsRequestType() cqrs.RequestType {
	return getUnitCategoryByIdQueryType
}

type GetUnitCategoryByIdResult = crud.OpResult[*UnitCategory]

// Search Query
var searchUnitCategoriesQueryType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "unit_category",
	Action:    "search",
}

type SearchUnitCategoriesQuery struct {
	crud.SearchQuery
	Criteria *string `json:"criteria,omitempty" query:"criteria"`
}

func (this SearchUnitCategoriesQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		// Add validation rules if needed
	}
	return val.ApiBased.ValidateStruct(&this, rules...)
}

func (SearchUnitCategoriesQuery) CqrsRequestType() cqrs.RequestType {
	return searchUnitCategoriesQueryType
}

func (this *SearchUnitCategoriesQuery) SetDefaults() {
	this.SearchQuery.SetDefaults()
}

type SearchUnitCategoriesResultData = crud.PagedResult[UnitCategory]
type SearchUnitCategoriesResult = crud.OpResult[*SearchUnitCategoriesResultData]
