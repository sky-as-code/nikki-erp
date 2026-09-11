package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

type CreateStockMoveDependencyRequest = itStock.CreateStockMoveDependencyCommand
type CreateStockMoveDependencyResponse = httpserver.RestCreateResponse

type UpdateStockMoveDependencyRequest = itStock.UpdateStockMoveDependencyCommand
type UpdateStockMoveDependencyResponse = httpserver.RestMutateResponse

type DeleteStockMoveDependencyRequest = itStock.DeleteStockMoveDependencyCommand
type DeleteStockMoveDependencyResponse = httpserver.RestMutateResponse

type SetStockMoveDependencyArchivedRequest = itStock.SetStockMoveDependencyArchivedCommand
type SetStockMoveDependencyArchivedResponse = httpserver.RestMutateResponse

type GetStockMoveDependencyRequest = itStock.GetStockMoveDependencyByIdQuery
type GetStockMoveDependencyResponse = httpserver.RestGetOneResponse[dmodel.DynamicFields]

type StockMoveDependencyExistsRequest = itStock.StockMoveDependencyExistsQuery
type StockMoveDependencyExistsResponse = dyn.ExistsResultData

type SearchStockMoveDependenciesRequest = itStock.SearchStockMoveDependenciesQuery
type SearchStockMoveDependenciesResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
