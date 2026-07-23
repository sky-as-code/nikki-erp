package v1

import (
	"mime/multipart"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	itVariant "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/variant"
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

type UploadVariantImageRequest struct {
	ProductId  string                `param:"product_id"`
	VariantId  string                `param:"id"`
	FileHeader *multipart.FileHeader `json:"-" form-file:"file"`
}

type UploadVariantImageResponse = httpserver.RestMutateResponse
