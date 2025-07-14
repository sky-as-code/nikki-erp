package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
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

func (this GroupRest) CreateGroup(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create group"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.GroupSvc.CreateGroup,
		func(request CreateGroupRequest) it.CreateGroupCommand {
			return it.CreateGroupCommand(request)
		},
		func(result it.CreateGroupResult) CreateGroupResponse {
			response := CreateGroupResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonCreated,
	)
	return err
}

func (this GroupRest) UpdateGroup(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST update group"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.GroupSvc.UpdateGroup,
		func(request UpdateGroupRequest) it.UpdateGroupCommand {
			return it.UpdateGroupCommand(request)
		},
		func(result it.UpdateGroupResult) UpdateGroupResponse {
			response := UpdateGroupResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this GroupRest) GetGroupById(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get group by id"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.GroupSvc.GetGroupById,
		func(request GetGroupByIdRequest) it.GetGroupByIdQuery {
			return it.GetGroupByIdQuery(request)
		},
		func(result it.GetGroupByIdResult) GetGroupByIdResponse {
			response := GetGroupByIdResponse{}
			response.FromGroup(*result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this GroupRest) DeleteGroup(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST delete group"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.GroupSvc.DeleteGroup,
		func(request DeleteGroupRequest) it.DeleteGroupCommand {
			return it.DeleteGroupCommand(request)
		},
		func(result it.DeleteGroupResult) DeleteGroupResponse {
			response := DeleteGroupResponse{}
			response.FromNonEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this GroupRest) SearchGroups(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST delete group"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.GroupSvc.SearchGroups,
		func(request SearchGroupsRequest) it.SearchGroupsQuery {
			return it.SearchGroupsQuery(request)
		},
		func(result it.SearchGroupsResult) SearchGroupsResponse {
			response := SearchGroupsResponse{}
			response.FromResult(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this GroupRest) ManageUsers(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST manage users"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.GroupSvc.AddRemoveUsers,
		func(request ManageUsersRequest) it.AddRemoveUsersCommand {
			return it.AddRemoveUsersCommand(request)
		},
		func(result it.AddRemoveUsersResult) ManageUsersResponse {
			response := ManageUsersResponse{}
			response.FromNonEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}
