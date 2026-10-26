package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	itVariant "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/legacy/interfaces/variant"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
)

type CreateVariantRequest = itVariant.CreateVariantCommand
type CreateVariantResponse = httpserver.RestCreateResponse

type UpdateVariantRequest = itVariant.UpdateVariantCommand
type UpdateVariantResponse = httpserver.RestMutateResponse

type DeleteVariantRequest = itVariant.DeleteVariantCommand
type DeleteVariantResponse = httpserver.RestDeleteResponse2

type GetVariantRequest = itVariant.GetVariantQuery
type GetVariantResponse = dmodel.DynamicFields

type SearchVariantsRequest = itVariant.SearchVariantsQuery
type SearchVariantsResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
