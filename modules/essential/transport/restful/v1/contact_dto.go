package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/contact"
)

type CreateContactRequest struct {
	dmodel.DynamicFields
}
type CreateContactResponse = httpserver.RestCreateResponse

type DeleteContactRequest = it.DeleteContactCommand
type DeleteContactResponse = httpserver.RestDeleteResponse2

type GetContactRequest = it.GetContactQuery
type GetContactResponse = dmodel.DynamicFields

type ContactExistsRequest = it.ContactExistsQuery
type ContactExistsResponse = dyn.ExistsResultData

type SearchContactsRequest = it.SearchContactsQuery
type SearchContactsResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]

type UpdateContactRequest struct {
	dmodel.DynamicFields
	ContactId string `param:"id"`
}
type UpdateContactResponse = httpserver.RestMutateResponse
