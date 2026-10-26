package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

type CreateStockTransferRequest = itStock.CreateStockTransferCommand
type CreateStockTransferResponse = httpserver.RestCreateResponse

type UpdateStockTransferRequest = itStock.UpdateStockTransferCommand
type UpdateStockTransferResponse = httpserver.RestMutateResponse

type DeleteStockTransferRequest = itStock.DeleteStockTransferCommand
type DeleteStockTransferResponse = httpserver.RestMutateResponse

type SetStockTransferArchivedRequest = itStock.SetStockTransferArchivedCommand
type SetStockTransferArchivedResponse = httpserver.RestMutateResponse

type GetStockTransferRequest = itStock.GetStockTransferByIdQuery
type GetStockTransferResponse = httpserver.RestGetOneResponse[dmodel.DynamicFields]

type StockTransferExistsRequest = itStock.StockTransferExistsQuery
type StockTransferExistsResponse = dyn.ExistsResultData

type SearchStockTransfersRequest = itStock.SearchStockTransfersQuery
type SearchStockTransfersResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
