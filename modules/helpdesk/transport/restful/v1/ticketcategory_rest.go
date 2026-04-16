package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/ticketcategory"
)

type ticketCategoryRestParams struct {
	dig.In
	Service it.TicketCategoryService
}

func NewTicketCategoryRest(params ticketCategoryRestParams) *TicketCategoryRest {
	return &TicketCategoryRest{Service: params.Service}
}

type TicketCategoryRest struct {
	httpserver.RestBase
	Service it.TicketCategoryService
}

func (this TicketCategoryRest) CreateTicketCategory(echoCtx *echo.Context) (err error) {
	return httpserver.ServeCreate("create ticketCategory", echoCtx, &it.CreateTicketCategoryCommand{}, this.Service.CreateTicketCategory)
}
func (this TicketCategoryRest) DeleteTicketCategory(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate("delete ticketCategory", echoCtx, this.Service.DeleteTicketCategory)
}
func (this TicketCategoryRest) GetTicketCategory(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGetOne("get ticketCategory", echoCtx, this.Service.GetTicketCategory)
}
func (this TicketCategoryRest) TicketCategoryExists(echoCtx *echo.Context) (err error) {
	return httpserver.ServeExists("ticketCategory exists", echoCtx, this.Service.TicketCategoryExists)
}
func (this TicketCategoryRest) SearchTicketCategories(echoCtx *echo.Context) (err error) {
	return httpserver.ServeSearch("search ticketCategorys", echoCtx, this.Service.SearchTicketCategories, true)
}
func (this TicketCategoryRest) UpdateTicketCategory(echoCtx *echo.Context) (err error) {
	return httpserver.ServeUpdate("update ticketCategory", echoCtx, &it.UpdateTicketCategoryCommand{}, this.Service.UpdateTicketCategory)
}
func (this TicketCategoryRest) SetTicketCategoryIsArchived(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate("set ticketCategory is_archived", echoCtx, this.Service.SetTicketCategoryIsArchived)
}
