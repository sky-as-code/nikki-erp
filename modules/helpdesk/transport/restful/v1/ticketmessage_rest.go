package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/ticketmessage"
)

type ticketMessageRestParams struct {
	dig.In
	Service it.TicketMessageService
}

func NewTicketMessageRest(params ticketMessageRestParams) *TicketMessageRest {
	return &TicketMessageRest{Service: params.Service}
}

type TicketMessageRest struct {
	httpserver.RestBase
	Service it.TicketMessageService
}

func (this TicketMessageRest) CreateTicketMessage(echoCtx *echo.Context) (err error) {
	return httpserver.ServeCreate("create ticketMessage", echoCtx, &it.CreateTicketMessageCommand{}, this.Service.CreateTicketMessage)
}
func (this TicketMessageRest) DeleteTicketMessage(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate("delete ticketMessage", echoCtx, this.Service.DeleteTicketMessage)
}
func (this TicketMessageRest) GetTicketMessage(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGetOne("get ticketMessage", echoCtx, this.Service.GetTicketMessage)
}
func (this TicketMessageRest) TicketMessageExists(echoCtx *echo.Context) (err error) {
	return httpserver.ServeExists("ticketMessage exists", echoCtx, this.Service.TicketMessageExists)
}
func (this TicketMessageRest) SearchTicketMessages(echoCtx *echo.Context) (err error) {
	return httpserver.ServeSearch("search ticketMessages", echoCtx, this.Service.SearchTicketMessages)
}
func (this TicketMessageRest) UpdateTicketMessage(echoCtx *echo.Context) (err error) {
	return httpserver.ServeUpdate("update ticketMessage", echoCtx, &it.UpdateTicketMessageCommand{}, this.Service.UpdateTicketMessage)
}
