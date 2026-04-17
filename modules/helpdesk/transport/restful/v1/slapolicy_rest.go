package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/slapolicy"
)

type slaPolicyRestParams struct {
	dig.In
	Service it.SlaPolicyService
}

func NewSlaPolicyRest(params slaPolicyRestParams) *SlaPolicyRest {
	return &SlaPolicyRest{Service: params.Service}
}

type SlaPolicyRest struct {
	httpserver.RestBase
	Service it.SlaPolicyService
}

func (this SlaPolicyRest) CreateSlaPolicy(echoCtx *echo.Context) (err error) {
	return httpserver.ServeCreate("create slaPolicy", echoCtx, &it.CreateSlaPolicyCommand{}, this.Service.CreateSlaPolicy)
}
func (this SlaPolicyRest) DeleteSlaPolicy(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate("delete slaPolicy", echoCtx, this.Service.DeleteSlaPolicy)
}
func (this SlaPolicyRest) GetSlaPolicy(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGetOne("get slaPolicy", echoCtx, this.Service.GetSlaPolicy)
}
func (this SlaPolicyRest) SlaPolicyExists(echoCtx *echo.Context) (err error) {
	return httpserver.ServeExists("slaPolicy exists", echoCtx, this.Service.SlaPolicyExists)
}
func (this SlaPolicyRest) SearchSlaPolicies(echoCtx *echo.Context) (err error) {
	return httpserver.ServeSearch("search slaPolicys", echoCtx, this.Service.SearchSlaPolicies)
}
func (this SlaPolicyRest) UpdateSlaPolicy(echoCtx *echo.Context) (err error) {
	return httpserver.ServeUpdate("update slaPolicy", echoCtx, &it.UpdateSlaPolicyCommand{}, this.Service.UpdateSlaPolicy)
}
func (this SlaPolicyRest) SetSlaPolicyIsArchived(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate("set slaPolicy is_archived", echoCtx, this.Service.SetSlaPolicyIsArchived)
}
