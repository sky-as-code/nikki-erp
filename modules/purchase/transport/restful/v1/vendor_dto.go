package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/vendor"
)

type CreateVendorRequest struct{ dmodel.DynamicFields }
type CreateVendorResponse = httpserver.RestCreateResponse
type DeleteVendorRequest = it.DeleteVendorCommand
type DeleteVendorResponse = httpserver.RestDeleteResponse2
type GetVendorRequest = it.GetVendorQuery
type GetVendorResponse = dmodel.DynamicFields
type VendorExistsRequest = it.VendorExistsQuery
type VendorExistsResponse = dyn.ExistsResultData
type SearchVendorsRequest = it.SearchVendorsQuery
type SearchVendorsResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
type SetVendorIsArchivedRequest = it.SetVendorIsArchivedCommand
type SetVendorIsArchivedResponse = httpserver.RestMutateResponse
type UpdateVendorRequest struct {
	dmodel.DynamicFields
	VendorId string `param:"id"`
}
type UpdateVendorResponse = httpserver.RestMutateResponse
