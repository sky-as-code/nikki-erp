package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	c "github.com/sky-as-code/nikki-erp/modules/auditlog/constants"
	itAction "github.com/sky-as-code/nikki-erp/modules/auditlog/interfaces/action"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

type actionRestParams struct {
	dig.In

	Engine    composable.DynamicResourceEngineOnion `name:"dynengine_auditlog_action"`
	ActionSvc itAction.ActionApplicationService
}

func NewActionRest(params actionRestParams) *ActionRest {
	rest := &ActionRest{actionSvc: params.ActionSvc}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

// ActionRest serves the read-only built-ins and the per-record history.
type ActionRest struct {
	composable.CrudRestBase

	actionSvc itAction.ActionApplicationService
}

var historyQueryParams = []string{
	basemodel.FieldOrgId, "resource_code", "resource_id", basemodel.FieldPage, basemodel.FieldSize,
}

// GetHistory is mounted by hand at c.RouteActionHistory rather than through the engine: the
// engine would place it under /auditlog_action, and the frontend codes against /actions/history.
func (this *ActionRest) GetHistory(echoCtx *echo.Context) error {
	payload := map[string]any{}
	for _, name := range historyQueryParams {
		if value := echoCtx.QueryParam(name); value != "" {
			payload[name] = value
		}
	}
	return composable.ServeAction(echoCtx, "get "+c.ResourceAuditLogAction+" history", payload,
		this.actionSvc.GetHistory, composable.Identity[itAction.HistoryResultData])
}
