package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	itUnit "github.com/sky-as-code/nikki-erp/modules/inventory/unit/interfaces/unit"
)

type CreateUnitRequest = itUnit.CreateUnitCommand
type CreateUnitResponse = httpserver.RestCreateResponse

type UpdateUnitRequest = itUnit.UpdateUnitCommand
type UpdateUnitResponse = httpserver.RestMutateResponse

type DeleteUnitRequest = itUnit.DeleteUnitCommand
type DeleteUnitResponse = httpserver.RestDeleteResponse2

type GetUnitRequest = itUnit.GetUnitQuery
type GetUnitResponse = dmodel.DynamicFields

type SearchUnitsRequest = itUnit.SearchUnitsQuery
type SearchUnitsResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
