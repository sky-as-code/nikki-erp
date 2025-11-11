package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/inventory/attributegroup/interfaces"
)

type attributeGroupRestParams struct {
	dig.In

	AttributeGroupSvc it.AttributeGroupService
}

func NewAttributeGroupRest(params attributeGroupRestParams) *AttributeGroupRest {
	return &AttributeGroupRest{
		AttributeGroupSvc: params.AttributeGroupSvc,
	}
}

type AttributeGroupRest struct {
	httpserver.RestBase
	AttributeGroupSvc it.AttributeGroupService
}

func (this AttributeGroupRest) CreateAttributeGroup(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create attribute group"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.AttributeGroupSvc.CreateAttributeGroup,
		func(request CreateAttributeGroupRequest) it.CreateAttributeGroupCommand {
			return it.CreateAttributeGroupCommand(request)
		},
		func(result it.CreateAttributeGroupResult) CreateAttributeGroupResponse {
			response := CreateAttributeGroupResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonCreated,
	)
	return err
}

func (this AttributeGroupRest) UpdateAttributeGroup(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST update attribute group"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.AttributeGroupSvc.UpdateAttributeGroup,
		func(request UpdateAttributeGroupRequest) it.UpdateAttributeGroupCommand {
			return it.UpdateAttributeGroupCommand(request)
		},
		func(result it.UpdateAttributeGroupResult) UpdateAttributeGroupResponse {
			response := UpdateAttributeGroupResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this AttributeGroupRest) DeleteAttributeGroup(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST delete attribute group"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.AttributeGroupSvc.DeleteAttributeGroup,
		func(request DeleteAttributeGroupRequest) it.DeleteAttributeGroupCommand {
			return it.DeleteAttributeGroupCommand(request)
		},
		func(result it.DeleteAttributeGroupResult) DeleteAttributeGroupResponse {
			response := DeleteAttributeGroupResponse{}
			response.FromNonEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this AttributeGroupRest) GetAttributeGroupById(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get attribute group by id"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.AttributeGroupSvc.GetAttributeGroupById,
		func(request GetAttributeGroupByIdRequest) it.GetAttributeGroupByIdQuery {
			return it.GetAttributeGroupByIdQuery(request)
		},
		func(result it.GetAttributeGroupByIdResult) GetAttributeGroupByIdResponse {
			response := GetAttributeGroupByIdResponse{}
			response.FromAttributeGroup(*result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this AttributeGroupRest) SearchAttributeGroups(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST search attribute groups"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.AttributeGroupSvc.SearchAttributeGroups,
		func(request SearchAttributeGroupsRequest) it.SearchAttributeGroupsQuery {
			return it.SearchAttributeGroupsQuery(request)
		},
		func(result it.SearchAttributeGroupsResult) SearchAttributeGroupsResponse {
			response := SearchAttributeGroupsResponse{}
			response.FromResult(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}
