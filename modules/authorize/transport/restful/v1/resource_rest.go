package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"

	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/resource"
)

type resourceRestParams struct {
	dig.In

	ResourceSvc it.ResourceService
}

func NewResourceRest(params resourceRestParams) *ResourceRest {
	return &ResourceRest{
		ResourceSvc: params.ResourceSvc,
	}
}

type ResourceRest struct {
	ResourceSvc it.ResourceService
}

func (this ResourceRest) CreateResource(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST create resource"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.ResourceSvc.CreateResource,
		func(request CreateResourceRequest) it.CreateResourceCommand {
			return it.CreateResourceCommand(request)
		},
		func(result it.CreateResourceResult) CreateResourceResponse {
			response := CreateResourceResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonCreated,
	)

	return err
}

func (this ResourceRest) UpdateResource(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST update resource"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.ResourceSvc.UpdateResource,
		func(request UpdateResourceRequest) it.UpdateResourceCommand {
			return it.UpdateResourceCommand(request)
		},
		func(result it.UpdateResourceResult) UpdateResourceResponse {
			response := UpdateResourceResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)

	return err
}

func (this ResourceRest) DeleteHardResource(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST delete hard resource"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.ResourceSvc.DeleteHardResource,
		func(request DeleteHardResourceRequest) it.DeleteHardResourceCommand {
			return it.DeleteHardResourceCommand(request)
		},
		func(result it.DeleteHardResourceResult) DeleteHardResourceResponse {
			response := DeleteHardResourceResponse{}
			response.FromNonEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)

	return err
}

func (this ResourceRest) GetResourceByName(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST get resource by name"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.ResourceSvc.GetResourceByName,
		func(request GetResourceByNameRequest) it.GetResourceByNameQuery {
			return it.GetResourceByNameQuery(request)
		},
		func(result it.GetResourceByNameResult) GetResourceByNameResponse {
			response := GetResourceByNameResponse{}
			response.FromResource(*result.Data)
			return response
		},
		httpserver.JsonOk,
	)

	return err
}

func (this ResourceRest) SearchResources(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST search resources"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.ResourceSvc.SearchResources,
		func(request SearchResourcesRequest) it.SearchResourcesQuery {
			return it.SearchResourcesQuery(request)
		},
		func(result it.SearchResourcesResult) SearchResourcesResponse {
			response := SearchResourcesResponse{}
			response.FromResult(result.Data)
			return response
		},
		httpserver.JsonOk,
	)

	return err
}
