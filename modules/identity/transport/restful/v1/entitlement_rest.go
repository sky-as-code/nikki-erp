package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/entitlement"
)

type entitlementRestParams struct {
	dig.In

	EntitlementSvc it.EntitlementService
}

func NewEntitlementRest(params entitlementRestParams) *EntitlementRest {
	return &EntitlementRest{EntitlementSvc: params.EntitlementSvc}
}

type EntitlementRest struct {
	httpserver.RestBase
	EntitlementSvc it.EntitlementService
}

func (this EntitlementRest) CreateEntitlement(echoCtx *echo.Context) (err error) {
	return httpserver.ServeCreate(
		"create entitlement",
		echoCtx,
		&it.CreateEntitlementCommand{},
		this.EntitlementSvc.CreateEntitlement,
	)
}

func (this EntitlementRest) DeleteEntitlement(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate(
		"delete entitlement",
		echoCtx,
		this.EntitlementSvc.DeleteEntitlement,
	)
}

func (this EntitlementRest) GetEntitlement(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGetOne(
		"get entitlement",
		echoCtx,
		this.EntitlementSvc.GetEntitlement,
	)
}

func (this EntitlementRest) EntitlementExists(echoCtx *echo.Context) (err error) {
	return httpserver.ServeExists(
		"entitlement exists",
		echoCtx,
		this.EntitlementSvc.EntitlementExists,
	)
}

func (this EntitlementRest) ManageEntitlementRoles(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate(
		"manage entitlement roles",
		echoCtx,
		this.EntitlementSvc.ManageEntitlementRoles,
	)
}

func (this EntitlementRest) SearchEntitlements(echoCtx *echo.Context) (err error) {
	return httpserver.ServeSearch(
		"search entitlements",
		echoCtx,
		this.EntitlementSvc.SearchEntitlements,
		true,
	)
}

func (this EntitlementRest) SetEntitlementIsArchived(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate(
		"set entitlement is_archived",
		echoCtx,
		this.EntitlementSvc.SetEntitlementIsArchived,
	)
}

func (this EntitlementRest) UpdateEntitlement(echoCtx *echo.Context) (err error) {
	return httpserver.ServeUpdate(
		"update entitlement",
		echoCtx,
		&it.UpdateEntitlementCommand{},
		this.EntitlementSvc.UpdateEntitlement,
	)
}
