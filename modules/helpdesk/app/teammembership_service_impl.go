package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/domain"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/teammembership"
)

func NewTeamMembershipServiceImpl(repo it.TeamMembershipRepository, cqrsBus cqrs.CqrsBus) it.TeamMembershipService {
	return &TeamMembershipServiceImpl{cqrsBus: cqrsBus, repo: repo}
}

type TeamMembershipServiceImpl struct {
	cqrsBus cqrs.CqrsBus
	repo    it.TeamMembershipRepository
}

func (this *TeamMembershipServiceImpl) CreateTeamMembership(
	ctx corectx.Context, cmd it.CreateTeamMembershipCommand,
) (*it.CreateTeamMembershipResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[domain.TeamMembership, *domain.TeamMembership]{Action: "create teamMembership", BaseRepoGetter: this.repo, Data: cmd})
}

func (this *TeamMembershipServiceImpl) DeleteTeamMembership(
	ctx corectx.Context, cmd it.DeleteTeamMembershipCommand,
) (*it.DeleteTeamMembershipResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{Action: "delete teamMembership", DbRepoGetter: this.repo, Cmd: dyn.DeleteOneCommand(cmd)})
}

func (this *TeamMembershipServiceImpl) GetTeamMembership(
	ctx corectx.Context, query it.GetTeamMembershipQuery,
) (*it.GetTeamMembershipResult, error) {
	return corecrud.GetOne[domain.TeamMembership](ctx, corecrud.GetOneParam{Action: "get teamMembership", DbRepoGetter: this.repo, Query: dyn.GetOneQuery(query)})
}

func (this *TeamMembershipServiceImpl) TeamMembershipExists(
	ctx corectx.Context, query it.TeamMembershipExistsQuery,
) (*it.TeamMembershipExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{Action: "check if teamMembership exists", DbRepoGetter: this.repo, Query: dyn.ExistsQuery(query)})
}

func (this *TeamMembershipServiceImpl) SearchTeamMemberships(
	ctx corectx.Context, query it.SearchTeamMembershipsQuery,
) (*it.SearchTeamMembershipsResult, error) {
	return corecrud.Search[domain.TeamMembership](ctx, corecrud.SearchParam{Action: "search teamMemberships", DbRepoGetter: this.repo, Query: dyn.SearchQuery(query)})
}

func (this *TeamMembershipServiceImpl) UpdateTeamMembership(
	ctx corectx.Context, cmd it.UpdateTeamMembershipCommand,
) (*it.UpdateTeamMembershipResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.TeamMembership, *domain.TeamMembership]{Action: "update teamMembership", DbRepoGetter: this.repo, Data: cmd})
}
