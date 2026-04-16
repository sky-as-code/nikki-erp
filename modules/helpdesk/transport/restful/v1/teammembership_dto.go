package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/teammembership"
)

type CreateTeamMembershipRequest = it.CreateTeamMembershipCommand
type CreateTeamMembershipResponse = httpserver.RestCreateResponse
type DeleteTeamMembershipRequest = it.DeleteTeamMembershipCommand
type DeleteTeamMembershipResponse = httpserver.RestDeleteResponse2
type GetTeamMembershipRequest = it.GetTeamMembershipQuery
type GetTeamMembershipResponse = dmodel.DynamicFields
type TeamMembershipExistsRequest = it.TeamMembershipExistsQuery
type TeamMembershipExistsResponse = dyn.ExistsResultData
type SearchTeamMembershipsRequest = it.SearchTeamMembershipsQuery
type SearchTeamMembershipsResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
type UpdateTeamMembershipRequest = it.UpdateTeamMembershipCommand
type UpdateTeamMembershipResponse = httpserver.RestMutateResponse
