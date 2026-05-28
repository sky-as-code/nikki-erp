package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/party"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
)

type CreatePartyRequest = party.CreatePartyCommand
type CreatePartyResponse = httpserver.RestCreateResponse

type UpdatePartyRequest = party.UpdatePartyCommand
type UpdatePartyResponse = httpserver.RestMutateResponse

type DeletePartyRequest = party.DeletePartyCommand
type DeletePartyResponse = httpserver.RestMutateResponse

type GetPartyRequest = party.GetPartyQuery
type GetPartyResponse = dmodel.DynamicFields

type PartyExistsRequest = party.PartyExistsQuery
type PartyExistsResponse = dynamicmodel.ExistsResultData

type SearchPartiesRequest = party.SearchPartiesQuery
type SearchPartiesResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
