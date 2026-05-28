package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/teammembership"
)

func NewTeamMembershipApplicationServiceImpl(teamMembershipSvc it.TeamMembershipDomainService) it.TeamMembershipAppService {
	return &TeamMembershipApplicationServiceImpl{teamMembershipSvc: teamMembershipSvc}
}

type TeamMembershipApplicationServiceImpl struct {
	teamMembershipSvc it.TeamMembershipDomainService
}

func (this *TeamMembershipApplicationServiceImpl) CreateTeamMembership(ctx corectx.Context, cmd it.CreateTeamMembershipCommand) (*it.CreateTeamMembershipResult, error) {
	return this.teamMembershipSvc.CreateTeamMembership(ctx, cmd)
}

func (this *TeamMembershipApplicationServiceImpl) DeleteTeamMembership(ctx corectx.Context, cmd it.DeleteTeamMembershipCommand) (*it.DeleteTeamMembershipResult, error) {
	return this.teamMembershipSvc.DeleteTeamMembership(ctx, cmd)
}

func (this *TeamMembershipApplicationServiceImpl) GetTeamMembership(ctx corectx.Context, query it.GetTeamMembershipQuery) (*it.GetTeamMembershipResult, error) {
	return this.teamMembershipSvc.GetTeamMembership(ctx, query)
}

func (this *TeamMembershipApplicationServiceImpl) TeamMembershipExists(ctx corectx.Context, query it.TeamMembershipExistsQuery) (*it.TeamMembershipExistsResult, error) {
	return this.teamMembershipSvc.TeamMembershipExists(ctx, query)
}

func (this *TeamMembershipApplicationServiceImpl) SearchTeamMemberships(ctx corectx.Context, query it.SearchTeamMembershipsQuery) (*it.SearchTeamMembershipsResult, error) {
	return this.teamMembershipSvc.SearchTeamMemberships(ctx, query)
}

func (this *TeamMembershipApplicationServiceImpl) UpdateTeamMembership(ctx corectx.Context, cmd it.UpdateTeamMembershipCommand) (*it.UpdateTeamMembershipResult, error) {
	return this.teamMembershipSvc.UpdateTeamMembership(ctx, cmd)
}
