package productcategory

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
)

// Create

var createProductCategoryCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "product_category",
	Action:    "create",
}

type CreateProductCategoryCommand struct {
	ParentId    *model.Id       `json:"parentId,omitempty"`
	Name        model.LangJson  `json:"name"`
	Description *model.LangJson `json:"description,omitempty"`
	Path        *string         `json:"path,omitempty"`
	Level       *int            `json:"level,omitempty"`
	SortIndex   *int            `json:"sortIndex,omitempty"`
}

func (CreateProductCategoryCommand) CqrsRequestType() cqrs.RequestType {
	return createProductCategoryCommandType
}

type CreateProductCategoryResult = GetProductCategoryByIdResult

// Update

var updateProductCategoryCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "product_category",
	Action:    "update",
}

type UpdateProductCategoryCommand struct {
	Id          model.Id        `param:"id" json:"id"`
	Etag        model.Etag      `json:"etag"`
	ParentId    *model.Id       `json:"parentId,omitempty"`
	Name        *model.LangJson `json:"name,omitempty"`
	Description *model.LangJson `json:"description,omitempty"`
	Path        *string         `json:"path,omitempty"`
	Level       *int            `json:"level,omitempty"`
	SortIndex   *int            `json:"sortIndex,omitempty"`
}

func (UpdateProductCategoryCommand) CqrsRequestType() cqrs.RequestType {
	return updateProductCategoryCommandType
}

type UpdateProductCategoryResult = GetProductCategoryByIdResult

// Delete

var deleteProductCategoryCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "product_category",
	Action:    "delete",
}

type DeleteProductCategoryCommand struct {
	Id model.Id `json:"id" param:"id"`
}

func (this DeleteProductCategoryCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}
	return val.ApiBased.ValidateStruct(&this, rules...)
}

func (DeleteProductCategoryCommand) CqrsRequestType() cqrs.RequestType {
	return deleteProductCategoryCommandType
}

type DeleteProductCategoryResult = crud.DeletionResult

// Get by ID

var getProductCategoryByIdQueryType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "product_category",
	Action:    "getById",
}

type GetProductCategoryByIdQuery struct {
	Id model.Id `param:"id" json:"id"`
}

func (this GetProductCategoryByIdQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}
	return val.ApiBased.ValidateStruct(&this, rules...)
}

func (GetProductCategoryByIdQuery) CqrsRequestType() cqrs.RequestType {
	return getProductCategoryByIdQueryType
}

type GetProductCategoryByIdResult = crud.OpResult[*domain.ProductCategory]

var searchProductCategoriesQueryType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "product_category",
	Action:    "search",
}

// Search (advanced)

type SearchProductCategoriesQuery struct {
	// Filled by service from Graph
	crud.SearchQuery
}

func (this SearchProductCategoriesQuery) CqrsRequestType() cqrs.RequestType {
	return searchProductCategoriesQueryType
}

func (this SearchProductCategoriesQuery) Validate() ft.ValidationErrors {
	rules := this.SearchQuery.ValidationRules()

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type SearchProductCategoriesResultData = crud.PagedResult[domain.ProductCategory]
type SearchProductCategoriesResult = crud.OpResult[*SearchProductCategoriesResultData]
