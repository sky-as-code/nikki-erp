package v1

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	domain "github.com/sky-as-code/nikki-erp/modules/identity/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/group"
)

type groupRestParams struct {
	dig.In

	GroupSvc it.GroupAppService
}

func NewGroupRest(params groupRestParams) *GroupRest {
	return &GroupRest{
		GroupSvc: params.GroupSvc,
	}
}

type GroupRest struct {
	httpserver.RestBase
	GroupSvc it.GroupAppService
}

func (this GroupRest) CreateGroup(echoCtx *echo.Context) (err error) {
	return httpserver.ServeCreate[CreateGroupRequest, CreateGroupResponse, domain.Group](
		"create group",
		echoCtx,
		&it.CreateGroupCommand{},
		this.GroupSvc.CreateGroup,
	)
}

func (this GroupRest) DeleteGroup(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate[DeleteGroupRequest, DeleteGroupResponse](
		"delete group",
		echoCtx,
		this.GroupSvc.DeleteGroup,
	)
}

func (this GroupRest) GetGroup(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGetOne2[GetGroupRequest, GetGroupResponse, domain.Group](
		"get group by id",
		echoCtx,
		this.GroupSvc.GetGroup,
	)
}

func (this GroupRest) GroupExists(echoCtx *echo.Context) (err error) {
	return httpserver.ServeExists[GroupExistsRequest, GroupExistsResponse](
		"group exists",
		echoCtx,
		this.GroupSvc.GroupExists,
	)
}

func (this GroupRest) ManageGroupUsers(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate[ManageGroupUsersRequest, ManageGroupUsersResponse](
		"manage group users",
		echoCtx,
		this.GroupSvc.ManageGroupUsers,
	)
}

func (this GroupRest) SearchGroups(echoCtx *echo.Context) (err error) {
	return httpserver.ServeSearch[SearchGroupsRequest, SearchGroupsResponse, domain.Group](
		"search groups",
		echoCtx,
		this.GroupSvc.SearchGroups,
	)
}

func (this GroupRest) UpdateGroup(echoCtx *echo.Context) (err error) {
	return httpserver.ServeUpdate[UpdateGroupRequest, UpdateGroupResponse](
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
