package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/slapolicy"
)

type CreateSlaPolicyRequest = it.CreateSlaPolicyCommand
type CreateSlaPolicyResponse = httpserver.RestCreateResponse
type DeleteSlaPolicyRequest = it.DeleteSlaPolicyCommand
type DeleteSlaPolicyResponse = httpserver.RestDeleteResponse2
type GetSlaPolicyRequest = it.GetSlaPolicyQuery
type GetSlaPolicyResponse = dmodel.DynamicFields
type SlaPolicyExistsRequest = it.SlaPolicyExistsQuery
type SlaPolicyExistsResponse = dyn.ExistsResultData
type SearchSlaPoliciesRequest = it.SearchSlaPoliciesQuery
type SearchSlaPoliciesResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
type UpdateSlaPolicyRequest = it.UpdateSlaPolicyCommand
type UpdateSlaPolicyResponse = httpserver.RestMutateResponse
type SetSlaPolicyIsArchivedRequest = it.SetSlaPolicyIsArchivedCommand
type SetSlaPolicyIsArchivedResponse = httpserver.RestMutateResponse
