package v1

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	domain "github.com/sky-as-code/nikki-erp/modules/identity/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/entitlement"
)

type entitlementRestParams struct {
	dig.In

	EntitlementSvc it.EntitlementAppService
}

func NewEntitlementRest(params entitlementRestParams) *EntitlementRest {
	return &EntitlementRest{EntitlementSvc: params.EntitlementSvc}
}

type EntitlementRest struct {
	httpserver.RestBase
	EntitlementSvc it.EntitlementAppService
}

func (this EntitlementRest) CreateEntitlement(echoCtx *echo.Context) (err error) {
	return httpserver.ServeCreate[CreateEntitlementRequest, CreateEntitlementResponse, domain.Entitlement](
		"create entitlement",
		echoCtx,
		&it.CreateEntitlementCommand{},
		this.EntitlementSvc.CreateEntitlement,
	)
}

func (this EntitlementRest) DeleteEntitlement(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate[DeleteEntitlementRequest, DeleteEntitlementResponse](
		"delete entitlement",
		echoCtx,
		this.EntitlementSvc.DeleteEntitlement,
	)
}

func (this EntitlementRest) GetEntitlement(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGetOne2[GetEntitlementRequest, GetEntitlementResponse, domain.Entitlement](
		"get entitlement",
		echoCtx,
		this.EntitlementSvc.GetEntitlement,
	)
}

func (this EntitlementRest) EntitlementExists(echoCtx *echo.Context) (err error) {
	return httpserver.ServeExists[EntitlementExistsRequest, EntitlementExistsResponse](
		"entitlement exists",
		echoCtx,
		this.EntitlementSvc.EntitlementExists,
	)
}

func (this EntitlementRest) ManageEntitlementRoles(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate[ManageEntitlementRolesRequest, ManageEntitlementRolesResponse](
		"manage entitlement roles",
		echoCtx,
		this.EntitlementSvc.ManageEntitlementRoles,
	)
}

func (this EntitlementRest) SearchEntitlements(echoCtx *echo.Context) (err error) {
	return httpserver.ServeSearch[SearchEntitlementsRequest, SearchEntitlementsResponse, domain.Entitlement](
		"search entitlements",
		echoCtx,
		this.EntitlementSvc.SearchEntitlements,
	)
}

func (this EntitlementRest) SetEntitlementIsArchived(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate[SetEntitlementIsArchivedRequest, SetEntitlementIsArchivedResponse](
		"set entitlement is_archived",
		echoCtx,
		this.EntitlementSvc.SetEntitlementIsArchived,
	)
}

func (this EntitlementRest) UpdateEntitlement(echoCtx *echo.Context) (err error) {
	return httpserver.ServeUpdate[UpdateEntitlementRequest, UpdateEntitlementResponse](
		"update entitlement",
		echoCtx,
		&it.UpdateEntitlementCommand{},
		this.EntitlementSvc.UpdateEntitlement,
	)
}

/*
 * Non-CRUD APIs
 */

func (this EntitlementRest) GetModelSchema(echoCtx *echo.Context) (err error) {
	schema := dmodel.MustGetSchema(domain.EntitlementSchemaName)
	echoCtx.JSON(http.StatusOK, schema.ToSimplized())
	return nil
}
