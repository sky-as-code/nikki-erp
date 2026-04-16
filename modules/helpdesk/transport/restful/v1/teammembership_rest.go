package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/teammembership"
)

type teamMembershipRestParams struct {
	dig.In
	Service it.TeamMembershipService
}

func NewTeamMembershipRest(params teamMembershipRestParams) *TeamMembershipRest {
	return &TeamMembershipRest{Service: params.Service}
}

type TeamMembershipRest struct {
	httpserver.RestBase
	Service it.TeamMembershipService
}

func (this TeamMembershipRest) CreateTeamMembership(echoCtx *echo.Context) (err error) {
	return httpserver.ServeCreate("create teamMembership", echoCtx, &it.CreateTeamMembershipCommand{}, this.Service.CreateTeamMembership)
}
func (this TeamMembershipRest) DeleteTeamMembership(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate("delete teamMembership", echoCtx, this.Service.DeleteTeamMembership)
}
func (this TeamMembershipRest) GetTeamMembership(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGetOne("get teamMembership", echoCtx, this.Service.GetTeamMembership)
}
func (this TeamMembershipRest) TeamMembershipExists(echoCtx *echo.Context) (err error) {
	return httpserver.ServeExists("teamMembership exists", echoCtx, this.Service.TeamMembershipExists)
}
func (this TeamMembershipRest) SearchTeamMemberships(echoCtx *echo.Context) (err error) {
	return httpserver.ServeSearch("search teamMemberships", echoCtx, this.Service.SearchTeamMemberships, true)
}
func (this TeamMembershipRest) UpdateTeamMembership(echoCtx *echo.Context) (err error) {
	return httpserver.ServeUpdate("update teamMembership", echoCtx, &it.UpdateTeamMembershipCommand{}, this.Service.UpdateTeamMembership)
}
