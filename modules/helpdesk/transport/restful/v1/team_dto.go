package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/team"
)

type CreateTeamRequest = it.CreateTeamCommand
type CreateTeamResponse = httpserver.RestCreateResponse
type DeleteTeamRequest = it.DeleteTeamCommand
type DeleteTeamResponse = httpserver.RestDeleteResponse2
type GetTeamRequest = it.GetTeamQuery
type GetTeamResponse = dmodel.DynamicFields
type TeamExistsRequest = it.TeamExistsQuery
type TeamExistsResponse = dyn.ExistsResultData
type SearchTeamsRequest = it.SearchTeamsQuery
type SearchTeamsResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
type UpdateTeamRequest = it.UpdateTeamCommand
type UpdateTeamResponse = httpserver.RestMutateResponse
type SetTeamIsArchivedRequest = it.SetTeamIsArchivedCommand
type SetTeamIsArchivedResponse = httpserver.RestMutateResponse
