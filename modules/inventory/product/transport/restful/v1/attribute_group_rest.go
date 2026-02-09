package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	itAttributeGroup "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attributegroup"
)

type attributeGroupRestParams struct {
	dig.In

	AttributeGroupSvc itAttributeGroup.AttributeGroupService
}

func NewAttributeGroupRest(params attributeGroupRestParams) *AttributeGroupRest {
	return &AttributeGroupRest{
		AttributeGroupSvc: params.AttributeGroupSvc,
	}
}

type AttributeGroupRest struct {
	httpserver.RestBase
	AttributeGroupSvc itAttributeGroup.AttributeGroupService
}

func (this AttributeGroupRest) CreateAttributeGroup(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create attribute group"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.AttributeGroupSvc.CreateAttributeGroup,
		func(request CreateAttributeGroupRequest) itAttributeGroup.CreateAttributeGroupCommand {
			return itAttributeGroup.CreateAttributeGroupCommand(request)
		},
		func(result itAttributeGroup.CreateAttributeGroupResult) CreateAttributeGroupResponse {
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
		func(request UpdateAttributeGroupRequest) itAttributeGroup.UpdateAttributeGroupCommand {
			return itAttributeGroup.UpdateAttributeGroupCommand(request)
		},
		func(result itAttributeGroup.UpdateAttributeGroupResult) UpdateAttributeGroupResponse {
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
		func(request DeleteAttributeGroupRequest) itAttributeGroup.DeleteAttributeGroupCommand {
			return itAttributeGroup.DeleteAttributeGroupCommand(request)
		},
		func(result itAttributeGroup.DeleteAttributeGroupResult) DeleteAttributeGroupResponse {
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
		func(request GetAttributeGroupByIdRequest) itAttributeGroup.GetAttributeGroupByIdQuery {
			return itAttributeGroup.GetAttributeGroupByIdQuery(request)
		},
		func(result itAttributeGroup.GetAttributeGroupByIdResult) GetAttributeGroupByIdResponse {
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
		func(request SearchAttributeGroupsRequest) itAttributeGroup.SearchAttributeGroupsQuery {
			return itAttributeGroup.SearchAttributeGroupsQuery(request)
		},
		func(result itAttributeGroup.SearchAttributeGroupsResult) SearchAttributeGroupsResponse {
			response := SearchAttributeGroupsResponse{}
			response.FromResult(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}
