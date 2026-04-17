package v1

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/group"
)

type groupRestParams struct {
	dig.In

	GroupSvc it.GroupService
}

func NewGroupRest(params groupRestParams) *GroupRest {
	return &GroupRest{
		GroupSvc: params.GroupSvc,
	}
}

type GroupRest struct {
	httpserver.RestBase
	GroupSvc it.GroupService
}

func (this GroupRest) CreateGroup(echoCtx *echo.Context) (err error) {
	return httpserver.ServeCreate(
		"create group",
		echoCtx,
		&it.CreateGroupCommand{},
		this.GroupSvc.CreateGroup,
	)
}

func (this GroupRest) DeleteGroup(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate(
		"delete group",
		echoCtx,
		this.GroupSvc.DeleteGroup,
	)
}

func (this GroupRest) GetGroup(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGetOne(
		"get group by id",
		echoCtx,
		this.GroupSvc.GetGroup,
	)
}

func (this GroupRest) GroupExists(echoCtx *echo.Context) (err error) {
	return httpserver.ServeExists(
		"group exists",
		echoCtx,
		this.GroupSvc.GroupExists,
	)
}

func (this GroupRest) ManageGroupUsers(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate(
		"manage group users",
		echoCtx,
		this.GroupSvc.ManageGroupUsers,
	)
}

func (this GroupRest) SearchGroups(echoCtx *echo.Context) (err error) {
	return httpserver.ServeSearch(
		"search groups",
		echoCtx,
		this.GroupSvc.SearchGroups,
	)
}

func (this GroupRest) UpdateGroup(echoCtx *echo.Context) (err error) {
	return httpserver.ServeUpdate(
		"update group",
		echoCtx,
		&it.UpdateGroupCommand{},
		this.GroupSvc.UpdateGroup,
	)
}

/*
 * Non-CRUD APIs
 */

func (this GroupRest) GetModelSchema(echoCtx *echo.Context) (err error) {
	schema := dmodel.MustGetSchema(domain.GroupSchemaName)
	echoCtx.JSON(http.StatusOK, schema.ToSimplized())
	return nil
}
