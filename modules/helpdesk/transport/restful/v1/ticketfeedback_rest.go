package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/ticketfeedback"
)

type ticketFeedbackRestParams struct {
	dig.In
	Service it.TicketFeedbackService
}

func NewTicketFeedbackRest(params ticketFeedbackRestParams) *TicketFeedbackRest {
	return &TicketFeedbackRest{Service: params.Service}
}

type TicketFeedbackRest struct {
	httpserver.RestBase
	Service it.TicketFeedbackService
}

func (this TicketFeedbackRest) CreateTicketFeedback(echoCtx *echo.Context) (err error) {
	return httpserver.ServeCreate("create ticketFeedback", echoCtx, &it.CreateTicketFeedbackCommand{}, this.Service.CreateTicketFeedback)
}
func (this TicketFeedbackRest) DeleteTicketFeedback(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate("delete ticketFeedback", echoCtx, this.Service.DeleteTicketFeedback)
}
func (this TicketFeedbackRest) GetTicketFeedback(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGetOne("get ticketFeedback", echoCtx, this.Service.GetTicketFeedback)
}
func (this TicketFeedbackRest) TicketFeedbackExists(echoCtx *echo.Context) (err error) {
	return httpserver.ServeExists("ticketFeedback exists", echoCtx, this.Service.TicketFeedbackExists)
}
func (this TicketFeedbackRest) SearchTicketFeedbacks(echoCtx *echo.Context) (err error) {
	return httpserver.ServeSearch("search ticketFeedbacks", echoCtx, this.Service.SearchTicketFeedbacks, true)
}
func (this TicketFeedbackRest) UpdateTicketFeedback(echoCtx *echo.Context) (err error) {
	return httpserver.ServeUpdate("update ticketFeedback", echoCtx, &it.UpdateTicketFeedbackCommand{}, this.Service.UpdateTicketFeedback)
}
