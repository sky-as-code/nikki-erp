package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/ticketactivity"
)

type ticketActivityRestParams struct {
	dig.In
	Service it.TicketActivityService
}

func NewTicketActivityRest(params ticketActivityRestParams) *TicketActivityRest {
	return &TicketActivityRest{Service: params.Service}
}

type TicketActivityRest struct {
	httpserver.RestBase
	Service it.TicketActivityService
}

func (this TicketActivityRest) CreateTicketActivity(echoCtx *echo.Context) (err error) {
	return httpserver.ServeCreate("create ticketActivity", echoCtx, &it.CreateTicketActivityCommand{}, this.Service.CreateTicketActivity)
}
func (this TicketActivityRest) DeleteTicketActivity(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate("delete ticketActivity", echoCtx, this.Service.DeleteTicketActivity)
}
func (this TicketActivityRest) GetTicketActivity(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGetOne("get ticketActivity", echoCtx, this.Service.GetTicketActivity)
}
func (this TicketActivityRest) TicketActivityExists(echoCtx *echo.Context) (err error) {
	return httpserver.ServeExists("ticketActivity exists", echoCtx, this.Service.TicketActivityExists)
}
func (this TicketActivityRest) SearchTicketActivities(echoCtx *echo.Context) (err error) {
	return httpserver.ServeSearch("search ticketActivitys", echoCtx, this.Service.SearchTicketActivities)
}
func (this TicketActivityRest) UpdateTicketActivity(echoCtx *echo.Context) (err error) {
	return httpserver.ServeUpdate("update ticketActivity", echoCtx, &it.UpdateTicketActivityCommand{}, this.Service.UpdateTicketActivity)
}
