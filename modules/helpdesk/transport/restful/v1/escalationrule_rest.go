package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/escalationrule"
)

type escalationRuleRestParams struct {
	dig.In
	Service it.EscalationRuleService
}

func NewEscalationRuleRest(params escalationRuleRestParams) *EscalationRuleRest {
	return &EscalationRuleRest{Service: params.Service}
}

type EscalationRuleRest struct {
	httpserver.RestBase
	Service it.EscalationRuleService
}

func (this EscalationRuleRest) CreateEscalationRule(echoCtx *echo.Context) (err error) {
	return httpserver.ServeCreate("create escalationRule", echoCtx, &it.CreateEscalationRuleCommand{}, this.Service.CreateEscalationRule)
}
func (this EscalationRuleRest) DeleteEscalationRule(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate("delete escalationRule", echoCtx, this.Service.DeleteEscalationRule)
}
func (this EscalationRuleRest) GetEscalationRule(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGetOne("get escalationRule", echoCtx, this.Service.GetEscalationRule)
}
func (this EscalationRuleRest) EscalationRuleExists(echoCtx *echo.Context) (err error) {
	return httpserver.ServeExists("escalationRule exists", echoCtx, this.Service.EscalationRuleExists)
}
func (this EscalationRuleRest) SearchEscalationRules(echoCtx *echo.Context) (err error) {
	return httpserver.ServeSearch("search escalationRules", echoCtx, this.Service.SearchEscalationRules)
}
func (this EscalationRuleRest) UpdateEscalationRule(echoCtx *echo.Context) (err error) {
	return httpserver.ServeUpdate("update escalationRule", echoCtx, &it.UpdateEscalationRuleCommand{}, this.Service.UpdateEscalationRule)
}
