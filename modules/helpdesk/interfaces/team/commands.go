package team

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/domain"
)

func init() {
	var req cqrs.Request
	req = (*CreateTeamCommand)(nil)
	req = (*DeleteTeamCommand)(nil)
	req = (*GetTeamQuery)(nil)
	req = (*TeamExistsQuery)(nil)
	req = (*SearchTeamsQuery)(nil)
	req = (*UpdateTeamCommand)(nil)
	req = (*SetTeamIsArchivedCommand)(nil)
	util.Unused(req)
}

var createTeamCommandType = cqrs.RequestType{Module: "helpdesk", Submodule: "team", Action: "createTeam"}

type CreateTeamCommand struct{ domain.Team }

func (CreateTeamCommand) CqrsRequestType() cqrs.RequestType { return createTeamCommandType }
func (CreateTeamCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.TeamSchemaName)
}

type CreateTeamResult = dyn.OpResult[domain.Team]

var deleteTeamCommandType = cqrs.RequestType{Module: "helpdesk", Submodule: "team", Action: "deleteTeam"}

type DeleteTeamCommand dyn.DeleteOneCommand

func (DeleteTeamCommand) CqrsRequestType() cqrs.RequestType { return deleteTeamCommandType }

type DeleteTeamResult = dyn.OpResult[dyn.MutateResultData]

var getTeamQueryType = cqrs.RequestType{Module: "helpdesk", Submodule: "team", Action: "getTeam"}

type GetTeamQuery dyn.GetOneQuery

func (GetTeamQuery) CqrsRequestType() cqrs.RequestType { return getTeamQueryType }

type GetTeamResult = dyn.OpResult[domain.Team]

var teamExistsQueryType = cqrs.RequestType{Module: "helpdesk", Submodule: "team", Action: "teamExists"}

type TeamExistsQuery dyn.ExistsQuery

func (TeamExistsQuery) CqrsRequestType() cqrs.RequestType { return teamExistsQueryType }

type TeamExistsResult = dyn.OpResult[dyn.ExistsResultData]

var searchTeamsQueryType = cqrs.RequestType{Module: "helpdesk", Submodule: "team", Action: "searchTeams"}

type SearchTeamsQuery dyn.SearchQuery

func (SearchTeamsQuery) CqrsRequestType() cqrs.RequestType { return searchTeamsQueryType }

type SearchTeamsResultData = dyn.PagedResultData[domain.Team]
type SearchTeamsResult = dyn.OpResult[SearchTeamsResultData]

var updateTeamCommandType = cqrs.RequestType{Module: "helpdesk", Submodule: "team", Action: "updateTeam"}

type UpdateTeamCommand struct{ domain.Team }

func (UpdateTeamCommand) CqrsRequestType() cqrs.RequestType { return updateTeamCommandType }
func (UpdateTeamCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.TeamSchemaName)
}

type UpdateTeamResult = dyn.OpResult[dyn.MutateResultData]

var setTeamIsArchivedCommandType = cqrs.RequestType{
	Module:    "helpdesk",
	Submodule: "team",
	Action:    "setTeamIsArchived",
}

type SetTeamIsArchivedCommand dyn.SetIsArchivedCommand

func (SetTeamIsArchivedCommand) CqrsRequestType() cqrs.RequestType {
	return setTeamIsArchivedCommandType
}

type SetTeamIsArchivedResult = dyn.OpResult[dyn.MutateResultData]
