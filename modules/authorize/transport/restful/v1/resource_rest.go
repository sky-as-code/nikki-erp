package v1

import (
	"github.com/labstack/echo/v4"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/resource"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"go.uber.org/dig"
)

type resourceRestParams struct {
	dig.In

	Config  config.ConfigService
	Logger  logging.LoggerService
	CqrsBus cqrs.CqrsBus
}

func NewResourceRest(params resourceRestParams) *ResourceRest {
	return &ResourceRest{
		RestBase: httpserver.RestBase{
			ConfigSvc: params.Config,
			Logger:    params.Logger,
			CqrsBus:   params.CqrsBus,
		},
	}
}

type ResourceRest struct {
	httpserver.RestBase
}

func (this ResourceRest) CreateResource(echoCtx echo.Context) (err error) {
	request := &CreateResourceRequest{}
	if err = echoCtx.Bind(request); err != nil {
		return err
	}

	result := it.CreateResourceResult{}
	err = this.CqrsBus.Request(echoCtx.Request().Context(), *request, &result)

	if err != nil {
		return err
	}

	if result.ClientError != nil {
		return httpserver.JsonBadRequest(echoCtx, result.ClientError)
	}

	response := CreateResourceResponse{}
	response.FromResource(*result.Data)

	return httpserver.JsonCreated(echoCtx, response)
}

func (this ResourceRest) UpdateResource(echoCtx echo.Context) (err error) {
	request := &UpdateResourceRequest{}
	if err = echoCtx.Bind(request); err != nil {
		return err
	}

	result := it.UpdateResourceResult{}
	err = this.CqrsBus.Request(echoCtx.Request().Context(), *request, &result)

	if err != nil {
		return err
	}

	if result.ClientError != nil {
		return httpserver.JsonBadRequest(echoCtx, result.ClientError)
	}

	response := UpdateResourceResponse{}
	response.FromResource(*result.Data)

	return httpserver.JsonOk(echoCtx, response)
}

func (this ResourceRest) GetResourceByName(echoCtx echo.Context) (err error) {
	request := &GetResourceByNameRequest{}
	if err = echoCtx.Bind(request); err != nil {
		return err
	}

	result := it.GetResourceByNameResult{}
	err = this.CqrsBus.Request(echoCtx.Request().Context(), *request, &result)

	if err != nil {
		return err
	}

	if result.ClientError != nil {
		return httpserver.JsonBadRequest(echoCtx, result.ClientError)
	}

	response := GetResourceByNameResponse{}
	response.FromResource(*result.Data)

	return httpserver.JsonOk(echoCtx, response)
}

func (this ResourceRest) SearchResources(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to list resources"); e != nil {
			err = e
		}
	}()

	request := &SearchResourcesRequest{}
	if err = echoCtx.Bind(request); err != nil {
		return err
	}

	result := it.SearchResourcesResult{}
	err = this.CqrsBus.Request(echoCtx.Request().Context(), *request, &result)

	if err != nil {
		return err
	}

	if result.ClientError != nil {
		return httpserver.JsonBadRequest(echoCtx, result.ClientError)
	}

	response := SearchResourcesResponse{}
	response.FromResult(result.Data)

	return httpserver.JsonOk(echoCtx, response)
}
