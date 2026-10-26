package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

type CreateStockMoveLineRequest = itStock.CreateStockMoveLineCommand
type CreateStockMoveLineResponse = httpserver.RestCreateResponse

type UpdateStockMoveLineRequest = itStock.UpdateStockMoveLineCommand
type UpdateStockMoveLineResponse = httpserver.RestMutateResponse

type DeleteStockMoveLineRequest = itStock.DeleteStockMoveLineCommand
type DeleteStockMoveLineResponse = httpserver.RestMutateResponse

type SetStockMoveLineArchivedRequest = itStock.SetStockMoveLineArchivedCommand
type SetStockMoveLineArchivedResponse = httpserver.RestMutateResponse

type GetStockMoveLineRequest = itStock.GetStockMoveLineByIdQuery
type GetStockMoveLineResponse = httpserver.RestGetOneResponse[dmodel.DynamicFields]

type StockMoveLineExistsRequest = itStock.StockMoveLineExistsQuery
type StockMoveLineExistsResponse = dyn.ExistsResultData

type SearchStockMoveLinesRequest = itStock.SearchStockMoveLinesQuery
type SearchStockMoveLinesResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
