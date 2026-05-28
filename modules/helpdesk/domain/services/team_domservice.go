package services

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/team"
)

func NewTeamDomainServiceImpl(repo it.TeamRepository, cqrsBus cqrs.CqrsBus) it.TeamDomainService {
	return &TeamDomainServiceImpl{cqrsBus: cqrsBus, repo: repo}
}

type TeamDomainServiceImpl struct {
	cqrsBus cqrs.CqrsBus
	repo    it.TeamRepository
}

func (this *TeamDomainServiceImpl) CreateTeam(
	ctx corectx.Context, cmd it.CreateTeamCommand,
) (*it.CreateTeamResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[models.Team, *models.Team]{Action: "create team", BaseRepoGetter: this.repo, Data: cmd})
}

func (this *TeamDomainServiceImpl) DeleteTeam(
	ctx corectx.Context, cmd it.DeleteTeamCommand,
) (*it.DeleteTeamResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{Action: "delete team", DbRepoGetter: this.repo, Cmd: dyn.DeleteOneCommand(cmd)})
}

func (this *TeamDomainServiceImpl) GetTeam(
	ctx corectx.Context, query it.GetTeamQuery,
) (*it.GetTeamResult, error) {
	return corecrud.GetOne[models.Team](ctx, corecrud.GetOneParam{Action: "get team", DbRepoGetter: this.repo, Query: dyn.GetOneQuery(query)})
}

func (this *TeamDomainServiceImpl) TeamExists(
	ctx corectx.Context, query it.TeamExistsQuery,
) (*it.TeamExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{Action: "check if team exists", DbRepoGetter: this.repo, Query: dyn.ExistsQuery(query)})
}

func (this *TeamDomainServiceImpl) SearchTeams(
	ctx corectx.Context, query it.SearchTeamsQuery,
) (*it.SearchTeamsResult, error) {
	return corecrud.Search[models.Team](ctx, corecrud.SearchParam{Action: "search teams", DbRepoGetter: this.repo, Query: dyn.SearchQuery(query)})
}

func (this *TeamDomainServiceImpl) UpdateTeam(
	ctx corectx.Context, cmd it.UpdateTeamCommand,
) (*it.UpdateTeamResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[models.Team, *models.Team]{Action: "update team", DbRepoGetter: this.repo, Data: cmd})
}

func (this *TeamDomainServiceImpl) SetTeamIsArchived(
	ctx corectx.Context, cmd it.SetTeamIsArchivedCommand,
) (*it.SetTeamIsArchivedResult, error) {
	return corecrud.SetIsArchived(ctx, this.repo, dyn.SetIsArchivedCommand(cmd))
}
