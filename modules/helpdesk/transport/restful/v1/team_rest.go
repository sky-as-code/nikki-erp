package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/team"
)

type teamRestParams struct {
	dig.In
	Service it.TeamService
}

func NewTeamRest(params teamRestParams) *TeamRest { return &TeamRest{Service: params.Service} }

type TeamRest struct {
	httpserver.RestBase
	Service it.TeamService
}

func (this TeamRest) CreateTeam(echoCtx *echo.Context) (err error) {
	return httpserver.ServeCreate("create team", echoCtx, &it.CreateTeamCommand{}, this.Service.CreateTeam)
}
func (this TeamRest) DeleteTeam(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate("delete team", echoCtx, this.Service.DeleteTeam)
}
func (this TeamRest) GetTeam(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGetOne("get team", echoCtx, this.Service.GetTeam)
}
func (this TeamRest) TeamExists(echoCtx *echo.Context) (err error) {
	return httpserver.ServeExists("team exists", echoCtx, this.Service.TeamExists)
}
func (this TeamRest) SearchTeams(echoCtx *echo.Context) (err error) {
	return httpserver.ServeSearch("search teams", echoCtx, this.Service.SearchTeams)
}
func (this TeamRest) UpdateTeam(echoCtx *echo.Context) (err error) {
	return httpserver.ServeUpdate("update team", echoCtx, &it.UpdateTeamCommand{}, this.Service.UpdateTeam)
}
func (this TeamRest) SetTeamIsArchived(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate("set team is_archived", echoCtx, this.Service.SetTeamIsArchived)
}
