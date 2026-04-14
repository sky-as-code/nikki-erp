package product

import (
	"github.com/shopspring/decimal"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
)

func init() {
	var req cqrs.Request
	req = (*CreateProductCommand)(nil)
	req = (*DeleteProductCommand)(nil)
	req = (*GetProductQuery)(nil)
	req = (*SearchProductsQuery)(nil)
	req = (*SetProductIsArchivedCommand)(nil)
	req = (*UpdateProductCommand)(nil)
	req = (*ProductExistsQuery)(nil)
	util.Unused(req)
}

var createProductCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "product",
	Action:    "create",
}

type CreateProductCommand struct {
	domain.Product

	Sku           string          `json:"sku"`
	BarCode       string          `json:"barcode"`
	ProposedPrice decimal.Decimal `json:"proposed_price"`
}

func (CreateProductCommand) CqrsRequestType() cqrs.RequestType {
	return createProductCommandType
}

func (this CreateProductCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.ProductSchemaName)
}

type CreateProductResult = dyn.OpResult[domain.Product]

var deleteProductCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "product",
	Action:    "delete",
}

type DeleteProductCommand dyn.DeleteOneCommand

func (DeleteProductCommand) CqrsRequestType() cqrs.RequestType {
	return deleteProductCommandType
}

type DeleteProductResult = dyn.OpResult[dyn.MutateResultData]

var getProductQueryType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "product",
	Action:    "getProduct",
}

type GetProductQuery dyn.GetOneQuery

func (GetProductQuery) CqrsRequestType() cqrs.RequestType {
	return getProductQueryType
}

type GetProductResult = dyn.OpResult[domain.Product]

var searchProductsQueryType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "product",
	Action:    "search",
}

type SearchProductsQuery dyn.SearchQuery

func (SearchProductsQuery) CqrsRequestType() cqrs.RequestType {
	return searchProductsQueryType
}

type SearchProductsResultData = dyn.PagedResultData[domain.Product]
type SearchProductsResult = dyn.OpResult[SearchProductsResultData]

var setProductIsArchivedCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "product",
	Action:    "setProductIsArchived",
}

type SetProductIsArchivedCommand dyn.SetIsArchivedCommand

func (SetProductIsArchivedCommand) CqrsRequestType() cqrs.RequestType {
	return setProductIsArchivedCommandType
}

type SetProductIsArchivedResult = dyn.OpResult[dyn.MutateResultData]

var productExistsQueryType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "product",
	Action:    "exists",
}

type ProductExistsQuery dyn.ExistsQuery

func (ProductExistsQuery) CqrsRequestType() cqrs.RequestType {
	return productExistsQueryType
}

type ProductExistsResult = dyn.OpResult[dyn.ExistsResultData]

var updateProductCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "product",
	Action:    "update",
}

type UpdateProductCommand struct {
	domain.Product
}

func (UpdateProductCommand) CqrsRequestType() cqrs.RequestType {
	return updateProductCommandType
}

func (this UpdateProductCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.ProductSchemaName)
}

type UpdateProductResult = dyn.OpResult[dyn.MutateResultData]
