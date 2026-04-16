package teammembership

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/domain"
)

func init() {
	var req cqrs.Request
	req = (*CreateTeamMembershipCommand)(nil)
	req = (*DeleteTeamMembershipCommand)(nil)
	req = (*GetTeamMembershipQuery)(nil)
	req = (*TeamMembershipExistsQuery)(nil)
	req = (*SearchTeamMembershipsQuery)(nil)
	req = (*UpdateTeamMembershipCommand)(nil)
	util.Unused(req)
}

var createTeamMembershipCommandType = cqrs.RequestType{Module: "helpdesk", Submodule: "teammembership", Action: "createTeamMembership"}

type CreateTeamMembershipCommand struct{ domain.TeamMembership }

func (CreateTeamMembershipCommand) CqrsRequestType() cqrs.RequestType {
	return createTeamMembershipCommandType
}
func (CreateTeamMembershipCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.TeamMembershipSchemaName)
}

type CreateTeamMembershipResult = dyn.OpResult[domain.TeamMembership]

var deleteTeamMembershipCommandType = cqrs.RequestType{Module: "helpdesk", Submodule: "teammembership", Action: "deleteTeamMembership"}

type DeleteTeamMembershipCommand dyn.DeleteOneCommand

func (DeleteTeamMembershipCommand) CqrsRequestType() cqrs.RequestType {
	return deleteTeamMembershipCommandType
}

type DeleteTeamMembershipResult = dyn.OpResult[dyn.MutateResultData]

var getTeamMembershipQueryType = cqrs.RequestType{Module: "helpdesk", Submodule: "teammembership", Action: "getTeamMembership"}

type GetTeamMembershipQuery dyn.GetOneQuery

func (GetTeamMembershipQuery) CqrsRequestType() cqrs.RequestType { return getTeamMembershipQueryType }

type GetTeamMembershipResult = dyn.OpResult[domain.TeamMembership]

var teamMembershipExistsQueryType = cqrs.RequestType{Module: "helpdesk", Submodule: "teammembership", Action: "teamMembershipExists"}

type TeamMembershipExistsQuery dyn.ExistsQuery

func (TeamMembershipExistsQuery) CqrsRequestType() cqrs.RequestType {
	return teamMembershipExistsQueryType
}

type TeamMembershipExistsResult = dyn.OpResult[dyn.ExistsResultData]

var searchTeamMembershipsQueryType = cqrs.RequestType{Module: "helpdesk", Submodule: "teammembership", Action: "searchTeamMemberships"}

type SearchTeamMembershipsQuery dyn.SearchQuery

func (SearchTeamMembershipsQuery) CqrsRequestType() cqrs.RequestType {
	return searchTeamMembershipsQueryType
}

type SearchTeamMembershipsResultData = dyn.PagedResultData[domain.TeamMembership]
type SearchTeamMembershipsResult = dyn.OpResult[SearchTeamMembershipsResultData]

var updateTeamMembershipCommandType = cqrs.RequestType{Module: "helpdesk", Submodule: "teammembership", Action: "updateTeamMembership"}

type UpdateTeamMembershipCommand struct{ domain.TeamMembership }

func (UpdateTeamMembershipCommand) CqrsRequestType() cqrs.RequestType {
	return updateTeamMembershipCommandType
}
func (UpdateTeamMembershipCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.TeamMembershipSchemaName)
}

type UpdateTeamMembershipResult = dyn.OpResult[dyn.MutateResultData]
