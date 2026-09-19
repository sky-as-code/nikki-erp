package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

type GetStockReservationRequest = itStock.GetStockReservationByIdQuery
type GetStockReservationResponse = httpserver.RestGetOneResponse[dmodel.DynamicFields]

type StockReservationExistsRequest = itStock.StockReservationExistsQuery
type StockReservationExistsResponse = dyn.ExistsResultData

type SearchStockReservationsRequest = itStock.SearchStockReservationsQuery
type SearchStockReservationsResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
