package product

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
)

// Create

var createProductCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "product",
	Action:    "create",
}

type CreateProductCommand struct {
	OrgId         model.Id        `param:"orgId" json:"orgId"`
	BarCode       string          `json:"barCode,omitempty"`
	Name          model.LangJson  `json:"name"`
	Description   *model.LangJson `json:"description,omitempty"`
	UnitId        *model.Id       `json:"unitId,omitempty"`
	Sku           string          `json:"sku,omitempty"`
	Status        *string         `json:"status,omitempty"`
	ThumbnailUrl  *string         `json:"thumbnailUrl,omitempty"`
	ProposedPrice float64         `json:"proposedPrice,omitempty"`
}

func (CreateProductCommand) CqrsRequestType() cqrs.RequestType {
	return createProductCommandType
}

func (this CreateProductCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.OrgId, true),
	}
	return val.ApiBased.ValidateStruct(&this, rules...)
}

type CreateProductResult = GetProductByIdResult

// Update

var updateProductCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "product",
	Action:    "update",
}

type UpdateProductCommand struct {
	Id                model.Id        `param:"id" json:"id"`
	Etag              model.Etag      `json:"etag"`
	Name              *model.LangJson `json:"name,omitempty"`
	Description       *model.LangJson `json:"description,omitempty"`
	UnitId            *model.Id       `json:"unitId,omitempty"`
	Status            *string         `json:"status,omitempty"`
	DefaultsVariantId *model.Id       `json:"defaultsVariantId,omitempty"`
	ThumbnailUrl      *string         `json:"thumbnailUrl,omitempty"`
}

func (UpdateProductCommand) CqrsRequestType() cqrs.RequestType {
	return updateProductCommandType
}

type UpdateProductResult = GetProductByIdResult

// Delete

var deleteProductCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "product",
	Action:    "delete",
}

type DeleteProductCommand struct {
	Id model.Id `json:"id" param:"id"`
}

func (this DeleteProductCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}
	return val.ApiBased.ValidateStruct(&this, rules...)
}

func (DeleteProductCommand) CqrsRequestType() cqrs.RequestType {
	return deleteProductCommandType
}

type DeleteProductResult = crud.DeletionResult

// Get by ID

var getProductByIdQueryType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "product",
	Action:    "getById",
}

type GetProductByIdQuery struct {
	Id           model.Id `param:"id" json:"id"`
	WithVariants bool     `json:"withVariants,omitempty"`
}

func (this GetProductByIdQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}
	return val.ApiBased.ValidateStruct(&this, rules...)
}

func (GetProductByIdQuery) CqrsRequestType() cqrs.RequestType {
	return getProductByIdQueryType
}

type GetProductByIdResult = crud.OpResult[*domain.Product]

var searchProductsQueryType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "product",
	Action:    "search",
}

// Search (advanced)

type SearchProductsQuery struct {
	// Filled by service from Graph
	crud.SearchQuery

	WithVariants bool `json:"withVariants,omitempty"`
}

func (this SearchProductsQuery) CqrsRequestType() cqrs.RequestType {
	return searchProductsQueryType
}

func (this SearchProductsQuery) Validate() ft.ValidationErrors {
	rules := this.SearchQuery.ValidationRules()

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type SearchProductsResultData = crud.PagedResult[domain.Product]
type SearchProductsResult = crud.OpResult[*SearchProductsResultData]
