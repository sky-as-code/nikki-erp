package services

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/teammembership"
)

func NewTeamMembershipDomainServiceImpl(repo it.TeamMembershipRepository, cqrsBus cqrs.CqrsBus) it.TeamMembershipDomainService {
	return &TeamMembershipDomainServiceImpl{cqrsBus: cqrsBus, repo: repo}
}

type TeamMembershipDomainServiceImpl struct {
	cqrsBus cqrs.CqrsBus
	repo    it.TeamMembershipRepository
}

func (this *TeamMembershipDomainServiceImpl) CreateTeamMembership(
	ctx corectx.Context, cmd it.CreateTeamMembershipCommand,
) (*it.CreateTeamMembershipResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[models.TeamMembership, *models.TeamMembership]{Action: "create teamMembership", BaseRepoGetter: this.repo, Data: cmd})
}

func (this *TeamMembershipDomainServiceImpl) DeleteTeamMembership(
	ctx corectx.Context, cmd it.DeleteTeamMembershipCommand,
) (*it.DeleteTeamMembershipResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{Action: "delete teamMembership", DbRepoGetter: this.repo, Cmd: dyn.DeleteOneCommand(cmd)})
}

func (this *TeamMembershipDomainServiceImpl) GetTeamMembership(
	ctx corectx.Context, query it.GetTeamMembershipQuery,
) (*it.GetTeamMembershipResult, error) {
	return corecrud.GetOne[models.TeamMembership](ctx, corecrud.GetOneParam{Action: "get teamMembership", DbRepoGetter: this.repo, Query: dyn.GetOneQuery(query)})
}

func (this *TeamMembershipDomainServiceImpl) TeamMembershipExists(
	ctx corectx.Context, query it.TeamMembershipExistsQuery,
) (*it.TeamMembershipExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{Action: "check if teamMembership exists", DbRepoGetter: this.repo, Query: dyn.ExistsQuery(query)})
}

func (this *TeamMembershipDomainServiceImpl) SearchTeamMemberships(
	ctx corectx.Context, query it.SearchTeamMembershipsQuery,
) (*it.SearchTeamMembershipsResult, error) {
	return corecrud.Search[models.TeamMembership](ctx, corecrud.SearchParam{Action: "search teamMemberships", DbRepoGetter: this.repo, Query: dyn.SearchQuery(query)})
}

func (this *TeamMembershipDomainServiceImpl) UpdateTeamMembership(
	ctx corectx.Context, cmd it.UpdateTeamMembershipCommand,
) (*it.UpdateTeamMembershipResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[models.TeamMembership, *models.TeamMembership]{Action: "update teamMembership", DbRepoGetter: this.repo, Data: cmd})
}
