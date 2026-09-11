package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/currency"
)

type CreateCurrencyRequest = it.CreateCurrencyCommand
type CreateCurrencyResponse = httpserver.RestCreateResponse

type UpdateCurrencyRequest = it.UpdateCurrencyCommand
type UpdateCurrencyResponse = httpserver.RestMutateResponse

type DeleteCurrencyRequest = it.DeleteCurrencyCommand
type DeleteCurrencyResponse = httpserver.RestMutateResponse

type SetCurrencyArchivedRequest = it.SetCurrencyArchivedCommand
type SetCurrencyArchivedResponse = httpserver.RestMutateResponse

type GetCurrencyRequest = it.GetCurrencyByIdQuery
type GetCurrencyResponse = httpserver.RestGetOneResponse[dmodel.DynamicFields]

type CurrencyExistsRequest = it.CurrencyExistsQuery
type CurrencyExistsResponse = dyn.ExistsResultData

type SearchCurrenciesRequest = it.SearchCurrenciesQuery
type SearchCurrenciesResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
