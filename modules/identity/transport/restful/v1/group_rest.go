package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
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

func (this GroupRest) Create(echoCtx echo.Context) (err error) {
	return httpserver.ServeCreate(
		"create group",
		echoCtx,
		&it.CreateGroupCommand{},
		this.GroupSvc.CreateGroup,
	)
}

func (this GroupRest) Delete(echoCtx echo.Context) (err error) {
	return httpserver.ServeGeneralMutate(
		"delete group",
		echoCtx,
		this.GroupSvc.DeleteGroup,
	)
}

func (this GroupRest) GetOne(echoCtx echo.Context) (err error) {
	return httpserver.ServeGetOne(
		"get group by id",
		echoCtx,
		this.GroupSvc.GetGroup,
	)
}

func (this GroupRest) Exists(echoCtx echo.Context) (err error) {
	return httpserver.ServeExists(
		"group exists",
		echoCtx,
		this.GroupSvc.GroupExists,
	)
}

func (this GroupRest) ManageGroupUsers(echoCtx echo.Context) (err error) {
	return httpserver.ServeGeneralMutate(
		"manage group users",
		echoCtx,
		this.GroupSvc.ManageGroupUsers,
	)
}

func (this GroupRest) Search(echoCtx echo.Context) (err error) {
	return httpserver.ServeSearch(
		"search groups",
		echoCtx,
		this.GroupSvc.SearchGroups,
	)
}

func (this GroupRest) Update(echoCtx echo.Context) (err error) {
	return httpserver.ServeUpdate(
		"update group",
		echoCtx,
		&it.UpdateGroupCommand{},
		this.GroupSvc.UpdateGroup,
	)
}
