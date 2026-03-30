package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
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

func (this GroupRest) CreateGroup(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create group"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequestDynamic(
		echoCtx,
		this.GroupSvc.CreateGroup,
		func(requestFields dmodel.DynamicFields) it.CreateGroupCommand {
			cmd := it.CreateGroupCommand{}
			cmd.SetFieldData(requestFields)
			return cmd
		},
		func(data domain.Group) CreateGroupResponse {
			response := httpserver.NewRestCreateResponseDyn(data.GetFieldData())
			return *response
		},
		httpserver.JsonCreated,
	)
}

func (this GroupRest) DeleteGroup(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST delete group"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.GroupSvc.DeleteGroup,
		func(request DeleteGroupRequest) it.DeleteGroupCommand {
			return it.DeleteGroupCommand(request)
		},
		func(data dyn.MutateResultData) DeleteGroupResponse {
			return httpserver.NewRestDeleteResponse2(data)
		},
		httpserver.JsonOk,
	)
}

func (this GroupRest) GetGroup(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get group by id"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.GroupSvc.GetGroup,
		func(request GetGroupRequest) it.GetGroupQuery {
			return request
		},
		func(data domain.Group) GetGroupResponse {
			return data.GetFieldData()
		},
		httpserver.JsonOk,
	)
}

func (this GroupRest) GroupExists(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST group exists"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.GroupSvc.GroupExists,
		func(request GroupExistsRequest) it.GroupExistsQuery {
			return it.GroupExistsQuery(request)
		},
		func(data dyn.ExistsResultData) GroupExistsResponse {
			return UserExistsResponse(data)
		},
		httpserver.JsonOk,
	)
}
func (this GroupRest) ManageGroupUsers(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST manage group users"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.GroupSvc.ManageGroupUsers,
		func(request ManageGroupUsersRequest) it.ManageGroupUsersCommand {
			return it.ManageGroupUsersCommand(request)
		},
		func(data dyn.MutateResultData) ManageGroupUsersResponse {
			return httpserver.NewRestMutateResponse(data)
		},
		httpserver.JsonOk,
	)
}

func (this GroupRest) SearchGroups(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST search groups"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.GroupSvc.SearchGroups,
		func(request SearchGroupsRequest) it.SearchGroupsQuery {
			return it.SearchGroupsQuery(request)
		},
		func(data it.SearchGroupsResultData) SearchGroupsResponse {
			return httpserver.NewSearchUsersResponseDyn(data)
		},
		httpserver.JsonOk,
		true,
	)
}

func (this GroupRest) UpdateGroup(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST update group"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequestDynamic(
		echoCtx,
		this.GroupSvc.UpdateGroup,
		func(requestFields dmodel.DynamicFields) it.UpdateGroupCommand {
			cmd := it.UpdateGroupCommand{}
			cmd.SetFieldData(requestFields)
			return cmd
		},
		func(data dyn.MutateResultData) UpdateGroupResponse {
			return httpserver.NewRestUpdateResponse2(data)
		},
		httpserver.JsonOk,
	)
}
