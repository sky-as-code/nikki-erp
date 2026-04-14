package productcategory

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
)

func init() {
	var req cqrs.Request
	req = (*CreateProductCategoryCommand)(nil)
	req = (*DeleteProductCategoryCommand)(nil)
	req = (*ProductCategoryExistsQuery)(nil)
	req = (*GetProductCategoryQuery)(nil)
	req = (*SearchProductCategoriesQuery)(nil)
	req = (*UpdateProductCategoryCommand)(nil)
	util.Unused(req)
}

var createProductCategoryCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "product_category",
	Action:    "create",
}

type CreateProductCategoryCommand struct {
	domain.ProductCategory
}

func (CreateProductCategoryCommand) CqrsRequestType() cqrs.RequestType {
	return createProductCategoryCommandType
}

func (this CreateProductCategoryCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.ProductCategorySchemaName)
}

type CreateProductCategoryResult = dyn.OpResult[domain.ProductCategory]

var deleteProductCategoryCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "product_category",
	Action:    "delete",
}

type DeleteProductCategoryCommand dyn.DeleteOneCommand

func (DeleteProductCategoryCommand) CqrsRequestType() cqrs.RequestType {
	return deleteProductCategoryCommandType
}

type DeleteProductCategoryResult = dyn.OpResult[dyn.MutateResultData]

var getProductCategoryQueryType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "product_category",
	Action:    "getProductCategory",
}

type GetProductCategoryQuery dyn.GetOneQuery

func (GetProductCategoryQuery) CqrsRequestType() cqrs.RequestType {
	return getProductCategoryQueryType
}

type GetProductCategoryResult = dyn.OpResult[domain.ProductCategory]

var productCategoryExistsQueryType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "product_category",
	Action:    "productCategoryExists",
}

type ProductCategoryExistsQuery dyn.ExistsQuery

func (ProductCategoryExistsQuery) CqrsRequestType() cqrs.RequestType {
	return productCategoryExistsQueryType
}

type ProductCategoryExistsResult = dyn.OpResult[dyn.ExistsResultData]

var searchProductCategoriesQueryType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "product_category",
	Action:    "search",
}

type SearchProductCategoriesQuery dyn.SearchQuery

func (SearchProductCategoriesQuery) CqrsRequestType() cqrs.RequestType {
	return searchProductCategoriesQueryType
}

type SearchProductCategoriesResultData = dyn.PagedResultData[domain.ProductCategory]
type SearchProductCategoriesResult = dyn.OpResult[SearchProductCategoriesResultData]

var updateProductCategoryCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "product_category",
	Action:    "update",
}

type UpdateProductCategoryCommand struct {
	domain.ProductCategory
}

func (UpdateProductCategoryCommand) CqrsRequestType() cqrs.RequestType {
	return updateProductCategoryCommandType
}

func (this UpdateProductCategoryCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.ProductCategorySchemaName)
}

type UpdateProductCategoryResult = dyn.OpResult[dyn.MutateResultData]
