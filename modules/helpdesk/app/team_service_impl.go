package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/domain"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/team"
)

func NewTeamServiceImpl(repo it.TeamRepository, cqrsBus cqrs.CqrsBus) it.TeamService {
	return &TeamServiceImpl{cqrsBus: cqrsBus, repo: repo}
}

type TeamServiceImpl struct {
	cqrsBus cqrs.CqrsBus
	repo    it.TeamRepository
}

func (this *TeamServiceImpl) CreateTeam(
	ctx corectx.Context, cmd it.CreateTeamCommand,
) (*it.CreateTeamResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[domain.Team, *domain.Team]{Action: "create team", BaseRepoGetter: this.repo, Data: cmd})
}

func (this *TeamServiceImpl) DeleteTeam(
	ctx corectx.Context, cmd it.DeleteTeamCommand,
) (*it.DeleteTeamResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{Action: "delete team", DbRepoGetter: this.repo, Cmd: dyn.DeleteOneCommand(cmd)})
}

func (this *TeamServiceImpl) GetTeam(
	ctx corectx.Context, query it.GetTeamQuery,
) (*it.GetTeamResult, error) {
	return corecrud.GetOne[domain.Team](ctx, corecrud.GetOneParam{Action: "get team", DbRepoGetter: this.repo, Query: dyn.GetOneQuery(query)})
}

func (this *TeamServiceImpl) TeamExists(
	ctx corectx.Context, query it.TeamExistsQuery,
) (*it.TeamExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{Action: "check if team exists", DbRepoGetter: this.repo, Query: dyn.ExistsQuery(query)})
}

func (this *TeamServiceImpl) SearchTeams(
	ctx corectx.Context, query it.SearchTeamsQuery,
) (*it.SearchTeamsResult, error) {
	return corecrud.Search[domain.Team](ctx, corecrud.SearchParam{Action: "search teams", DbRepoGetter: this.repo, Query: dyn.SearchQuery(query)})
}

func (this *TeamServiceImpl) UpdateTeam(
	ctx corectx.Context, cmd it.UpdateTeamCommand,
) (*it.UpdateTeamResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.Team, *domain.Team]{Action: "update team", DbRepoGetter: this.repo, Data: cmd})
}

func (this *TeamServiceImpl) SetTeamIsArchived(
	ctx corectx.Context, cmd it.SetTeamIsArchivedCommand,
) (*it.SetTeamIsArchivedResult, error) {
	return corecrud.SetIsArchived(ctx, this.repo, dyn.SetIsArchivedCommand(cmd))
}
