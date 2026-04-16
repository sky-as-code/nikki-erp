package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/slabreach"
)

type slaBreachRestParams struct {
	dig.In
	Service it.SlaBreachService
}

func NewSlaBreachRest(params slaBreachRestParams) *SlaBreachRest {
	return &SlaBreachRest{Service: params.Service}
}

type SlaBreachRest struct {
	httpserver.RestBase
	Service it.SlaBreachService
}

func (this SlaBreachRest) CreateSlaBreach(echoCtx *echo.Context) (err error) {
	return httpserver.ServeCreate("create slaBreach", echoCtx, &it.CreateSlaBreachCommand{}, this.Service.CreateSlaBreach)
}
func (this SlaBreachRest) DeleteSlaBreach(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate("delete slaBreach", echoCtx, this.Service.DeleteSlaBreach)
}
func (this SlaBreachRest) GetSlaBreach(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGetOne("get slaBreach", echoCtx, this.Service.GetSlaBreach)
}
func (this SlaBreachRest) SlaBreachExists(echoCtx *echo.Context) (err error) {
	return httpserver.ServeExists("slaBreach exists", echoCtx, this.Service.SlaBreachExists)
}
func (this SlaBreachRest) SearchSlaBreaches(echoCtx *echo.Context) (err error) {
	return httpserver.ServeSearch("search slaBreachs", echoCtx, this.Service.SearchSlaBreaches, true)
}
func (this SlaBreachRest) UpdateSlaBreach(echoCtx *echo.Context) (err error) {
	return httpserver.ServeUpdate("update slaBreach", echoCtx, &it.UpdateSlaBreachCommand{}, this.Service.UpdateSlaBreach)
}
