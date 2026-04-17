package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/ticketassignment"
)

type ticketAssignmentRestParams struct {
	dig.In
	Service it.TicketAssignmentService
}

func NewTicketAssignmentRest(params ticketAssignmentRestParams) *TicketAssignmentRest {
	return &TicketAssignmentRest{Service: params.Service}
}

type TicketAssignmentRest struct {
	httpserver.RestBase
	Service it.TicketAssignmentService
}

func (this TicketAssignmentRest) CreateTicketAssignment(echoCtx *echo.Context) (err error) {
	return httpserver.ServeCreate("create ticketAssignment", echoCtx, &it.CreateTicketAssignmentCommand{}, this.Service.CreateTicketAssignment)
}
func (this TicketAssignmentRest) DeleteTicketAssignment(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate("delete ticketAssignment", echoCtx, this.Service.DeleteTicketAssignment)
}
func (this TicketAssignmentRest) GetTicketAssignment(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGetOne("get ticketAssignment", echoCtx, this.Service.GetTicketAssignment)
}
func (this TicketAssignmentRest) TicketAssignmentExists(echoCtx *echo.Context) (err error) {
	return httpserver.ServeExists("ticketAssignment exists", echoCtx, this.Service.TicketAssignmentExists)
}
func (this TicketAssignmentRest) SearchTicketAssignments(echoCtx *echo.Context) (err error) {
	return httpserver.ServeSearch("search ticketAssignments", echoCtx, this.Service.SearchTicketAssignments)
}
func (this TicketAssignmentRest) UpdateTicketAssignment(echoCtx *echo.Context) (err error) {
	return httpserver.ServeUpdate("update ticketAssignment", echoCtx, &it.UpdateTicketAssignmentCommand{}, this.Service.UpdateTicketAssignment)
}
