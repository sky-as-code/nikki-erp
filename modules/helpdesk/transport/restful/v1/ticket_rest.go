package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/ticket"
)

type ticketRestParams struct {
	dig.In
	Service it.TicketService
}

func NewTicketRest(params ticketRestParams) *TicketRest { return &TicketRest{Service: params.Service} }

type TicketRest struct {
	httpserver.RestBase
	Service it.TicketService
}

func (this TicketRest) CreateTicket(echoCtx *echo.Context) (err error) {
	return httpserver.ServeCreate("create ticket", echoCtx, &it.CreateTicketCommand{}, this.Service.CreateTicket)
}
func (this TicketRest) DeleteTicket(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate("delete ticket", echoCtx, this.Service.DeleteTicket)
}
func (this TicketRest) GetTicket(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGetOne("get ticket", echoCtx, this.Service.GetTicket)
}
func (this TicketRest) TicketExists(echoCtx *echo.Context) (err error) {
	return httpserver.ServeExists("ticket exists", echoCtx, this.Service.TicketExists)
}
func (this TicketRest) SearchTickets(echoCtx *echo.Context) (err error) {
	return httpserver.ServeSearch("search tickets", echoCtx, this.Service.SearchTickets)
}
func (this TicketRest) UpdateTicket(echoCtx *echo.Context) (err error) {
	return httpserver.ServeUpdate("update ticket", echoCtx, &it.UpdateTicketCommand{}, this.Service.UpdateTicket)
}
func (this TicketRest) SetTicketIsArchived(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate("set ticket is_archived", echoCtx, this.Service.SetTicketIsArchived)
}
func (this TicketRest) ManageTicketCategories(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate("manage ticket categories", echoCtx, this.Service.ManageTicketCategories)
}
