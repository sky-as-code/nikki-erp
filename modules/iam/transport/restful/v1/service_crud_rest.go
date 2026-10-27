package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itRole "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/role"
	itService "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/service"
)

type serviceCrudRestParams struct {
	dig.In

	Engine         composable.DynamicResourceEngineOnion `name:"dynengine_iam_service"`
	PrincipalRoles itRole.PrincipalRolesApplicationService
}

func NewServiceCrudRest(params serviceCrudRestParams) *ServiceCrudRest {
	svc := params.Engine.ApplicationService().(itService.ServiceCrudApplicationService)
	rest := &ServiceCrudRest{serviceSvc: svc, rolesSvc: params.PrincipalRoles}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

// ServiceCrudRest serves the built-in CRUD of the iam_service resource and its role assignment.
type ServiceCrudRest struct {
	composable.CrudRestBase
	serviceSvc itService.ServiceCrudApplicationService
	rolesSvc   itRole.PrincipalRolesApplicationService
}

func (this *ServiceCrudRest) ManageRoles(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "manage service roles", payload, this.serviceSvc.ManageRoles,
		composable.MutateResponse)
}

func (this *ServiceCrudRest) Roles(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "search service roles", payload, this.rolesSvc.ServiceRoles,
		composable.SearchResponse)
}
