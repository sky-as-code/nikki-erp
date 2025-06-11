package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/group"
)

type groupRestParams struct {
	dig.In

	Config  config.ConfigService
	Logger  logging.LoggerService
	CqrsBus cqrs.CqrsBus
}

func NewGroupRest(params groupRestParams) *GroupRest {
	return &GroupRest{
		RestBase: httpserver.RestBase{
			ConfigSvc: params.Config,
			Logger:    params.Logger,
			CqrsBus:   params.CqrsBus,
		},
	}
}

type GroupRest struct {
	httpserver.RestBase
}

func (this GroupRest) CreateGroup(echoCtx echo.Context) (err error) {
	request := &CreateGroupRequest{}
	if err = echoCtx.Bind(request); err != nil {
		return err
	}

	result := it.CreateGroupResult{}
	err = this.CqrsBus.Request(echoCtx.Request().Context(), *request, &result)

	if err != nil {
		return err
	}

	if result.ClientError != nil {
		return httpserver.JsonBadRequest(echoCtx, result.ClientError)
	}

	response := CreateGroupResponse{}
	response.FromGroup(*result.Data)

	return httpserver.JsonCreated(echoCtx, response)
}

func (this GroupRest) UpdateGroup(echoCtx echo.Context) (err error) {
	request := &UpdateGroupRequest{}
	if err = echoCtx.Bind(request); err != nil {
		return err
	}

	result := it.UpdateGroupResult{}
	err = this.CqrsBus.Request(echoCtx.Request().Context(), *request, &result)

	if err != nil {
		return err
	}

	if result.ClientError != nil {
		return httpserver.JsonBadRequest(echoCtx, result.ClientError)
	}

	response := UpdateGroupResponse{}
	response.FromGroup(*result.Data)

	return httpserver.JsonOk(echoCtx, response)
}

func (this GroupRest) GetGroupById(echoCtx echo.Context) (err error) {
	request := &GetGroupByIdRequest{}
	if err = echoCtx.Bind(request); err != nil {
		return err
	}

	result := it.GetGroupByIdResult{}
	err = this.CqrsBus.Request(echoCtx.Request().Context(), *request, &result)

	if err != nil {
		return err
	}

	if result.ClientError != nil {
		return httpserver.JsonBadRequest(echoCtx, result.ClientError)
	}

	response := GetGroupByIdResponse{}
	response.FromGroup(*result.Data)

	return httpserver.JsonOk(echoCtx, response)
}

func (this GroupRest) DeleteGroup(echoCtx echo.Context) (err error) {
	request := &DeleteGroupRequest{}
	if err = echoCtx.Bind(request); err != nil {
		return err
	}

	result := it.DeleteGroupResult{}
	err = this.CqrsBus.Request(echoCtx.Request().Context(), *request, &result)

	if err != nil {
		return err
	}

	if result.ClientError != nil {
		return httpserver.JsonBadRequest(echoCtx, result.ClientError)
	}

	response := DeleteGroupResponse{
		DeletedAt: result.Data.DeletedAt.UnixMilli(),
	}

	return httpserver.JsonOk(echoCtx, response)
}

func (this GroupRest) SearchGroups(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to search users"); e != nil {
			err = e
		}
	}()

	request := &SearchGroupsRequest{}
	if err = echoCtx.Bind(request); err != nil {
		return err
	}

	result := it.SearchGroupsResult{}
	err = this.CqrsBus.Request(echoCtx.Request().Context(), *request, &result)

	if err != nil {
		return err
	}

	if result.ClientError != nil {
		return httpserver.JsonBadRequest(echoCtx, result.ClientError)
	}

	response := SearchGroupsResponse{}
	response.FromResult(result.Data)

	return httpserver.JsonOk(echoCtx, response)
}

func (this GroupRest) ManageUsers(echoCtx echo.Context) (err error) {
	request := &ManageUsersRequest{}
	if err = echoCtx.Bind(request); err != nil {
		return err
	}

	result := it.AddRemoveUsersResult{}
	err = this.CqrsBus.Request(echoCtx.Request().Context(), *request, &result)

	if err != nil {
		return err
	}

	if result.ClientError != nil {
		return httpserver.JsonBadRequest(echoCtx, result.ClientError)
	}

	response := ManageUsersResponse{}
	response.FromResult(result.Data)

	return httpserver.JsonOk(echoCtx, response)
}
