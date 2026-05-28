package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/team"
)

func NewTeamApplicationServiceImpl(teamSvc it.TeamDomainService) it.TeamAppService {
	return &TeamApplicationServiceImpl{teamSvc: teamSvc}
}

type TeamApplicationServiceImpl struct {
	teamSvc it.TeamDomainService
}

func (this *TeamApplicationServiceImpl) CreateTeam(ctx corectx.Context, cmd it.CreateTeamCommand) (*it.CreateTeamResult, error) {
	return this.teamSvc.CreateTeam(ctx, cmd)
}

func (this *TeamApplicationServiceImpl) DeleteTeam(ctx corectx.Context, cmd it.DeleteTeamCommand) (*it.DeleteTeamResult, error) {
	return this.teamSvc.DeleteTeam(ctx, cmd)
}

func (this *TeamApplicationServiceImpl) GetTeam(ctx corectx.Context, query it.GetTeamQuery) (*it.GetTeamResult, error) {
	return this.teamSvc.GetTeam(ctx, query)
}

func (this *TeamApplicationServiceImpl) TeamExists(ctx corectx.Context, query it.TeamExistsQuery) (*it.TeamExistsResult, error) {
	return this.teamSvc.TeamExists(ctx, query)
}

func (this *TeamApplicationServiceImpl) SearchTeams(ctx corectx.Context, query it.SearchTeamsQuery) (*it.SearchTeamsResult, error) {
	return this.teamSvc.SearchTeams(ctx, query)
}

func (this *TeamApplicationServiceImpl) UpdateTeam(ctx corectx.Context, cmd it.UpdateTeamCommand) (*it.UpdateTeamResult, error) {
	return this.teamSvc.UpdateTeam(ctx, cmd)
}

func (this *TeamApplicationServiceImpl) SetTeamIsArchived(ctx corectx.Context, cmd it.SetTeamIsArchivedCommand) (*it.SetTeamIsArchivedResult, error) {
	return this.teamSvc.SetTeamIsArchived(ctx, cmd)
}
