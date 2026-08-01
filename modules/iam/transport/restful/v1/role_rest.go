package v1

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/role"
)

type roleRestParams struct {
	dig.In

	RoleSvc it.RoleAppService
}

func NewRoleRest(params roleRestParams) *RoleRest {
	return &RoleRest{RoleSvc: params.RoleSvc}
}

type RoleRest struct {
	httpserver.RestBase
	RoleSvc it.RoleAppService
}

func (this RoleRest) CreateRole(echoCtx *echo.Context) (err error) {
	return httpserver.ServeCreate[CreateRoleRequest, CreateRoleResponse, domain.Role](
		"create role",
		echoCtx,
		&it.CreateRoleCommand{},
		this.RoleSvc.CreateRole,
	)
}

func (this RoleRest) DeleteRole(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate[DeleteRoleRequest, DeleteRoleResponse](
		"delete role",
		echoCtx,
		this.RoleSvc.DeleteRole,
	)
}

func (this RoleRest) GetRole(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGetOne2[GetRoleRequest, GetRoleResponse, domain.Role](
		"get role",
		echoCtx,
		this.RoleSvc.GetRole,
	)
}

func (this RoleRest) ManageRoleEntitlements(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate[ManageRoleEntitlementsRequest, ManageRoleEntitlementsResponse](
		"manage role entitlements",
		echoCtx,
		this.RoleSvc.ManageRoleEntitlements,
	)
}

func (this RoleRest) RoleExists(echoCtx *echo.Context) (err error) {
	return httpserver.ServeExists[RoleExistsRequest, RoleExistsResponse](
		"role exists",
		echoCtx,
		this.RoleSvc.RoleExists,
	)
}

func (this RoleRest) SearchRoles(echoCtx *echo.Context) (err error) {
	return httpserver.ServeSearch[SearchRolesRequest, SearchRolesResponse, domain.Role](
		"search roles",
		echoCtx,
		this.RoleSvc.SearchRoles,
	)
}

// SearchUserRoles serves GET /users/:user_id/roles. It lives on RoleRest rather than UserRest
// because it searches the role schema, mirroring how GET /resources/:resource_id/actions is
// served by ActionRest.
func (this RoleRest) SearchUserRoles(echoCtx *echo.Context) (err error) {
	return httpserver.ServeSearch[SearchUserRolesRequest, SearchUserRolesResponse, domain.Role](
		"search user roles",
		echoCtx,
		this.RoleSvc.SearchUserRoles,
	)
}

func (this RoleRest) SearchGroupRoles(echoCtx *echo.Context) (err error) {
	return httpserver.ServeSearch[SearchGroupRolesRequest, SearchGroupRolesResponse, domain.Role](
		"search group roles",
		echoCtx,
		this.RoleSvc.SearchGroupRoles,
	)
}

func (this RoleRest) DescribeRoles(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST describe roles"); e != nil {
			err = e
		}
	}()
	// skipNotFoundError: an empty role list is a valid answer, not a 404.
	return httpserver.ServeRequest2[DescribeRolesRequest, DescribeRolesResponse](
		echoCtx,
		this.RoleSvc.DescribeRoles,
		httpserver.ItsMeMario,
		httpserver.ItsMeMario,
		httpserver.JsonOk,
		true,
	)
}

func (this RoleRest) SetRoleIsArchived(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate[SetRoleIsArchivedRequest, SetRoleIsArchivedResponse](
		"set role is_archived",
		echoCtx,
		this.RoleSvc.SetRoleIsArchived,
	)
}

func (this RoleRest) UpdateRole(echoCtx *echo.Context) (err error) {
	return httpserver.ServeUpdate[UpdateRoleRequest, UpdateRoleResponse](
		"update role",
		echoCtx,
		&it.UpdateRoleCommand{},
		this.RoleSvc.UpdateRole,
	)
}

/*
 * Non-CRUD APIs
 */

func (this RoleRest) GetModelSchema(echoCtx *echo.Context) (err error) {
	schema := dmodel.MustGetSchema(domain.RoleSchemaName)
	echoCtx.JSON(http.StatusOK, schema.ToSimplized())
	return nil
}
