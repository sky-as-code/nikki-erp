package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/resource"
)

type resourceRestParams struct {
	dig.In

	ResourceSvc it.ResourceService
}

func NewResourceRest(params resourceRestParams) *ResourceRest {
	return &ResourceRest{ResourceSvc: params.ResourceSvc}
}

type ResourceRest struct {
	httpserver.RestBase
	ResourceSvc it.ResourceService
}

func (this ResourceRest) CreateResource(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create resource"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequestDynamic(
		echoCtx,
		this.ResourceSvc.CreateResource,
		func(requestFields dmodel.DynamicFields) it.CreateResourceCommand {
			cmd := it.CreateResourceCommand{}
			cmd.SetFieldData(requestFields)
			return cmd
		},
		func(data domain.Resource) CreateResourceResponse {
			response := httpserver.NewRestCreateResponseDyn(data.GetFieldData())
			return *response
		},
		httpserver.JsonCreated,
	)
}

func (this ResourceRest) DeleteResource(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST delete resource"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.ResourceSvc.DeleteResource,
		func(request DeleteResourceRequest) it.DeleteResourceCommand {
			return it.DeleteResourceCommand(request)
		},
		func(data dyn.MutateResultData) DeleteResourceResponse {
			return httpserver.NewRestDeleteResponse2(data)
		},
		httpserver.JsonOk,
	)
}

func (this ResourceRest) GetResource(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get resource"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.ResourceSvc.GetResource,
		func(request GetResourceRequest) it.GetResourceQuery {
			return it.GetResourceQuery(request)
		},
		func(data domain.Resource) GetResourceResponse {
			return data.GetFieldData()
		},
		httpserver.JsonOk,
	)
}

func (this ResourceRest) ResourceExists(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST resource exists"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.ResourceSvc.ResourceExists,
		func(request ResourceExistsRequest) it.ResourceExistsQuery {
			return it.ResourceExistsQuery(request)
		},
		func(data dyn.ExistsResultData) ResourceExistsResponse {
			return ResourceExistsResponse(data)
		},
		httpserver.JsonOk,
	)
}

func (this ResourceRest) SearchResources(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST search resources"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.ResourceSvc.SearchResources,
		func(request SearchResourcesRequest) it.SearchResourcesQuery {
			return it.SearchResourcesQuery(request)
		},
		func(data it.SearchResourcesResultData) SearchResourcesResponse {
			return httpserver.NewSearchUsersResponseDyn(data)
		},
		httpserver.JsonOk,
		true,
	)
}

func (this ResourceRest) UpdateResource(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST update resource"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequestDynamic(
		echoCtx,
		this.ResourceSvc.UpdateResource,
		func(requestFields dmodel.DynamicFields) it.UpdateResourceCommand {
			cmd := it.UpdateResourceCommand{}
			cmd.SetFieldData(requestFields)
			return cmd
		},
		func(data dyn.MutateResultData) UpdateResourceResponse {
			return httpserver.NewRestUpdateResponse2(data)
		},
		httpserver.JsonOk,
	)
}
